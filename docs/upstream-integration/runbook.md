# 集成候选运行说明

权威 Git 仓库：/home/ubuntu/RAG_Weknora；本轮隔离工作区：/home/ubuntu/RAG_Weknora-upstream-20261020。
全部操作以 ubuntu 执行。固定官方目标 eecfe4b3，生产保持 a418adc45 和旧94号账本。
本说明只操作隔离验收；生产升级、数据库恢复和服务切换须另行确认维护窗口。

## 1. 中断后先自检

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short --branch
python3 scripts/ljc_status.py
python3 scripts/ljc_integration_selfcheck.py
docker ps --filter label=ljc.operation=upstream-20261020
tail -n 40 runtime/upstream/jobs/backend-tests.resume.log
~~~

自检只读核对生产六个容器ID/镜像、生产94 clean账本、原工作区HEAD和验收存储隔离。
长任务用下述安全启动器；重复启动会因锁退出75，不并行重放。
它仅允许测试/构建。状态running但对应进程已消失时，下述命令在先通过生产边界检查后自动重启这些任务；失败状态保留，须诊断后重跑。
迁移、数据库恢复、Compose启动不在自动重启名单内。

~~~bash
python3 scripts/ljc_integration_selfcheck.py --resume-safe-jobs
nohup bash scripts/ljc_integration_job.sh backend-tests > runtime/upstream/backend-tests-manual.log 2>&1 < /dev/null &
cat runtime/upstream/jobs/backend-tests.status
tail -n 40 runtime/upstream/backend-tests-manual.log
~~~

## 2. 持久化工具链与国内源

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
docker build -f deploy/upstream-integration/Dockerfile.builder -t weknora-builder:upstream-20261020 .
docker build -f deploy/upstream-integration/Dockerfile.frontend-builder -t weknora-frontend-builder:upstream-20261020 .
docker build -f deploy/upstream-integration/Dockerfile.rust-builder -t weknora-rust-builder:upstream-20261020 .
docker build -f deploy/upstream-integration/Dockerfile.neo4j -t weknora-neo4j:ljc-upstream-20261020 .
~~~

Dockerfile持久配置APT/Python/npm/Cargo/Go国内源；Go模块缓存使用权威仓库.go-modcache，npm使用/home/ubuntu/.npm，Cargo和前端依赖使用命名卷。
不要在候选运行配置中加入test-hosts.txt；它只供使用假客户端的单元测试绕过不稳定公网DNS，不能证明模型服务联网。
现有后端、定制UI和Docreader基础镜像属于私有恢复基线，不应清理；新机器需先从备份恢复对应镜像。禁止用docker system prune清理验收环境。

## 3. 按顺序测试和构建

~~~bash
python3 -m unittest discover -s scripts -p test_ljc_status.py -v
bash scripts/ljc_integration_job.sh backend-tests
bash scripts/ljc_integration_job.sh native-assets
bash scripts/ljc_integration_job.sh native-backend
bash scripts/ljc_integration_job.sh official-frontend
bash scripts/ljc_integration_job.sh docreader-image
~~~

后端候选构建依赖原生库、BrowserSkill及五份历史本地预装技能资产。技能是备份文件，不是新版租户技能目录的自动注册证明。

~~~bash
mkdir -p runtime/upstream/preloaded-skills
tar -tf /home/ubuntu/weknora-fork-migration/20261020-upstream/preloaded-skills.tar.gz | head
tar -xzf /home/ubuntu/weknora-fork-migration/20261020-upstream/preloaded-skills.tar.gz -C runtime/upstream/preloaded-skills --strip-components=1
bash scripts/ljc_integration_job.sh backend-image
bash scripts/ljc_integration_job.sh custom-frontend
~~~

定制前端严格构建目前因韩语/俄语15个翻译键缺失而失败：416项中414通过、2失败，类型检查通过。
28182当前是单独生成的review-only预览，不能当作严格验收通过。前端源码尚未按官方模块移植。
不得跳过这两项失败测试来标记正式通过。

## 4. 新建独立数据库并显式迁移

首次创建外部隔离网络（已存在时不必重建）：

~~~bash
docker network inspect ljc-upstream-validate >/dev/null 2>&1 || docker network create --internal --label ljc.operation=upstream-20261020 ljc-upstream-validate
python3 scripts/ljc_prepare_acceptance.py
export LJC_ACCEPTANCE_ENV=/home/ubuntu/weknora-fork-migration/20261020-upstream/acceptance.env
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml config --quiet
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml up -d postgres redis neo4j docreader
docker run --rm --network ljc-upstream-validate --env-file "$LJC_ACCEPTANCE_ENV" --entrypoint /app/ljc-migrate weknora-app:ljc-upstream-eecfe4b3 -mode status
docker run --rm --network ljc-upstream-validate --env-file "$LJC_ACCEPTANCE_ENV" --entrypoint /app/ljc-migrate weknora-app:ljc-upstream-eecfe4b3 -mode up
docker run --rm --network ljc-upstream-validate --env-file "$LJC_ACCEPTANCE_ENV" --entrypoint /app/ljc-migrate weknora-app:ljc-upstream-eecfe4b3 -mode up
~~~

应为官方128 clean、个人4 clean。初始化先vector再pg_search，独立ParadeDB0.25.11。
旧94生产账本不得直接执行官方up。旧副本衔接和恢复步骤见README.md。
专用实例内另留有早期失败的ljc_acceptance库（旧迁移验证容器），其8号dirty状态保留，没有force；此处使用全新的ljc-upstream-acceptance-postgres容器，二者不可混淆。

