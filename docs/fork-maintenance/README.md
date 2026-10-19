# WeKnora：ljc 定制维护说明

权威目录：/home/ubuntu/RAG_Weknora。公开 fork：https://github.com/VvMyWife/WeKnora。
官方上游：https://github.com/Tencent/WeKnora。最初来源为 Gitee 镜像，保留官方 v0.8.0 的祖先及现有47个定制提交。
现有 frontend 和 frontend-custom 冻结，不合入官方界面更新。日常观察不修改源码，由用户指定要引入的更新。

## 1. 分支、远端与生产关系

- main：官方完整代码镜像，仅用于观察；更新它不会自动部署。
- ljc/custom：正式定制维护分支。
- ljc/integrate-YYYYMMDD-topic：单次选定更新的集成分支。
- origin：自己的 fork；upstream：官方仓库；upstream-mirror：Gitee 备用镜像。
- 原 frontend-custom 分支和独立前端历史保留为私有恢复材料，不强推、不重写。
- Git 提交、构建产物与运行镜像是三种状态，必须分别记录。切分支不会自动更新容器。
- 首次迁移不增加新的官方功能；本次新增的是维护检查和迁移账本隔离。
- 生产原有前端镜像、数据卷和后台任务不因 fork 创建或推送发生切换。

进入服务器终端后：

~~~bash
cd /home/ubuntu/RAG_Weknora
git status --short --branch
git remote -v
git log -8 --oneline
python3 scripts/ljc_status.py
~~~

Windows 可使用本仓库的 Paramiko 工具；密码由隐藏提示输入，不保存：

~~~powershell
python scripts/ljc_remote.py "git -C /home/ubuntu/RAG_Weknora status --short --branch"
~~~

本机已有 Paramiko。其他电脑缺少时，先持久设置国内源，再安装：

~~~powershell
python -m pip config --user set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
python -m pip install paramiko==5.0.0
~~~

## 2. 选择上游更新并建立集成分支

先整理所选 PR 的提交、必要依赖、接口变化、配置、数据库和前端影响。
不直接 merge upstream/main 到 ljc/custom。
混合前后端的 PR 不能整包 cherry-pick；需要拆分已确认可独立工作的后端补丁，记录原始来源。
依赖会引入额外功能或要求更新前端时，先报告范围再实施。

~~~bash
cd /home/ubuntu/RAG_Weknora
python3 scripts/ljc_status.py
git status --porcelain
git fetch upstream --no-tags
git switch ljc/custom
git switch -c ljc/integrate-YYYYMMDD-topic
git show --stat OFFICIAL_COMMIT_SHA
git diff OFFICIAL_COMMIT_SHA^ OFFICIAL_COMMIT_SHA -- frontend frontend-custom
git cherry-pick -x OFFICIAL_COMMIT_SHA
~~~

将日期、topic和SHA替换为本次确认值。开始前工作区必须干净。
官方代码可从 Gitee 备用镜像取得，但必须比对官方SHA；镜像落后时不能宣称已经同步到最新官方提交。
源码来源追踪保留GitHub官方远端；依赖下载源在构建配置中持久设置。

## 3. 冲突与中断恢复

~~~bash
python3 scripts/ljc_status.py
git status
git diff --name-only --diff-filter=U
~~~

脚本返回码：0表示前端冻结检查通过；2表示前端发生变化；3表示Git操作未完成。
它不会自动重放不确定是否已完成的提交、迁移或部署。

解决冲突后明确暂存修复文件，再执行：

~~~bash
git diff --check
git add CONFIRMED_RESOLVED_FILES
git cherry-pick --continue
python3 scripts/ljc_status.py
~~~

放弃本次尚未完成的选取使用 git cherry-pick --abort。
已完成并发布的提交优先用 git revert 建立反向提交，不使用 reset --hard 丢弃历史。
每个后续阶段测试通过后独立提交，Author和Committer使用ljc，不额外添加其他署名。

## 4. 前端冻结与测试

frontend-custom 的656个原始文件逐文件SHA256记录在 frontend-freeze.json。
原 frontend 与迁移前后端提交中的 frontend 树比对。
导入时保留原有尾随空格，不格式化冻结源码；其他新改动仍执行正常 diff --check。

~~~bash
python3 scripts/ljc_status.py
python3 -m unittest discover -s scripts -p test_ljc_status.py -v
~~~

冻结前端的类型及功能测试（只挂载已有依赖，不下载或发布前端）：

