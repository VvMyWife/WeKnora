# 图谱独立并发与延后重试

服务器代码：`/home/ubuntu/RAG_Weknora`，当前私有分支继续使用，不推送远程。

图谱独立 worker 池为 8，Redis 按 `graph:<model_id>` 在应用实例间共享 8 个名额；不会叠加普通模型的默认限额 2。图谱限流存储异常时拒绝新请求，租约丢失会取消正在执行的模型请求。摘要/Wiki 等保留原限额。9B 服务的 `max-num-seqs=12` 是整个服务的容量，并不为其他应用硬预留名额。

普通 Graph/Wiki 失败至少冷却 30 分钟，到期进入队尾。Graph 保留 3 次普通重试，Wiki 文档保留 5 次；Wiki 模型调用不在单次调用内部连续重试。暂停、取消、锁争用不消耗相应的失败机会。

普通机会耗尽后，`enrichment_recoveries` 持久化保存恢复队列，每个逻辑任务额外尝试两次，每次至少间隔两小时。恢复不重置普通预算。后台每分钟自检，异常后下轮自动继续；派发轮次防止旧回调误结束新一轮，Wiki 领取中断保留原操作并通过租约回收继续处理。删除、取消、新轮次替代和已成功任务跳过。首次仅补入政策知识库 `8575ca20-5e7e-48c4-9fc0-786d979fdf5d` 的有效历史死信；其他知识库只处理上线后新增死信。

## 1. 自己运行自动化测试

以下命令在服务器普通 ubuntu 会话执行。现有 Docker 镜像和 Go 模块缓存持久化复用，不需要安装新依赖。专用构建镜像把国内源设置为构建期全局 `GOPROXY=https://goproxy.cn`。

```bash
cd /home/ubuntu/RAG_Weknora
docker build -f deploy/enrichment-retry/Dockerfile.builder -t weknora-builder:enrichment-retry-20261016 .
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:enrichment-retry-20261016 go test ./internal/models/limiter ./internal/models/chat ./internal/types ./internal/router ./internal/handler ./internal/middleware/asynqdl ./internal/application/repository ./internal/application/service
git diff --check
```

真实 PostgreSQL/Redis 集成测试使用独立容器，没有宿主机映射端口。测试会重置专用 `enrichment_test` 数据库及专用测试 Redis，绝不能替换为生产地址。

```bash
docker network inspect weknora-retry-test-net >/dev/null 2>&1 || docker network create weknora-retry-test-net
docker inspect weknora-retry-test-pg >/dev/null 2>&1 || docker run -d --name weknora-retry-test-pg --network weknora-retry-test-net --network-alias pg -e POSTGRES_USER=tester -e POSTGRES_PASSWORD=enrichment-test-only -e POSTGRES_DB=enrichment_test paradedb/paradedb:v0.22.2-pg17
docker inspect weknora-retry-test-redis >/dev/null 2>&1 || docker run -d --name weknora-retry-test-redis --network weknora-retry-test-net --network-alias retry-redis redis:7.0-alpine
docker start weknora-retry-test-pg weknora-retry-test-redis
docker exec weknora-retry-test-pg pg_isready -U tester -d enrichment_test
docker run --rm --network weknora-retry-test-net -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src -e WEKNORA_ENRICHMENT_TEST_DSN='host=pg user=tester password=enrichment-test-only dbname=enrichment_test sslmode=disable' -e WEKNORA_ENRICHMENT_TEST_REDIS=retry-redis:6379 weknora-builder:enrichment-retry-20261016 go test -p 1 -v ./internal/models/limiter ./internal/application/service -run 'TestGraphRealRedis|TestEnrichmentRecoveryReal'
```

验收覆盖跨实例 8 个名额、第 9 个等待、普通默认 2 不生效、限流故障拒绝放行、冷却与同组去重、两次额外恢复、重复投递、旧回调、暂停及新解析覆盖时的计数。

## 2. 构建

首次上线时三个原有未提交的 archive 修改只保留在工作区，构建通过只读覆盖使用其 HEAD 版本，未随本次部署。下面创建相同的构建视图；不要覆盖或 reset 原文件。

