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

## 8. 重做隔离数据库验收

以下命令仅操作本次创建的测试容器和测试库，不连接生产容器。先核对专用标记；标记不符立即停止。

~~~bash
cd /home/ubuntu/RAG_Weknora
test "$(docker inspect -f '{{index .Config.Labels "ljc.operation"}}' ljc-weknora-migration-postgres)" = weknora-fork-20261019
docker start ljc-weknora-migration-postgres
docker exec ljc-weknora-migration-postgres sh -c 'pg_isready -h 127.0.0.1 -U "$POSTGRES_USER"'
~~~

等待上述TCP检查成功后，恢复专用副本。这会清空 ljc_migration_test 测试库，生产库不在此容器中。

~~~bash
docker exec ljc-weknora-migration-postgres sh -ec '
dropdb -U "$POSTGRES_USER" --if-exists --force ljc_migration_test
createdb -U "$POSTGRES_USER" -T template0 ljc_migration_test
pg_restore -U "$POSTGRES_USER" -d ljc_migration_test --no-owner --no-privileges --exit-on-error /backup/database.dump
'
~~~

再执行第6节的Go恢复验收命令，成功后查询账本应得到90和4，dirty均为false：

~~~bash
docker exec ljc-weknora-migration-postgres sh -c 'psql -U "$POSTGRES_USER" -d ljc_migration_test -Atc "select version,dirty from schema_migrations; select version,dirty from ljc_schema_migrations;"'
docker stop ljc-weknora-migration-postgres
~~~

测试卷和备份保留，停止测试容器不影响生产。

## 9. 当前服务器没有推送密钥时手动发布

本机Windows已有GitHub SSH授权；服务器不复制本机私钥。先在服务器生成仅包含维护分支的Git bundle，完成公开检查后再下载，勿传私有备份：

~~~bash
cd /home/ubuntu/RAG_Weknora
python3 scripts/ljc_status.py
git status --short
git bundle create /home/ubuntu/weknora-fork-migration/20261019/public-custom.bundle ljc/custom
git bundle verify /home/ubuntu/weknora-fork-migration/20261019/public-custom.bundle
~~~

在Windows PowerShell中，进入已有本地克隆目录或准备新的E盘发布目录，下载时隐藏输入Ubuntu密码：

~~~powershell
New-Item -ItemType Directory -Force E:/weknora-publish | Out-Null
Set-Location E:/weknora-publish
@'
import getpass
import paramiko
c = paramiko.SSHClient()
c.load_system_host_keys()
c.set_missing_host_key_policy(paramiko.RejectPolicy())
c.connect("10.20.0.15", username="ubuntu", password=getpass.getpass("Ubuntu SSH password: "), allow_agent=False, look_for_keys=False, timeout=15)
try:
    with c.open_sftp() as s:
        s.get("/home/ubuntu/weknora-fork-migration/20261019/public-custom.bundle", "public-custom.bundle")
finally:
    c.close()
'@ | python -
~~~

首次连接需通过可信渠道核对服务器SSH主机指纹并保存到known_hosts；不要关闭主机密钥校验。首次发布可建立独立bare仓库：

~~~powershell
git init --bare publish.git
git -C publish.git remote add origin git@github.com:VvMyWife/WeKnora.git
~~~

后续每次在同一目录执行：

~~~powershell
git -C publish.git fetch ../public-custom.bundle refs/heads/ljc/custom:refs/heads/ljc/custom
git -C publish.git push origin refs/heads/ljc/custom:refs/heads/ljc/custom
git -C publish.git ls-remote origin refs/heads/ljc/custom
~~~

核对输出SHA与服务器git rev-parse HEAD一致。只推指定分支，不使用push --all、--mirror或--force。

## 10. 本次阶段记录与源码回退

| 阶段 | 已完成提交 |
| --- | --- |
| 原始服务器定制基线 | 5cb8382b80ac7395feec1654bb92195a9812f44e |
| 基线与fork维护规则 | a355b077f3e32c6deeebf67d5d790b937b4aef597 |
| 冻结前端纳入统一源码 | 741624726acbc958380df1d64ca769ea2abd6647 |
| Git中断检测和冻结校验 | f20cf2fa4f3c5010cf48d004ea0ea82dd9c0e710 |
| 官方及个人迁移账本隔离 | 250b3299e99a70e016e6707fcd5ec1fa909f2203 |
| 测试及维护使用说明 | 31634e52c193e0735fbed1327a55d0ff636bf687 |

本次候选后端镜像源提交为31634e52，属于准备完成但未部署的产物。完整对应信息见delivery.json。
生产仍运行原来的document-graph镜像，生产账本仍为94；只有恢复副本执行了90+4衔接。

查看原始源码无需切换现有前端工作目录，可从保留的原始分支建立独立回退工作区：

~~~bash
cd /home/ubuntu/RAG_Weknora
git show --stat 5cb8382b80ac7395feec1654bb92195a9812f44e
git worktree add -b ljc/rollback-review-YYYYMMDD /home/ubuntu/RAG_Weknora-rollback-YYYYMMDD 5cb8382b80ac7395feec1654bb92195a9812f44e
~~~

原独立前端Git历史另有frontend.bundle，完整后端历史为backend.bundle，私有配置归档和database.dump均保留在/home/ubuntu/weknora-fork-migration/20261019。
不要把这些私有恢复材料推送到公开fork。已发布错误优先revert；真正恢复数据库之前必须核对停机范围和备份时间。
