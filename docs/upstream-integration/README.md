# 当前交付状态（2026-10-21）

官方固定目标eecfe4b3全量合并，18182对应定制前端已接入六组官方功能，同时保留原有新增能力、品牌、布局与原件预览。用户已批准全部模块，不再等待设计范围选择。
当前运行的是隔离候选：后端28081、官方界面28180、定制界面28182。生产18180/18182和生产数据库均未切换。

开始使用请按[runbook.md](runbook.md)中的自检、测试、构建、启动和回退步骤操作；验收结果见[acceptance.md](acceptance.md)，前端模块及外部集成限制见[frontend-modules.md](frontend-modules.md)，源码—镜像映射见[artifact-manifest.json](artifact-manifest.json)。
1479项定制前端测试、1448项官方前端测试、160个后端测试包和331项Docreader测试通过；Docreader有12项原有跳过。候选重排已改为物理GPU1。

以下内容按实施阶段保留，描述的是当时状态；当前状态以本节、checkpoint.json及验收记录为准。

# 全量上游集成 2026-10-20

权威仓库 /home/ubuntu/RAG_Weknora；隔离工作区 /home/ubuntu/RAG_Weknora-upstream-20261020。
定制起点 a418adc45f383b9accb0d51bdcf226c468e656fb，官方目标 eecfe4b3c8c34c94e9b559e5dd02187b4ae447cd。
分支 ljc/integrate-20261020-upstream-main。用户已批准完整合并，旧勾选限制作废；只在业务功能取舍和18182模块设计去向上询问。

## 边界及基线
只使用隔离工作区、专用数据副本及候选28081/28180/28182。不改生产数据库及18180/18182服务；复制数据不得触发真实外发或生产任务。
18180完整官方候选；18182定制模块经用户看过对比后确认。
私有备份：/home/ubuntu/weknora-fork-migration/20261020-upstream，含backend.bundle、database.dump、source-and-private-config.tar.gz和baseline-private.json，恢复材料不入公开Git。
隔离检出发现frontend-custom/Dockerfile.local被本地忽略漏跟踪，本阶段原样补入，字节与冻结清单一致。

## 分阶段提交
1. 基线、备份、维护规则、可重复检出的源码保护。
2. schema-only状态/旧账本引导/升级入口，副本94→90+个人4测试。
3. 正式merge官方SHA，解决技术冲突并保留定制接口。
4. OCR/解析、队列、引用/原件/外部分块、文档级图谱、模型权限分别测试提交。
5. 官方前端候选及定制模块对比，按用户决定移植。
6. 联合验收、镜像映射、合入ljc/custom并推送fork；生产另确认。

## 手工核对
~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short --branch
python3 scripts/ljc_status.py
python3 -m unittest discover -s scripts -p test_ljc_status.py -v
git -C /home/ubuntu/RAG_Weknora status --short --branch
docker inspect -f '{{.Name}} {{.Id}} {{.Config.Image}}' WeKnora-app WeKnora-frontend WeKnora-frontend-custom-preview
~~~
每阶段测试通过后独立提交，作者和提交者ljc，记录SHA；后续补充实际测试、构建和恢复命令。
中断先核对Git、进程和数据库状态，不盲目重放；依赖镜像持久化配置，不提交凭据、缓存和数据。

## Schema-only迁移工具

只读取环境变量LJC_MIGRATION_DSN或LJC_SQLITE_PATH，禁止将密码写在命令参数和Git中。模式：status只读状态；bootstrap-legacy按冻结SQL指纹识别旧个人账本并拆分；up运行官方和个人迁移，不启动应用任务。
冻结SQL来自a418adc45中的官方PostgreSQL90/SQLite12；manifest校验每个文件SHA256。dirty、缺表、歧义或篡改均失败，不执行force或业务表清理。

构建（使用已有镜像及持久化国内依赖镜像）：
~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
docker run --rm -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -c 'go test ./internal/database ./cmd/ljc-migrate -count=1 && mkdir -p runtime/upstream && go build -o runtime/upstream/ljc-migrate ./cmd/ljc-migrate'
~~~

专用备份副本恢复完成且状态为94/clean后，运行真实数据行数及重复执行验收：
~~~bash
docker run --rm --network ljc-upstream-validate --env-file /home/ubuntu/weknora-fork-migration/20261020-upstream/migration-legacy.env -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 go test ./internal/database -run '^TestLJCMigrationPostgresRestore$' -count=1 -v
~~~
该测试只接受ljc_migration_test数据库名。完成后账本已拆分，不可直接当旧94再跑；先恢复专用副本。

检查及重复引导（只使用已创建的专用环境文件）：
~~~bash
docker run --rm --network ljc-upstream-validate --env-file /home/ubuntu/weknora-fork-migration/20261020-upstream/migration-legacy.env -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod:ro -w /src weknora-builder:upstream-20261020 runtime/upstream/ljc-migrate -mode status
docker run --rm --network ljc-upstream-validate --env-file /home/ubuntu/weknora-fork-migration/20261020-upstream/migration-legacy.env -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod:ro -w /src weknora-builder:upstream-20261020 runtime/upstream/ljc-migrate -mode bootstrap-legacy
~~~

## 副本恢复与扩展依赖

