# WeKnora 18182 生产升级

服务器为 `ubuntu@10.20.0.15`，权威仓库为 `/home/ubuntu/RAG_Weknora`。本次升级使用独立发布工作区 `/home/ubuntu/RAG_Weknora-release-20261105`。服务配置、备份、发布脚本及状态保存在服务器目录 `/home/ubuntu/weknora-fork-migration/20261112-production`，其中包含私有信息，保留在服务器。

| 组件 | 源码或版本 | 镜像或入口 |
|---|---|---|
| 上游主线 | `e8e0d2ed4bd660c46313884d2f2b3b243d7c318d` | 已纳入后端源码 |
| 后端 | `61c41d0eece3a0bc9334ed5e829b67328bb783f2` | `weknora-app:ljc-production-61c41d0ee` |
| DocReader | `fda761a6bbdbd7158b7fe2544ee4567b86e44910` | `weknora-docreader:ljc-sync-fda761a6b` |
| 定制前端 | `fda761a6bbdbd7158b7fe2544ee4567b86e44910` | `weknora-ui-custom:ljc-sync-fda761a6b`，端口 `18182` |
| PostgreSQL | 官方 `134`＋个人 `5` | `weknora_release_20261112`，原 `weknora` 保留 |

2026-11-12 16:07（北京时间）切换完成。第一次切换时，PostgreSQL 在按模板复制原库时发生进程异常，旧后端恢复。第二次在停写后将数据直接流式恢复到升级库。12 张业务表在原库、恢复后和迁移后的内容指纹一致，11 个检索索引有效。服务端健康检查、18180 和 18182 代理、18182 登录页、匿名 `policy-qa` 实际生成及 1 条引用明细均已验证。登录后的页面操作本次未实测。

## 1. 检查运行状态和恢复点

在 Linux 服务器运行。先读取状态及退出码。后台切换中断时，先确认任务和容器实际状态，再决定如何恢复。发布脚本持有 `release.lock`，切换过程中不要再启动一次。

```bash
release_dir=/home/ubuntu/weknora-fork-migration/20261112-production
cd "$release_dir"
python3 release.py status
cat backend-final.status build-release.status cutover.status
cat rehearsal-results.json preflight-database.sha256
python3 verify-live.py
docker ps --format '{{.Names}} {{.Image}} {{.Ports}}' | grep -E 'WeKnora-(app|frontend|docreader|postgres)'
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18182/config.js
curl -fsS -o /dev/null http://127.0.0.1:18180/
```

查看升级前后两个数据库的迁移账本。此命令只读。账本应分别为 `128/5` 和 `134/5`，四项 `dirty` 均为 `f`。

```bash
for release_db in weknora weknora_release_20261112; do
  printf '%s\n' "$release_db"
  docker exec WeKnora-postgres-20261021 psql -XAt -v ON_ERROR_STOP=1 \
    -U weknora -d "$release_db" \
    -c 'SELECT version, dirty FROM schema_migrations;' \
    -c 'SELECT version, dirty FROM ljc_schema_migrations;'
done
```

## 2. 重跑后端测试及构建

以下命令使用已持久化的 Go 构建镜像、国内镜像源配置与缓存。在服务端发布工作区运行，构建产物写入发布目录内的独立子目录。

```bash
release_dir=/home/ubuntu/weknora-fork-migration/20261112-production
cd /home/ubuntu/RAG_Weknora-release-20261105
git diff --exit-code 61c41d0ee -- ':!docs/production-release-20261112'
bash "$release_dir/backend-tests.sh"
mkdir -p "$release_dir/rebuild"
cp -a migrations/ljc "$release_dir/rebuild/"
cp "$release_dir/build-context/Dockerfile" "$release_dir/rebuild/"
docker run --pull=never --rm -e COMMIT_ID=61c41d0eece3a0bc9334ed5e829b67328bb783f2 \
  -e VERSION=0.8.3 -e BUILD_TIME="$(date -u +%FT%TZ)" -e GOFLAGS=-buildvcs=false \
  -e GOMAXPROCS=10 -v "$PWD:/src" -v "$release_dir/rebuild:/output" \
  -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build -w /src \
  weknora-builder:release-20261112-tests sh -ec '
    export GO_VERSION=$(go version | cut -d" " -f3)
    make build-prod GO_BUILD_TAGS=anydoc BINARY_NAME=/output/WeKnora
    go build -o /output/ljc-migrate ./cmd/ljc-migrate
  '
docker build --pull=false --build-arg SOURCE_REVISION=61c41d0eece3a0bc9334ed5e829b67328bb783f2 \
  -t weknora-app:ljc-rebuild-61c41d0ee "$release_dir/rebuild"
docker run --pull=never --rm --network none --entrypoint /app/ljc-migrate \
  weknora-app:ljc-rebuild-61c41d0ee -h
```

DocReader 和前端复用已验收镜像。用下面的命令确认源码相同，并核对服务实际镜像。

```bash
git -C /home/ubuntu/RAG_Weknora-release-20261105 diff --exit-code \
  fda761a6b 61c41d0ee -- frontend frontend-custom docreader third_party/anydoc-go
docker inspect --format '{{.Name}} {{.Image}}' \
  WeKnora-app WeKnora-docreader WeKnora-frontend-custom-preview
```

## 3. 验证 18182

在获准使用匿名 API 的内网客户端运行。测试会调用真实模型，并创建一条测试响应记录。随后打开 18182，用已有账号检查登录后的知识库和问答。

```powershell
Set-Location E:/desktop/文档RAG/reports/weknora-production/2026-11-12
python verify-production.py --output client-recheck.json
Start-Process 'http://10.20.0.15:18182/'
```

## 4. 恢复旧版本

旧数据库 `weknora` 和旧镜像均保留。若新版已接收业务写入，先暂停入口，备份新版数据库及相关新增文件，确认新增数据如何保留，再用下面的旧配置恢复。18081 后端同时供 18180 和 18182 使用，恢复期间两端的 API 会短暂中断。请在维护窗口执行。

```bash
(
set -euo pipefail
release_dir=/home/ubuntu/weknora-fork-migration/20261112-production
cd "$release_dir"
umask 077
docker stop -t 60 WeKnora-app
docker exec WeKnora-postgres-20261021 pg_dump -U weknora -Fc \
  -d weknora_release_20261112 > "database.before-rollback.$(date +%Y%m%d-%H%M%S).dump"
docker compose -p ljc-weknora-production -f compose.before.json up -d --no-deps --pull never docreader
docker compose -p ljc-weknora-production -f compose.before.json up -d --no-deps --pull never app
docker compose -p ljc-weknora-production -f compose.before.json up -d --no-deps --pull never frontend
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18182/config.js
)
```

本次切换沿用已完成的 `preflight-database.dump`，按你的要求停止新增文件归档和切换前第二份数据库转储。原库 `weknora` 是切换时点的优先回退来源。若原库不可用，可用以下命令把较早的快照恢复到独立库。恢复前核对快照时间与业务数据差异。

```bash
cd /home/ubuntu/weknora-fork-migration/20261112-production
sha256sum -c preflight-database.sha256
docker exec -i WeKnora-postgres-20261021 pg_restore -l < preflight-database.dump > /dev/null
docker exec WeKnora-postgres-20261021 createdb -U weknora weknora_restore_20261112
docker exec -i WeKnora-postgres-20261021 pg_restore --exit-on-error \
  -U weknora -d weknora_restore_20261112 < preflight-database.dump
```