```bash
cd /home/ubuntu/RAG_Weknora
mkdir -p runtime/enrichment-build-baseline/internal/archivemapping runtime/enrichment-build-baseline/internal/handler
git show HEAD:internal/archivemapping/store.go > runtime/enrichment-build-baseline/internal/archivemapping/store.go
git show HEAD:internal/handler/knowledge_archive.go > runtime/enrichment-build-baseline/internal/handler/knowledge_archive.go
git show HEAD:internal/handler/knowledge_archive_test.go > runtime/enrichment-build-baseline/internal/handler/knowledge_archive_test.go
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -v "$PWD/runtime/enrichment-build-baseline/internal/archivemapping/store.go:/src/internal/archivemapping/store.go:ro" -v "$PWD/runtime/enrichment-build-baseline/internal/handler/knowledge_archive.go:/src/internal/handler/knowledge_archive.go:ro" -v "$PWD/runtime/enrichment-build-baseline/internal/handler/knowledge_archive_test.go:/src/internal/handler/knowledge_archive_test.go:ro" -w /src weknora-builder:enrichment-retry-20261016 sh -c 'go test ./internal/handler ./internal/application/service ./internal/application/repository && go build -buildvcs=false -o runtime/WeKnora.enrichment-retry-20261016 ./cmd/server && go build -buildvcs=false -o runtime/enrichment-acceptance ./cmd/enrichment-acceptance'
```

## 3. 备份与上线

正式知识库保持暂停。应用和 9B 容器重建会有短暂不可用，其他 9B 调用方也会受到影响。先备份、完成构建和测试，再进入切换。

```bash
cd /home/ubuntu/RAG_Weknora
mkdir -p backups/enrichment-20261016 runtime/enrichment-image-context/migrations
docker image inspect weknora-app:pre-enrichment-20261016 >/dev/null 2>&1 || docker image tag "$(docker inspect WeKnora-app --format '{{.Image}}')" weknora-app:pre-enrichment-20261016
test -s runtime/WeKnora.pre-enrichment-20261016 || docker cp WeKnora-app:/app/WeKnora runtime/WeKnora.pre-enrichment-20261016
test -s backups/enrichment-20261016/processing-before.dump || docker exec WeKnora-postgres pg_dump -U weknora -d weknora -Fc -t knowledges -t task_pending_ops -t task_dead_letters > backups/enrichment-20261016/processing-before.dump
cp runtime/WeKnora.enrichment-retry-20261016 runtime/enrichment-image-context/WeKnora
cp migrations/versioned/000091_archive_mapping.*.sql migrations/versioned/000092_enrichment_retry.*.sql runtime/enrichment-image-context/migrations/
docker build -f deploy/enrichment-retry/Dockerfile.runtime -t weknora-app:enrichment-retry-20261016 runtime/enrichment-image-context
python3 deploy/enrichment-retry/normalize_graph_retries.py
python3 deploy/enrichment-retry/reconcile_paused_counters.py
```

9B 的持久化配置需要先从 ubuntu 登录后切换 root；只修改指定服务的 `max-num-seqs`，脚本自动保留原文件备份。

```bash
su - root
python3 /home/ubuntu/RAG_Weknora/deploy/enrichment-retry/adjust_vllm.py
docker compose -f /root/kaifang/vllm/docker-compose-scu.yml config --services
exit
```

切换时先停应用，再重建 9B；确认模型接口就绪后才启动新应用，避免初始化期间业务失败。数据库迁移是增量、可重复执行的，不删除旧表或旧记录。

```bash
cd /home/ubuntu/RAG_Weknora
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml stop app
docker exec -i WeKnora-postgres psql -X -v ON_ERROR_STOP=1 -U weknora -d weknora < migrations/versioned/000092_enrichment_retry.up.sql
su - root
docker compose -f /root/kaifang/vllm/docker-compose-scu.yml up -d --no-deps scu-qwen35-9b
exit
curl --fail http://127.0.0.1:8101/v1/models
# 上一条成功，且模型启动日志出现 Application startup complete 后继续。
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
python3 deploy/enrichment-retry/normalize_graph_retries.py --apply
python3 deploy/enrichment-retry/reconcile_paused_counters.py --apply
```