~~~bash
docker run --rm --network none \
  -v "$PWD:/src" \
  -v /home/ubuntu/RAG_Weknora/frontend-custom/node_modules:/src/frontend-custom/node_modules \
  -w /src/frontend-custom node:22-bookworm \
  sh -c 'npm run type-check && npm test -- src/views/knowledge/wiki/documentGraphScope.test.ts src/views/knowledge/archiveMapping.test.ts src/views/knowledge/processingControl.test.ts src/views/knowledge/processingOverview.test.ts'
~~~

## 5. 后端测试和构建

使用服务器已有镜像和缓存。新增构建默认源见 deploy/fork-maintenance/Dockerfile.builder；npm使用原有.npmrc国内源。

~~~bash
docker build --pull=false -f deploy/fork-maintenance/Dockerfile.builder \
  -t weknora-builder:ljc-fork .
mkdir -p runtime/fork
docker run --rm --cpus=4 -e GOMAXPROCS=4 \
  -v "$PWD:/src" \
  -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build \
  -w /src weknora-builder:ljc-fork \
  sh -c 'go test -p 2 ./internal/database ./internal/archivemapping ./internal/archivalcitations ./internal/externalchunks ./internal/processingcontrol ./internal/modelcontext ./internal/application/repository ./internal/application/service ./internal/handler ./internal/router -count=1 && go build -p 2 -buildvcs=false -o runtime/fork/WeKnora ./cmd/server'
~~~

完整后端测试需要正常DNS，以验证URL/SSRF策略；不要用network none运行整套后端测试。依赖下载仍使用镜像内持久国内源配置。
检查测试失败后再修复；不能以健康接口通过替代功能验收。

## 6. 数据库迁移隔离

官方迁移仍位于 migrations/versioned 和 migrations/sqlite，使用schema_migrations。
个人迁移位于 migrations/ljc/postgres 和 migrations/ljc/sqlite，使用独立ljc_schema_migrations。

旧定制 PostgreSQL 91–94分别对应个人账本1–4；SQLite 13–15对应1–3。
兼容桥只识别冻结基线：官方PostgreSQL 90、SQLite 12，并验证已存在的定制表和必要列。
它在一个事务中拆分账本，不执行down migration，不删除或重建业务表；重复启动不会重复应用定制迁移。
脏状态、缺表、来源歧义及未先完成衔接就导入新官方SQL的情况直接失败，不猜测或跳号。

必须先在数据库备份副本验证，再决定生产执行。集成测试仅接受专用数据库名 ljc_migration_test：

~~~bash
docker run --rm --network ljc-weknora-validate \
  --env-file /home/ubuntu/weknora-fork-migration/20261019/migration-test.env \
  -v "$PWD:/src" \
  -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build \
  -w /src weknora-builder:ljc-fork \
  go test ./internal/database -run '^TestLJCMigrationPostgresRestore$' -count=1 -v
~~~

该测试要求测试库刚从旧94版本备份恢复；测试成功后已变为官方90+个人4，再运行前必须恢复专用副本，不能连接生产库。
恢复空库时使用 template0，等最终TCP监听可用，避免ParadeDB初始化阶段的临时服务与预建schema。
凭据env文件和完整恢复命令的现场材料留在私有备份目录，不入Git。

## 7. 推送、发布和回退

测试及公开检查通过后，将集成分支推送到origin，并创建以ljc/custom为目标的PR。
不要向腾讯仓库或自己的main分支发起定制代码合并。
服务器没有GitHub推送凭据时，用本机已授权SSH密钥中转Git bundle；不把本机私钥复制到服务器。

上线必须单独记录候选提交、二进制SHA256、前端冻结检查结果、镜像ID、数据库状态和恢复点。
现有部署脚本部分检查旧分支名称，不能不加核实直接运行。只使用本次维护流程验证过的命令。
后端运行镜像需同时包含新二进制和完整迁移目录；不能把新目录简单覆盖到仍残留旧91–94 SQL的父镜像。

构建后端候选镜像：

~~~bash
release_dir=$(python3 scripts/ljc_prepare_release.py)
docker build --pull=false \
  --build-arg RUNTIME_BASE=weknora-app:document-graph-20261019-195330 \
  --build-arg SOURCE_REVISION="$(git rev-parse HEAD)" \
  -f deploy/fork-maintenance/Dockerfile.runtime \
  -t weknora-app:ljc-candidate-YYYYMMDD "$release_dir"
~~~

首次维护迁移不会自动执行上述镜像的生产切换。正式部署前，需先说明重启、后台任务和账本衔接影响，并核对正在运行的容器配置。
保留原镜像和数据库备份。涉及不兼容数据库变更时，回退包含数据库恢复，不能只切回旧镜像。
仅建立fork和调整Git远端不需要重启应用，也不会自动更新前端。
