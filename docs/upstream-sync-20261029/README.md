# WeKnora 官方主线同步：2026-10-29（10月30日完成验收交付）

本轮固定官方目标为 `e8e0d2ed4bd660c46313884d2f2b3b243d7c318d`，从定制 `1f8564d64e7feb78afd5ea202f78205875c75e2f` 合并201个官方提交、702个文件，其中官方前端157个文件。新提交留待下一批。

**本轮不发布生产18182，不迁移正式数据库。** 18180保持原状，28182继续退役。全部新增官方功能适配到frontend-custom，同时保留品牌、布局、原件预览、档案Wiki/图谱、引用证据、外部分块、PDF页定位、任务控制和模型预算。

## Git阶段及产物

|阶段|提交|内容|
|---|---|---|
|冻结|f61a5f2e0|完整201提交/702文件清单、定制交集、恢复bundle和生产快照|
|正式merge|f8be70abe|保留官方历史，保留Asynq无限超时、Paddle/PDF校验和证据绑定|
|前端适配|a7b2502f3|三方移植157文件增量；保留全部定制功能|
|验收工具|92b42a036|隔离迁移、TLS Docker沙箱、完整测试和镜像构建|
|容器权限|ba6e9d055、0cc0301cc|修复私有umask下配置/迁移/静态资源的读取权限|
|检索兼容|1ff4282bd|ParadeDB评分后物化，再排除生成问题，再取TopK|
|PPTX兼容|fda761a6b|将WMF/EMF/SVG内嵌data URI光栅化，接入图片处理|

Author/Committer均为 `ljc <ljc@local>`。维护分支为 `ljc/custom`，集成分支为 `ljc/integrate-20261029-upstream-main`。最终维护SHA和远端核对见交付目录delivery-status.json，不强推、不覆盖并发提交。

候选镜像源码为 **fda761a6bbdbd7158b7fe2544ee4567b86e44910**；镜像ID见artifact-manifest.json。后续文档/验收工具提交与镜像源码分别记录。

## 已通过的检查及边界

- 完整相关后端测试、Asynq补丁、原生AnyDoc测试与编译；BM25修复后完整后端重跑。
- 官方前端2355项：2354通过、1条件跳过；定制前端2400项：2399通过、1条件跳过；两套类型检查和构建通过。
- DocReader496项：484通过、12条件跳过，含真实WMF PPTX回归。
- Go SDK与CLI测试/构建；MCP默认6项及SSE/preflight协议31项。
- PostgreSQL空库、正式128备份、旧133候选备份均升级为官方134＋个人4，重复up通过；两份真实副本保护表记录数和全行内容指纹不变。SQLite51＋个人3通过。
- 独立TLS Docker沙箱两轮创建、Unicode文件读写、执行、超时终止、删除和删除后访问拒绝。
- 真实扫描PDF OCR提取事实17、Excel/CSV、真实WMF PPTX媒体入库；默认配置保存读取、关键词/向量/混合检索、实际GPU重排。
- 原件双版本、历史下载及分块不变、外部分块幂等导入；跨空间访问拒绝、模型凭据隐藏、自动化CRUD；申请人自审被拒绝、管理员审批与只读角色授予通过。
- 普通问答和智能体最终答案的生成时引用；档案Wiki/图谱范围隔离和关系实际端点。
- 失败摘要重试、重复点击只接纳一份任务、文本分块不变、暂停恢复、取消后不被旧任务覆盖。
- dirty账本被拒绝且数据不变；通过完整备份恢复独立副本至官方128＋个人4。
- 浏览器原件PDF、档案入口、默认配置、上下文助手及图谱证据见acceptance-results.json。

共享生产Qwen9B验收时出现排队停滞，未重启它。本轮独立GPU1使用同一已有模型缓存的4-bit量化实例，16K上下文、8K输出预算；不能据此宣称生产原精度49K上下文或长任务压力验收通过。BGE保持GPU运行。模型生成的完整事实准确率、大规模吞吐、第三方服务真实凭据/全提供方组合不在本轮合成夹具验收内。

BM25物化保留过滤在LIMIT之前的正确语义；高频词匹配量很大时需另做性能压测。PPTX小于64×64的装饰图过滤保留；验收使用文档尺寸矢量图。真实EMF文件未跑，EMF路由有单元测试。

## 路径和生产边界

