# 图谱输出额度提高到 12288 tokens

仅把后台图谱 NewGraphExtractor 的输出上限从 8192 改成 12288；Wiki 维持 12288，查询时实体提取维持 4096。保持 temperature=0.3、关闭思考、图谱独立并发 8、9B max-num-seqs=12 与现有重试策略。模型上下文不变；9B 当前为 49152，4B 为 51200。提高上限不能保证 JSON 语法正确。

## 1. 测试和构建

在服务器 ubuntu 会话执行。复用持久模块/构建缓存；构建镜像全局使用 GOPROXY=https://goproxy.cn。

```bash
cd /home/ubuntu/RAG_Weknora
git status --short
docker run --rm --name weknora-graph-budget-build -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:enrichment-retry-20261016 sh -c 'go test ./internal/application/service/chat_pipeline -run "TestExtractionCompletionBudgets|TestGraphExtractorRejectsInvalidJSON|TestFormater" -count=1 && go test ./internal/application/service -run "TestGenerateWithTemplate|TestEnrichmentRecovery" -count=1 && go build -buildvcs=false -o runtime/WeKnora.graph-output-12288-20261016 ./cmd/server'
git diff --check
mkdir -p runtime/graph-output-12288-image-context
cp runtime/WeKnora.graph-output-12288-20261016 runtime/graph-output-12288-image-context/WeKnora
docker build -f deploy/graph-output-12288/Dockerfile.runtime -t weknora-app:graph-output-12288-20261016 runtime/graph-output-12288-image-context
```

当前基础镜像为 external-chunks-20261016，以保留已上线的外部分块导入、档案映射、Wiki 额度和队列修改。中断后先 docker inspect weknora-graph-budget-build 并查看 runtime/graph-output-12288-build.log，仍在运行则不重复构建，已退出失败则按原命令重试。

## 2. 实际模型请求验证

请求是独立探针，不处理生产文档，不消耗死信机会。只验证当前服务接受新额度；不声称一个短输出证明所有长输出都不会截断。

```bash
curl --fail http://127.0.0.1:8101/v1/chat/completions -H 'Content-Type: application/json' -d '{"model":"qwen35-9b","messages":[{"role":"user","content":"Return only a JSON array containing one object with key entity and value test."}],"max_tokens":12288,"temperature":0.3,"chat_template_kwargs":{"enable_thinking":false}}'
```

## 3. 持久化配置及部署

先备份两个 compose 配置，再将其默认 app 镜像都改成新 tag。只重启应用，模型容器不重启。知识库当前主动暂停，保持暂停；不重解析、不提前触发冷却恢复任务。

```bash
cd /home/ubuntu/RAG_Weknora
mkdir -p backups/graph-output-12288-20261016
cp -n deploy/pdf-unlimited/compose.yaml backups/graph-output-12288-20261016/pdf-compose.yaml
cp -n deploy/enrichment-retry/compose.yaml backups/graph-output-12288-20261016/enrichment-compose.yaml
python3 - <<'CONFIG_PY'
from pathlib import Path
for name in ['deploy/pdf-unlimited/compose.yaml','deploy/enrichment-retry/compose.yaml']:
    p=Path(name)
    s=p.read_text()
    for old in ['weknora-app:external-chunks-20261016','weknora-app:wiki-output-12288-20261016']:
        s=s.replace(old,'weknora-app:graph-output-12288-20261016')
    p.write_text(s)
CONFIG_PY
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml stop -t 120 app
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
```

## 4. 核验

```bash
cd /home/ubuntu/RAG_Weknora
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
curl --fail http://127.0.0.1:18081/health
sha256sum runtime/WeKnora.graph-output-12288-20261016
docker exec WeKnora-app sha256sum /app/WeKnora
docker logs --since 5m WeKnora-app 2>&1 | grep -E 'graph-pool|EnrichmentRecovery|panic'
docker exec WeKnora-postgres psql -X -U weknora -d weknora -c "SELECT state,attempts,count(*),min(available_at) FROM enrichment_recoveries WHERE knowledge_base_id='8575ca20-5e7e-48c4-9fc0-786d979fdf5d' GROUP BY state,attempts;"
```

二进制 SHA 应一致。当前暂停时模型活跃数为 0 属正常；用户恢复后，再查看新图谱 span.input.max_output_tokens 是否为 12288、图谱实际模型并发是否为 8。旧 span 保存的是历史 8192，不代表新代码未生效。采集脚本见本地 runtime/weknora-hourly-monitor/README.md。回查时区分 JSON 语法错误、输出截断、HTTP 超时和推理引擎错误。

## 5. 回退

保留前后镜像，回退应用不还原数据库、不重启模型。持久回退从备份恢复两个 compose 配置（先查看 diff，避免覆盖后续其他更改）。源码使用 git revert 本次提交，不使用 git reset --hard。

```bash
cd /home/ubuntu/RAG_Weknora
ENRICHMENT_APP_IMAGE=weknora-app:external-chunks-20261016 docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
```
