# WeKnora 生产目录和分支流程

权威源码目录为 `/home/ubuntu/RAG_Weknora`，生产级维护分支为 `ljc/custom`。服务器在发布时选择已验收镜像。源码分支切换和构建使用 Git 工作区，线上运行配置使用忽略目录 `runtime/production/`，日志继续写到 `data/processing-logs/`。

## 查看生产状态

```bash
cd /home/ubuntu/RAG_Weknora
git status --short --branch
git rev-parse HEAD
git worktree list
cat runtime/production/consolidation-state.json
docker inspect WeKnora-app --format '{{.State.Health.Status}} {{index .Config.Labels "org.opencontainers.image.revision"}} {{.Image}}'
docker inspect WeKnora-app --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18180/
curl -fsS -o /dev/null http://127.0.0.1:18182/
```

本次已验收的发布记录为 `168ec4ddcce61c430a6bb9e617c0f521a22c96f2`，实际后端镜像源码为 `61c41d0eece3a0bc9334ed5e829b67328bb783f2`，数据库为 `weknora_release_20261112`（官方 134＋个人 5）。主目录收拢复用了已部署镜像和数据库。仅重建了后端。18180、18182 代理、18182 匿名真实问答和 1 条引用明细通过，容器内配置和预置技能文件可读，处理日志挂载通过写读检查。GPU 服务 `18083` 返回 `cuda:0`，实际重排通过。

## 下一轮官方同步与验收

在主目录内执行。把经审阅的官方完整 SHA 输入命令，分支名中的主题可按实际工作修改。

```bash
cd /home/ubuntu/RAG_Weknora
git switch ljc/custom
test -z "$(git status --porcelain)" || exit 1
git fetch upstream
git switch -c "ljc/integrate-$(date +%Y%m%d)-upstream"
read -r -p '已审阅的官方完整 SHA: ' official_sha
git cat-file -e "$official_sha^{commit}" || exit 1
git merge-base --is-ancestor "$official_sha" upstream/main || exit 1
git -c user.name=ljc -c user.email=ljc@local merge --no-ff "$official_sha"
```

解决冲突后，分阶段测试并提交。沿用已持久化的国内源和缓存。在主目录运行后端测试，校验工作区和配置。

```bash
cd /home/ubuntu/RAG_Weknora
git diff --check
bash runtime/production/backend-tests.sh
docker compose -p ljc-weknora-production -f runtime/production/compose.json config --quiet
```

按改动范围补跑前端、DocReader、迁移和业务回归。以固定候选 SHA 构建镜像，验收环境单独配置端口、PostgreSQL、Redis、存储和模型。记录候选 SHA、镜像 ID 和结果，结束后停止候选服务。测试分支中途遇到维护分支更新时，先合入最新维护分支并补测，再执行下列收尾。

```bash
cd /home/ubuntu/RAG_Weknora
test -z "$(git status --porcelain)" || exit 1
candidate="$(git branch --show-current)"
case "$candidate" in ljc/integrate-*|ljc/feature-*) ;; *) exit 1 ;; esac
git merge-base --is-ancestor ljc/custom "$candidate" || exit 1
git switch ljc/custom
git merge --ff-only "$candidate" || exit 1
git merge-base --is-ancestor "$candidate" HEAD || exit 1
git status --short --branch
git branch -d "$candidate"
```

## 发布、验证与回退

正式发布前，在维护窗口核对活跃任务、镜像 SHA、数据库兼容性和可用回退路径。使用 `runtime/production/compose.json` 的私有生产参数更新已验收镜像。发布和回退分别核对镜像、配置和数据库，数据库发生变更时按该次发布记录恢复。

下面以本次挂载变更为例。回退清单保留了切换时的配置副本，使用当前后端镜像和当前数据库。切换或恢复前检查请求和队列，避免中断任务。

```bash
cd /home/ubuntu/RAG_Weknora
docker compose -p ljc-weknora-production -f runtime/production/compose.json config --quiet
docker compose -p ljc-weknora-production -f runtime/production/compose.json \
  up -d --no-deps --pull never app
curl -fsS http://127.0.0.1:18081/health
```

挂载回退时使用下列清单。之后复核 18180、18182 和实际问答。

```bash
cd /home/ubuntu/RAG_Weknora
docker compose -p ljc-weknora-production -f runtime/production/compose.rollback-mounts.json config --quiet
docker compose -p ljc-weknora-production -f runtime/production/compose.rollback-mounts.json \
  up -d --no-deps --pull never --force-recreate app
curl -fsS http://127.0.0.1:18081/health
```

从获准访问匿名 API 的内网客户端运行正式入口验收。该命令会生成一条测试响应。

```bash
python3 docs/production-release-20261112/verify-production.py --output runtime/production/client-recheck.json
```

操作中断时先查 `git status`、`git worktree list`、当前镜像、容器挂载和 `runtime/production/consolidation-state.json`，根据实际状态继续。若正在重建，先检查对应 PID、日志及退出码。合入后的版本和线上实际镜像分别记录。

## GPU 重排和旧工作树

生产 `weknora-bge-rerank.service` 从 `runtime/production/rerank/` 启动，沿用原 GPU、模型权重、Python 环境和 18083 端口。2026-11-12 从新路径实测 CUDA 重排，并通过 18182 问答和 3 条引用明细。原 Git 工作树已退役，验收资料归入 `runtime/retired-worktrees/`，两个停止的验收容器已移除，其数据卷保留。

运行以下只读检查，确认 GPU 工作目录和健康状态。

```bash
systemctl show weknora-bge-rerank.service -p WorkingDirectory -p MainPID -p ActiveState
curl -fsS http://10.20.0.15:18083/health
curl -fsS -H 'Content-Type: application/json' \
  --data '{"query":"数据备份","documents":["每日进行数据备份","今天气温较高"]}' \
  http://10.20.0.15:18083/rerank
```

在维护窗口重建重排启动目录时，先从已验收的维护分支准备运行文件。更新启动文件或 systemd 单元前记录正在使用的版本，以便回退到已验收的启动文件。随后重启并执行上述检查，健康响应应为 CUDA，第一条候选应排在首位。

```bash
cd /home/ubuntu/RAG_Weknora
rerank_dir="$PWD/runtime/production/rerank"
install -d "$rerank_dir/adapters/rerank"
install -m 0644 adapters/rerank/server.py "$rerank_dir/adapters/rerank/server.py"
install -m 0750 scripts/ljc_production_gpu_rerank.sh "$rerank_dir/start.sh"
bash -n "$rerank_dir/start.sh"
sudo install -m 0644 deploy/production-20261021/weknora-bge-rerank.service \
  /etc/systemd/system/weknora-bge-rerank.service
sudo systemctl daemon-reload
sudo systemctl restart weknora-bge-rerank.service
```
