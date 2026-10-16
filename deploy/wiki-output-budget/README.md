# Wiki 输出额度 12288

只将 Wiki generateWithTemplate 的 max_tokens 从 8192 提高到 12288；不修改模型上下文、分块、提示词、思考模式、图谱并发及重试策略。现有 Wiki 单次调用请求关闭思考。模型实际上下文仍为 51200。

## 1. 测试与构建

以下均在服务器 ubuntu 会话执行。依赖复用现有持久化缓存；构建镜像使用全局 GOPROXY=https://goproxy.cn，不安装临时依赖。

```bash
cd /home/ubuntu/RAG_Weknora
git status --short
docker run --rm --name weknora-wiki-budget-build -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:enrichment-retry-20261016 sh -c 'go test ./internal/application/service -run "TestGenerateWithTemplate|TestEnrichmentRecovery" -count=1 && go build -buildvcs=false -o runtime/WeKnora.wiki-output-12288-20261016 ./cmd/server'
git diff --check
mkdir -p runtime/wiki-output-image-context
cp runtime/WeKnora.wiki-output-12288-20261016 runtime/wiki-output-image-context/WeKnora
docker build -f deploy/wiki-output-budget/Dockerfile.runtime -t weknora-app:wiki-output-12288-20261016 runtime/wiki-output-image-context
```

镜像继承当前 archive-mapping 部署，仅替换应用二进制，保留迁移、系统依赖和此前功能。构建中断后先 docker inspect weknora-wiki-budget-build，检查 runtime/wiki-output-12288-build.log；已有构建仍在运行时不重复启动。

## 2. 模型接口验证

不写生产任务，只发送一个独立短请求确认 12288 输出额度能被当前模型接受。它不是完整长输出无截断的保证。

```bash
curl --fail http://127.0.0.1:8102/v1/chat/completions -H 'Content-Type: application/json' -d '{"model":"qwen3-4b-instruct-2507","messages":[{"role":"user","content":"Reply with only the JSON object {\"ok\":true}."}],"max_tokens":12288,"temperature":0,"chat_template_kwargs":{"enable_thinking":false}}'
```

## 3. 上线

仅重建 app，模型容器不重启。应用有短暂不可用，活跃图谱请求可能因退出取消并重新调度；任务保存在 Redis/PostgreSQL，不重新解析文档，不手动消耗冷却中的死信机会。先备份配置和当前镜像，再确认两个 compose 文件中的默认 image 均为 weknora-app:wiki-output-12288-20261016。

```bash
cd /home/ubuntu/RAG_Weknora
mkdir -p backups/wiki-output-12288-20261016
cp deploy/pdf-unlimited/compose.yaml backups/wiki-output-12288-20261016/pdf-compose.yaml
cp deploy/enrichment-retry/compose.yaml backups/wiki-output-12288-20261016/enrichment-compose.yaml
docker inspect WeKnora-app --format '{{.Config.Image}} {{.Image}}'
# 首次上线时旧镜像为 weknora-app:archive-mapping-20261016；保留，不删除。
# 把两个持久化默认 image 更新到新 tag（重复执行不重复修改）：
python3 - <<'CONFIG_PY'
from pathlib import Path
for name in ['deploy/pdf-unlimited/compose.yaml', 'deploy/enrichment-retry/compose.yaml']:
    p = Path(name)
    p.write_text(p.read_text().replace('weknora-app:archive-mapping-20261016', 'weknora-app:wiki-output-12288-20261016'))
CONFIG_PY
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml stop -t 120 app
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
```

## 4. 上线核验

```bash
cd /home/ubuntu/RAG_Weknora
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
curl --fail http://127.0.0.1:18081/health
docker exec WeKnora-app sha256sum /app/WeKnora
sha256sum runtime/WeKnora.wiki-output-12288-20261016
docker logs --since 5m WeKnora-app 2>&1 | grep -E 'graph-pool|EnrichmentRecovery|ERROR|panic'
docker exec WeKnora-postgres psql -X -U weknora -d weknora -c "SELECT state,attempts,count(*),min(available_at) FROM enrichment_recoveries WHERE knowledge_base_id='8575ca20-5e7e-48c4-9fc0-786d979fdf5d' GROUP BY state,attempts;"
```

还需读取目标库与图谱信号量实际运行状态，确认仍为未暂停、8 个活跃图谱请求，冷却记录没有提前消耗。实际 HTTP 端口以当前 docker port WeKnora-app 为准。两条摘要/抽取失败任务按原定冷却时间自动恢复；本次不强制提前重试。12288 仍是有限预算，更长的 JSON 仍可能截断。

## 5. 回退

先检查 git status 与当前容器，保留新旧镜像。仅回退应用镜像不回滚数据库数据，也不重启模型。持久回退需把两个 compose 默认 image 恢复到旧 tag；代码用 git revert 本次提交保留历史，不 reset 工作区。

```bash
cd /home/ubuntu/RAG_Weknora
ENRICHMENT_APP_IMAGE=weknora-app:archive-mapping-20261016 docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
```
