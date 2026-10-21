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
2026-10-21已获准接入全部官方前端功能；28182在新版构建及验收前仍运行review-only旧候选，具体版本以产物清单为准。
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


## 8. 重排验收入口

2026-10-21起使用下文GPU重排步骤。旧CPU脚本及结果只保留为历史对照，不作为当前启动方式。
候选模型URL为Docker桥接内网http://192.168.16.1:18085；真实后端检索必须返回meta.rerank.applied=true且outcome=ok。

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

## 复核本轮运行证据和候选产物

以下命令只检查隔离环境及私有验收结果，不执行生产迁移或切换。
在服务器先进入集成工作区：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 scripts/ljc_status.py
python3 scripts/ljc_integration_selfcheck.py
python3 scripts/ljc_capture_acceptance_manifest.py
python3 -m json.tool docs/upstream-integration/artifact-manifest.json
```

完整重跑后端、官方前端和构建的命令见上文。artifact-manifest中的镜像标签、不可变image_id和image revision
用于核对实际运行的候选；工作区dist哈希单独列出，不能把工作区文件自动当成服务已部署内容。
本轮只交付候选：28180为官方界面，28182为定制候选界面（具体源码及验收状态见产物清单）；生产18180/18182仍使用原镜像。

本轮实例的运行验收脚本保存在私有目录，所需密码从acceptance.env读取，禁止输出或提交此文件。
以下第一条为只读复核，第二条在已有本轮样本时只读，第三条会在已存在的隔离扫描件上重新解析，运行前等当前任务完成：

```bash
python3 /home/ubuntu/weknora-fork-migration/20261020-upstream/verify-wiki-doc-graph.py
python3 /home/ubuntu/weknora-fork-migration/20261020-upstream/acceptance-entity-graph.py
python3 /home/ubuntu/weknora-fork-migration/20261020-upstream/acceptance-scan-metadata.py
```

实体图谱脚本在已有本轮样本时仅查询；样本不存在时会创建专用测试知识库及上传合成文件。
不要重跑旧acceptance-scan-control.py：其首次引擎名称错误的历史记录保留用于排查。
完整故障重试证明见failed-retry-results.json；故障服务已停止，不能直接重跑依赖18084的脚本。
需要重做故障注入时先按上文启动专用故障服务，再运行acceptance-metadata-retry.py，完成后停止该专用进程。

## 2026-10-21 GPU重排（候选环境）

生产向量模型原本使用GPU。此前仅为隔离验收临时启动CPU重排；现已停止该临时CPU进程，候选模型URL改为Docker桥接内网18085。
使用现有缓存权重和持久化Python环境，不下载模型、不停止其他GPU服务。启动前检查选定GPU至少4GiB空闲：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
nvidia-smi
export LJC_ACCEPTANCE_GPU_UUID=GPU-918824a7-2144-fe49-5183-7646ab974946
nohup bash scripts/ljc_acceptance_gpu_rerank.sh > /home/ubuntu/weknora-fork-migration/20261020-upstream/gpu-rerank.log 2>&1 < /dev/null &
curl --fail http://192.168.16.1:18085/health
python3 /home/ubuntu/weknora-fork-migration/20261020-upstream/acceptance-gpu-rerank.py
```

期望health.device为cuda:0（CUDA可见设备被限定到上面UUID，即物理GPU1），后端meta.rerank.applied=true且outcome=ok。
12段相同输入、预热2次后5次测量：CPU中位2923.83ms，GPU中位45.33ms，仅代表本机该重排样本，不代表端到端聊天耗时。
不再启动旧CPU辅助脚本；CPU结果文件仅保留为历史对照。

## 定制前端完整接入后的复核

源码位于同一权威仓库的隔离工作区frontend-custom。保留档案智查品牌、原件映射/版本、外部分块、档案引用、全库进度/暂停和按档案图谱。
官方上传任务队列接管上传进度，切换页面后任务继续；全库处理进度面板仍单独显示。
每次修改后按以下顺序复核，生产端口不在这些命令范围内：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 -m unittest discover -s scripts -p test_ljc_status.py
python3 scripts/ljc_status.py
bash scripts/ljc_integration_job.sh custom-frontend
python3 scripts/ljc_integration_selfcheck.py
```

定制目录在外层历史Git规则中被忽略，因此新增源码必须显式git add -f具体文件，不能把node_modules/dist/临时文件强制加入。
获批源码校验使用custom-frontend-candidate.json；原frontend-freeze.json保留作历史基线。
SSR单元测试关闭无用watch，并限制并发2；不改主机inotify设置。完整依赖从持久化国内镜像配置安装。
PDF.js保留定制版本6.4.299，canvas固定1.0.11以保证多平台锁文件完整；所有npm包下载地址使用npmmirror。
