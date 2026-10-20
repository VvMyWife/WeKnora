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