计数核对脚本只处理目标知识库：正文解析、向量化、摘要已完成，仍有当前轮次图谱排队任务，且仅被 housekeeping 标超时的文档。执行前保存 JSON 清单；不会重新解析正文。新 housekeeping 不再把主动暂停当作超时。

## 4. 观察与实际模型验收

```bash
docker inspect scu-qwen35-9b --format '{{json .Args}}'
docker logs --since 10m WeKnora-app 2>&1 | grep -E 'graph-pool|EnrichmentRecovery|ModelLimiter'
docker exec WeKnora-postgres psql -U weknora -d weknora -P pager=off -c "SELECT state,attempts,count(*),min(available_at) FROM enrichment_recoveries GROUP BY state,attempts ORDER BY state,attempts;"
docker exec WeKnora-postgres psql -U weknora -d weknora -P pager=off -c "SELECT knowledge_id,state,attempts,dispatch_no,available_at,lease_until,last_error FROM enrichment_recoveries ORDER BY id DESC LIMIT 30;"
```

管理员运行队列接口 `/api/v1/system/admin/runtime/queues` 新增 graph 池、`graph:<model_id>` 模型限流记录及 `enrichment_retry` 状态。等待恢复、成功恢复和最终耗尽可分别查看，不依赖“当前失败文档数”。

`runtime/enrichment-acceptance` 发起 9 个受图谱限流约束的实际模型请求，先确认 8 个持有名额、第 9 个等待，再执行请求；不创建知识文档。它从环境读取 `REDIS_ADDR`、`REDIS_PASSWORD`，输出只包含并发计数。配套脚本会自动读取现有应用 Redis 配置并核对 Docker 网络，不输出密码，同时采样 vLLM 实际运行请求数：

```bash
python3 deploy/enrichment-retry/run_live_acceptance.py
```

通过条件为 `eight-admitted-ninth-waiting`、`global_default=2`、图谱 `limit=8/active=8`，以及 `completed=9 failed=0 provider_http_peak=8`。不能只用 `/health` 或配置值判断验收成功。

## 5. 回退与中断恢复

不要删除原镜像、备份和迁移表。切换中断时先检查当前镜像、模型就绪状态和知识库暂停标志；构建可原命令重跑，脚本重复运行不会重新发放死信预算。

若需回退，先保持相关知识库暂停；若已经存在 queued 恢复任务，先核对并停放相关知识库，不要让旧代码处理新恢复操作。使用保留的旧镜像回退应用，不回滚用户数据：

```bash
cd /home/ubuntu/RAG_Weknora
ENRICHMENT_APP_IMAGE=weknora-app:pre-enrichment-20261016 docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
```

9B 如需恢复旧容量，从 `docker-compose-scu.yml.before-enrichment-*` 中选择本次保存的确切备份，核对 diff 后恢复并仅重建 `scu-qwen35-9b`。应用回退本身不要求降低模型容量。不要自动执行数据库 down migration 或整库还原。

## 本次上线验收记录

2026-10-16 10:57（北京时间）：运行镜像为 `weknora-app:enrichment-retry-20261016`，健康状态正常；运行二进制与构建产物 SHA-256 一致，数据库版本 92。9B 容器实际参数为 `max-num-seqs=12`。

真实模型验收输出：

```text
global_default=2; graph limit=8; active=8; waiting=1
completed=9 failed=0 provider_http_peak=8
vllm_running_peak=8
```

正式知识库保持暂停，634 份文档中 completed=91、finalizing=543；28,306 个未结束计数与对应 Redis 图谱任务逐份完全一致。已依据备份清单恢复 9 份仅被旧 housekeeping 误判的文档、共 10 个计数，未重新解析正文。25 条有效 Wiki 死信为 ready、额外尝试数 0，首次最早在北京时间 12:52 后且知识库恢复处理时派发。

完整证据位于 `runtime/enrichment-delivery-evidence.json`、`runtime/enrichment-live-acceptance.log`、`runtime/enrichment-integration-final.log`、`runtime/enrichment-release-tests.log`。原有三处 archive 未提交改动保留。