## 5. 启动和浏览器检查

~~~bash
bash scripts/ljc_start_acceptance.sh
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml --profile custom up -d custom-ui
curl -fsS http://127.0.0.1:28081/health
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml ps
~~~

浏览器打开http://10.20.0.15:28180及http://10.20.0.15:28182。
首次使用acceptance.env中的测试邮箱/密码注册；用户名用ljc_acceptance（用户名不允许横杠）。
首次注册后仅重启ljc-upstream-acceptance-app，再登录，可通过引导邮箱成为候选系统管理员。
不要打印整个env文件或docker compose config；后者应始终使用--quiet。

本地模型通过新目录配置：qwen35-9b=http://10.20.0.15:8101/v1；bge-m3=http://10.20.0.15:18082/v1，1024维。
重排登记为bge-reranker-v2-m3=http://10.20.0.15:18083，目前服务未监听，回退顺序不能算真实重排通过。
内置快速问答需在智能体“模型配置”选择模型并保存，再进行聊天测试。

检查真实Markdown/表格上传、解析/摘要、关键词/向量/混合召回、原件PDF绑定后预览下载、重复外部ZIP预检及导入、聊天引用、任务暂停继续。
验收矩阵见acceptance.md，未完成项目不能用健康检查或单测替代。

## 6. 原子序号回归

~~~bash
docker run --rm --network ljc-upstream-validate --env-file "$LJC_ACCEPTANCE_ENV" -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec 'export LJC_EXTERNAL_POSTGRES_TEST_DSN="$LJC_MIGRATION_DSN"; go test ./internal/application/service -run "^TestExternalPersist" -count=1'
~~~

只接受名为ljc_acceptance的数据库；测试使用临时表、临时序列和事务回滚，验证普通分块→外部分块→摘要不会撞号，软删除重试保持原序号。SQLite既有测试同样保留。

## 7. 回退边界

当前尚未部署生产，候选失败时生产继续使用原镜像及94号数据库。
只停止候选容器可执行下述命令；保留所有数据卷，不加-v：

~~~bash
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml --profile custom stop
~~~

恢复候选前先记录镜像ID及数据备份；新结构不能交给旧镜像盲读。将来生产发生数据库或扩展升级时，回退必须包含对应数据库备份恢复，切回镜像不足以回退。
本轮验收未全部通过前，不推进ljc/custom、不推送最终完成状态。原工作区config/config.yaml被生产bind-mount，不能用checkout覆盖实时配置。


## 8. 候选CPU真实重排（可选）

原生产10.20.0.15:18083未监听。已有权重和Python依赖可用于CPU验收，无需下载、不占GPU，只绑定隔离Docker网桥地址。

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
nohup bash scripts/ljc_acceptance_cpu_rerank.sh > /home/ubuntu/weknora-fork-migration/20261020-upstream/cpu-rerank-manual.log 2>&1 < /dev/null &
gateway=$(docker network inspect -f '{{(index .IPAM.Config 0).Gateway}}' ljc-upstream-validate)
curl -fsS "http://$gateway:18083/health"
~~~

候选acceptance.env的SSRF_WHITELIST应包含该网桥IP（本机192.168.16.1），只在候选模型配置中把重排URL设为http://192.168.16.1:18083。
修改候选env后运行bash scripts/ljc_start_acceptance.sh；脚本在后端健康后刷新候选nginx的上游解析。
生产URL、模型配置、GPU和服务均不改变。使用真实权重的返回meta.rerank必须applied=true且outcome=ok，回退不算通过。
CPU模型依赖使用已有/home/ubuntu/RAG/runtime/eval-venv和/home/ubuntu/RAG/runtime/models/BAAI/bge-reranker-v2-m3；不同机器须配置对应路径。

## 9. 本机真实回归脚本和原始证据

本次现场脚本、凭据、生成文件与原始日志保存在私有恢复目录。后续重跑前先读取对应脚本和状态，部分脚本会在候选库新增或重解析样例，不能盲目并发重放。

~~~bash
private=/home/ubuntu/weknora-fork-migration/20261020-upstream
python3 "$private/verify-wiki-doc-graph.py"
python3 "$private/acceptance-permissions.py"
python3 "$private/acceptance-archive-versions.py"
python3 "$private/acceptance-cpu-rerank.py"
cat "$private/scan-control-results.json"
cat "$private/archive-version-results.json"
cat "$private/permissions-results.json"
~~~

扫描PDF通过正确engine=paddleocr_vl调用真实9960服务。Paddle页完整性证据在knowledge_processing_spans的docreader输出中，expected=2、processed=2、integrity=verified；不能只看parse_status=completed。
故障注入器scripts/ljc_acceptance_fault_server.py只监听同一内网网桥18084，健康检查成功但OCR POST故意返回500；只用于失败恢复测试，完成后停止其进程。不得用于生产模型或OCR配置。

## 10. 镜像恢复材料

基线四个应用/前端/Docreader镜像另存私有rollback-images.tar.gz，SHA256与完成状态文件同目录。只有状态为0且校验成功才是完整备份。

~~~bash
cd /home/ubuntu/weknora-fork-migration/20261020-upstream
cat rollback-images.status
sha256sum -c rollback-images.sha256
# 在镜像确实丢失、准备恢复时执行；只导入镜像，不切换服务：
gzip -dc rollback-images.tar.gz | docker image load
~~~

生产数据库恢复依然必须使用匹配的数据库备份，不能仅导入或切回旧镜像。