- 共享权威Git仓库：/home/ubuntu/RAG_Weknora。
- 集成工作区：/home/ubuntu/RAG_Weknora-upstream-20261029。
- 维护分支工作区：/home/ubuntu/RAG_Weknora-upstream-20261023。
- **生产配置挂载：/home/ubuntu/RAG_Weknora-upstream-20261020/config。禁止checkout/reset或替换该目录。**
- 私有验收材料：/home/ubuntu/weknora-fork-migration/20261029-upstream（700）。
- 日志：集成工作区runtime/upstream-20261029。
- 正式后端源码仍为f45137d3fd784e076336fcc11b50d17ba5fe85c0；18182前端仍为e5df8a440013e960ed3d865a21c3a8a4b914debc。
- 正式数据库仍为官方128＋个人4，clean。候选使用独立Postgres、Redis、Neo4j、存储、队列和日志，仅绑定127.0.0.1:28281/28282。
- 验收结束停止本轮服务，保留卷、镜像、备份和证据。

## 1．登录和中断自检

```bash
ssh ubuntu@10.20.0.15
cd /home/ubuntu/RAG_Weknora-upstream-20261029
git status --short --branch
git log -10 --oneline
git merge-base --is-ancestor e8e0d2ed4bd660c46313884d2f2b3b243d7c318d HEAD
bash scripts/ljc_sync_20261029.sh status
docker ps -a --filter label=ljc.operation=upstream-20261029 --format '{{.Names}} {{.Status}}'
python3 deploy/upstream-sync-20261029/migration_proof.py --status
```

先看PID、锁、容器和日志，运行中不要重放。测试/构建入口用flock拒绝同类并发。迁移和故障演练遇已有状态拒绝重放；失败不能标成成功，不通过删除保护材料绕过拒绝。

## 2．逐项重跑测试

复用服务器持久构建镜像和缓存。Go使用goproxy.cn及sum.golang.google.cn，npm使用npmmirror，Python使用清华镜像；不改生产配置和宿主inotify限制。

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261029
bash scripts/ljc_sync_20261029.sh backend
bash scripts/ljc_sync_20261029.sh custom-ui
bash scripts/ljc_sync_20261029.sh official-ui
bash scripts/ljc_sync_20261029.sh docreader
docker run --rm -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec 'cd client; go test ./...; go build ./...'
docker run --rm -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec 'cd cli; go test ./...; go build ./...'
bash deploy/upstream-sync-20261029/acceptance/run-mcp-tests.sh
```

真实BM25回归只在专用ljc_sync_数据库创建事务内临时schema，结束回滚：

```bash
docker start ljc-sync-20261029-postgres
docker run --rm --network ljc-sync-20261029 --env-file /home/ubuntu/weknora-fork-migration/20261029-upstream/bm25-test.env -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 go test ./internal/application/repository/retriever/postgres -run TestKeywordGeneratedQuestionBM25 -count=1 -v
```

## 3．候选构建与映射

源码先提交，构建期间不要编辑；脚本结束核对SHA和干净状态，变化则拒绝交付该次产物。

```bash
git status --short
bash scripts/ljc_sync_20261029.sh images
cat runtime/upstream-20261029/artifact-manifest.json
```

重建产生当前HEAD标签，不能把新SHA填到旧镜像上。初次包装权限失败、装饰图夹具过滤、缺少测试init/testdata的失败日志保留；最终成功以本目录验收清单为准。

## 4．迁移证明和恢复

```bash
python3 deploy/upstream-sync-20261029/migration_proof.py --status
cat /home/ubuntu/weknora-fork-migration/20261029-upstream/rollback-proof.json
```

无参数migration_proof.py是一次性创建/备份/迁移证明，已有实例或快照会拒绝重放。rollback-proof.py仅对专用ljc_sync_rollback128注入dirty并验证拒绝，然后完整恢复128快照；已完成时拒绝重复执行。不用force跳过错误，不连接生产执行故障注入。

保留pre-sync.bundle、production-baseline.json、production-128-snapshot.dump、candidate-133-snapshot.dump和独立恢复卷，均在服务器私有目录。密码、token、用户数据和数据库不入Git。

## 5．启动隔离候选并运行真实接口

先检查端口和GPU显存，不能为验收重启其他模型服务。保留的独立Qwen容器只接测试网络，无主机端口。

```bash
ss -ltn '( sport = :28281 or sport = :28282 )'
nvidia-smi
docker start ljc-sync-20261029-postgres
docker start ljc-sync-20261029-qwen
docker compose -p ljc-sync-20261029 -f /home/ubuntu/weknora-fork-migration/20261029-upstream/compose.acceptance.json up -d
curl --fail http://127.0.0.1:28281/health
curl --fail http://127.0.0.1:28282/config.js
```

等待模型实际可生成。测试用户、模型ID和私有.env已在服务器保留。源码副本不含密码，更新后复制到私有目录逐项执行，不并发运行会修改共享state的脚本：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261029
install -m 600 deploy/upstream-sync-20261029/acceptance/acceptance-*.py /home/ubuntu/weknora-fork-migration/20261029-upstream/
cd /home/ubuntu/weknora-fork-migration/20261029-upstream
python3 acceptance-realfile.py
python3 acceptance-archive-external.py
python3 acceptance-archive-versions.py
python3 acceptance-permissions.py
python3 acceptance-automation-crud.py
python3 acceptance-gpu-rerank.py
python3 acceptance-wiki-graph.py
python3 acceptance-graph-scope.py
python3 acceptance-generation-agent-citation.py
python3 acceptance-generation-quick-citation.py
python3 acceptance-formats-defaults.py
```

