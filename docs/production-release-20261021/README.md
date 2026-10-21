# 18182 正式发布与回退

用户于2026-10-21明确授权完整生产升级：使用正式数据迁移数据库/后端，把28182已验收定制前端发布到18182，验收后停止28182。18182是以后唯一指定日常前端；18180仅保留历史排查用途，不作为新版验收入口。

## Git阶段

1. ljc/release-20261021-18182：冻结身份、发布脚本、任务/配置检查及恢复材料。
2. 最新正式数据备份、全新ParadeDB0.25.11卷恢复、旧94拆分为90+个人4，再升级官方128+个人4，重复up和业务表核对。
3. 原件/OCR/模型/权限配置保持；正式后端和18182切换、真实浏览器验收；停止28182并取消重新绑定端口配置。
4. 记录镜像/源码/数据版本映射，合入ljc/custom并推送fork。

作者和提交者ljc；阶段通过后独立commit，不改写历史。

## 中断自检

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short --branch
python3 scripts/ljc_production_release.py status
docker ps --filter name=WeKnora --format '{{.Names}} {{.Image}} {{.Ports}}'
ps -eo pid,etime,args | grep -E 'ljc_production_release|pg_restore|pg_dump'
tail -30 /home/ubuntu/weknora-fork-migration/20261021-production/release.log
~~~

脚本有进程锁和阶段检查，正在执行或阶段不符时拒绝重放。观察超时不代表失败；先查同一PID、日志及state。不要自动重建数据库或重复迁移。

## 按阶段运行

以下命令仅适用于本次已授权发布。私有compose/env从真实容器生成，不手工复制密码到命令行或Git。stage成功后才运行下一步；历史失败保留。

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 -m py_compile scripts/ljc_production_release.py
python3 scripts/ljc_production_release.py preflight --confirm-production-release
python3 scripts/ljc_production_release.py backup --confirm-production-release
python3 scripts/ljc_production_release.py restore --confirm-production-release
python3 scripts/ljc_production_release.py migrate --confirm-production-release
python3 scripts/ljc_production_release.py cutover --confirm-production-release
python3 scripts/ljc_production_release.py verify --confirm-production-release
~~~

backup从停止前端/后端起进入维护窗口，保存最新数据库、角色、文件、Redis、Neo4j和配置。新数据库容器WeKnora-postgres-20261021使用新卷weknora-production-postgres-20261021；原PG旧卷保持不变。Redis、Neo4j和文件路径继续使用正式资源，不使用候选库/队列/文件。
migrate只运行显式schema-only入口，不启动工作队列；严禁force跳过dirty状态。
cutover使用已验收的不可变镜像ID，保留JWT/加密密钥及正式模型配置，AUTO_MIGRATE=false。

正式浏览器检查完成并保存私有browser-acceptance-results.json后：

~~~bash
python3 scripts/ljc_production_release.py retire --confirm-production-release
curl --fail http://10.20.0.15:18081/health
curl --fail -I http://10.20.0.15:18182/
docker inspect -f '{{.State.Status}} {{.HostConfig.RestartPolicy.Name}}' ljc-upstream-acceptance-custom-ui
ss -ltn | grep ':28182 ' || true
~~~

私有目录/home/ubuntu/weknora-fork-migration/20261021-production权限700。不要提交containers-private.json、compose.production.json、migration.env、数据库、token、日志或文件数据。

## 常规启动和重启

完成发布后只使用新的生产compose。旧运行目录仍为历史源码，不要从旧docker-compose启动旧94账本应用：

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json config --quiet
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json up -d postgres docreader app frontend
docker exec WeKnora-frontend-custom-preview nginx -t
~~~

后端重新创建导致IP变化时重载18182代理：

~~~bash
docker exec WeKnora-frontend-custom-preview nginx -s reload
~~~

## 回退原则与步骤

数据库已升级，不能只切回旧镜像。旧PG容器WeKnora-postgres和旧卷保留在94；旧应用/Docreader/定制前端改名为对应的-baseline-20261021且restart=no。
回退前先停止入口和新后端，保存新库和文件/Redis/Neo4j最新状态，以保留上线后的写入供核对。没有完成维护窗口内恢复及写入差异核对，不启动旧应用。

1. 停止新前端、新app、新docreader，阻止新写入；pg_dump新数据库到独立post-cutover.dump，备份当前文件和队列/图谱状态。
2. 停止生产Redis和Neo4j。从本次backup-sha256.json校验通过的redis.tar、neo4j.tar、files.tar恢复匹配的正式卷；不要使用候选卷。恢复前另存现有卷，保留上线后新写入。
3. 新数据库停止；启动旧WeKnora-postgres（旧94原卷从冻结以来未升级）。若旧卷损坏，从本次database.dump及globals.sql恢复到相同旧版本PG0.22.2新卷，再检查94 clean及保护表数量。
4. 把当前三个新应用容器改名为failed-release-时间；将-baseline-20261021三个容器恢复原名。启动Redis/Neo4j、旧Docreader和旧app，健康后启动旧前端18182，重载历史18180代理。
5. 检查正式数据、旧API和原件功能，再开放入口。保留新版库/上线后备份供人工核对，不清理镜像和卷。

完整原始镜像备份仍位于20261020-upstream/rollback-images.tar.gz；本次正式最新数据备份以20261021-production为准。

## 正式GPU重排服务

保留原正式模型地址http://10.20.0.15:18083，使用物理GPU1及已有缓存权重，不使用CPU。服务weknora-bge-rerank.service由systemd开机启动、异常退出后重启；原候选18085临时GPU进程已停止，未移动其他GPU模型。

~~~bash
systemctl is-active weknora-bge-rerank.service
systemctl is-enabled weknora-bge-rerank.service
curl --fail http://10.20.0.15:18083/health
nvidia-smi
journalctl -u weknora-bge-rerank.service -n 30 --no-pager
~~~

首次安装使用sudo安装deploy/production-20261021/weknora-bge-rerank.service到/etc/systemd/system，daemon-reload并enable；确认同GPU候选旧进程退出后才start。不得同时启动旧候选GPU脚本与正式服务。

## 本轮迁移结果

最新正式备份已在新卷恢复并完成旧94→官方90+个人4→官方128+个人4，重复up和保护表数量核对通过。首次初始化临时PostgreSQL误报就绪导致恢复连接中断；修复为确认PID1是最终postgres后再恢复，保留原失败日志，只重建未导入业务数据的新目标库。旧94库和原数据卷未迁移。详情见migration-results.json。