恢复任务和凭据留在私有备份目录，不入Git。restore-legacy.sh仅操作带upstream-20261020标签的ljc-upstream-legacy-postgres；restore-target.sh仅操作同标签的ljc-upstream-postgres。两者都只重建专用ljc_migration_test库，使用template0，完整pg_restore带--exit-on-error。
新版ParadeDB先创建vector扩展，再恢复旧备份，满足pg_search新的依赖顺序。两个专用实例maintenance_work_mem为2GB、max_parallel_maintenance_workers为0，串行构建HNSW，避免Docker默认共享内存不足；生产配置不改。状态文件在开始时写running，完整恢复成功才写0，其他退出码均为失败。
运行及检查退出码：
~~~bash
backup=/home/ubuntu/weknora-fork-migration/20261020-upstream
bash "$backup/restore-legacy.sh" > "$backup/legacy-restore-manual.log" 2>&1
cat "$backup/legacy-restore.status"
bash "$backup/restore-target.sh" > "$backup/target-restore-manual.log" 2>&1
cat "$backup/target-restore.status"
~~~
恢复会清空专用测试库，先确认未在其上运行后续验收。长任务不因终端超时自动重放，先查状态文件及进程。

## 已完成的阶段验证

阶段1提交：daeb98f8fdce2a500f36c7dfc4302129f6d81203；冻结检查相关7项测试通过。
阶段2：数据库及CLI单元测试通过；SQLite新库、旧15账本、缺表、dirty和SQL篡改场景通过。
2026-10-20在ParadeDB0.22.2及0.25.11的两个真实完整备份副本上，94→90+个人4、连续两次up均通过；knowledges、chunks、wiki_pages、archive_mapping_versions、message_citation_evidence、enrichment_recoveries行数保持一致。
这只证明账本衔接；官方128升级和应用业务验收仍待阶段3以后验证。
阶段2恢复：代码可在隔离分支回到daeb98f8f；已拆分的测试数据库通过私有restore脚本恢复94账本。生产库从未执行本工具。

阶段3进展：真实副本已在ParadeDB0.25.11升级到官方128+个人4（均clean），重复up通过；与同一备份的旧版90+4副本对比，knowledges7313、chunks232877、wiki_pages19327、archive_mapping_versions639、message_citation_evidence5、enrichment_recoveries100，行数一致。
SQLite新库和旧15桥接升级到45+个人3自动测试通过。原始94账本的生产库保持不变。
部分官方测试需兼容定制行为：失败任务冷却期、有效PDF页数校验；测试使用过期冷却时间及真实单页PDF，原有排序、超时和取消断言保留。官方查询实体预算已升为8192，后台图谱保持定制12288。

## 全量主线合并的验证与兼容决策

后端内部包、cmd/server、迁移CLI、独立client模块及CLI模块测试共160个含测试的包通过；标准后端和schema-only工具均编译成功。启用anydoc后的解析测试和原生后端编译通过，Asynq无限超时及取消两项本地补丁测试通过。
官方前端类型检查、1446项测试、生产构建全部通过；frontend目录与eecfe4b3逐文件一致，frontend-custom冻结清单未变。
迁移桥接区分真正的上游独立账本与丢失个人账本的歧义状态；存在个人表但缺个人账本时拒绝猜测。
保留任务冷却期并结合官方失败次数排序、批量领取与后续调度；中断任务使用独立清理上下文释放领取并保留后续恢复，SQLite聚合时间显式解析。
PaddleOCR同时保留原件总页数/缺页校验与新增SourceBlocks；图谱后台预算12288、Wiki预算12288和文档任务无限超时保留。
旧list_knowledge_chunks的image_ocr证据读取迁入read_document；旧execute_skill_script缺参数提示迁入shell_exec。原件复制适配新的资源目录登记，下载和预览继续private,no-store；引用证据与截断提示同时保留。
沿用定制提交0fe3fa592的GitHub元数据停跟踪规则，不重新启用fork上的Actions。官方README恢复为当前主线说明，原部署README完整保存在legacy-deployment.md并标注历史版本。

测试环境说明：供应商协议及沙箱配置测试使用假客户端，但构造期间会解析公网默认URL。服务器DNS不稳定，测试容器对明确的模拟服务域名使用固定公网占位映射，并使用cgo DNS。此映射只用于单元测试，不能作为真实网络、模型或应用验收证据，也不用于候选应用。

阶段3正式merge提交：a00419105467f514c5999a3bb34a53b638734e01；双亲96bfe3d32和eecfe4b3，作者/提交者均ljc。官方目标已成为集成分支祖先，生产工作区仍在a418adc45。
构建环境已从固定Go镜像摘要及Node/Rust工具链重建，APT/Python/npm/Cargo的默认源均在Dockerfile持久配置。保留源码缓存、Cargo目标卷及npm缓存。官方前端镜像已构建；后端原生镜像通过schema-only空库状态及BrowserSkill0.3.1启动检查，接下来进行隔离应用验收。

## 当前可执行说明

完整测试、构建、隔离启动、中断自检及恢复命令见[runbook.md](runbook.md)，当前真实回归状态见[acceptance.md](acceptance.md)。历史阶段记录按当时状态保留；后续获批原件PDF适配是官方UI树的受控例外，见policy.json。

### 2026-10-20 运行回归补充

PDF结构化配置修复已完成失败自动重试→真实OCR成功重解析和两页扫描件复验。
实体图谱、来源隔离、viewer权限与成员撤权检查通过；产物对应关系见artifact-manifest.json。
定制前端模块范围、旧locale失败及联合浏览器验收仍未完成，维护分支和fork尚未推进。