格式脚本重新执行会新建测试知识库。失败重试夹具只适用于确实存在失败阶段的独立文档；本轮已利用模型停滞产生的失败完成恢复。成功恢复后不能把“没有失败可重试”当成再次故障演练。结果见control-retry-results.json。

重新生成合成文件：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261029
mkdir -p runtime/upstream-20261029/fixtures
cp deploy/upstream-sync-20261029/acceptance/make_fixtures.py runtime/upstream-20261029/fixtures/make_fixtures.py
docker run --rm -e PYTHONPATH=/app -v "$PWD/runtime/upstream-20261029/fixtures:/fixtures" --entrypoint /app/.venv/bin/python weknora-docreader:ljc-sync-fda761a6b /fixtures/make_fixtures.py
```

## 6．浏览器检查

在你电脑的另一个终端建立隧道：

```bash
ssh -N -L 28282:127.0.0.1:28282 ubuntu@10.20.0.15
```

打开http://127.0.0.1:28282，使用私有acceptance.env的合成测试账户。检查默认配置保存后重开、上下文助手、MD解析件映射的PDF完整预览、Wiki/图谱文档范围、处理与重试、取消和暂停恢复、最终答案引用及控制台。已有截图在本地交付目录，不用正式用户数据做测试。

## 7．停止候选、保留产物

先等接口和恢复演练结束并保存结果：

```bash
docker compose -p ljc-sync-20261029 -f /home/ubuntu/weknora-fork-migration/20261029-upstream/compose.acceptance.json stop
docker stop ljc-sync-20261029-qwen
docker stop ljc-sync-20261029-docker
docker stop ljc-sync-20261029-postgres
ss -ltn '( sport = :28281 or sport = :28282 or sport = :28182 )'
```

关闭本地隧道，不运行down -v。核对生产容器镜像、StartedAt、挂载及旧配置文件指纹不变。

## 8．维护分支交付和源码恢复

确认维护分支干净且未并发变化后再快进；若已变先整合补测：

```bash
git -C /home/ubuntu/RAG_Weknora-upstream-20261023 status --short --branch
git -C /home/ubuntu/RAG_Weknora-upstream-20261023 merge --ff-only ljc/integrate-20261029-upstream-main
git -C /home/ubuntu/RAG_Weknora-upstream-20261023 merge-base --is-ancestor e8e0d2ed4bd660c46313884d2f2b3b243d7c318d ljc/custom
```

本次从服务器导出已核对bundle，在Windows持久Git缓存中导入，使用已登录的GitHub HTTPS凭据显式推送两条分支。服务器没有可用SSH推送密钥，不扫描别的项目凭据。push.default=nothing保留，不强推。

```bash
git push --atomic https://github.com/VvMyWife/WeKnora.git refs/heads/ljc/custom:refs/heads/ljc/custom refs/heads/ljc/integrate-20261029-upstream-main:refs/heads/ljc/integrate-20261029-upstream-main
git ls-remote https://github.com/VvMyWife/WeKnora.git refs/heads/ljc/custom refs/heads/ljc/integrate-20261029-upstream-main
```

旧源码恢复用新工作区，不reset维护分支或生产配置：

```bash
git -C /home/ubuntu/RAG_Weknora bundle verify /home/ubuntu/weknora-fork-migration/20261029-upstream/pre-sync.bundle
git -C /home/ubuntu/RAG_Weknora worktree add -b ljc/recover-20261029-baseline /home/ubuntu/RAG_Weknora-recovery-20261029 1f8564d64e7feb78afd5ea202f78205875c75e2f
```

正式发布另行确认维护窗口、任务处理、当时备份、迁移和切换顺序。数据库升级后回退必须恢复数据库及相应存储快照，不能只换旧镜像。
