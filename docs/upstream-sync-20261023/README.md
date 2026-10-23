# 官方主线同步：2026-10-23

固定官方目标：bc1ade36b933ff7689f1659d670cfd5683c174c5。
定制基线：2b7d874de33ca562405829eb3f58cd721e2c67f8。
权威Git仓库：/home/ubuntu/RAG_Weknora；当前维护工作区：/home/ubuntu/RAG_Weknora-upstream-20261020。
本轮隔离工作区：/home/ubuntu/RAG_Weknora-upstream-20261023，分支ljc/integrate-20261023-upstream-main。

## 分阶段实施

1. 冻结：检查工作区干净、保存Git bundle及运行镜像基线，提交本说明。
2. 正式merge固定官方SHA，处理技术冲突，保留所有已实现定制（原件映射/外部分块/引用/文档图谱/任务暂停和无限超时/Asynq补丁/模型配置）。
3. 将官方frontend增量三方移植到frontend-custom，保留品牌布局、原件PDF预览和脚注；后端/数据迁移保持官方与个人账本分离。每组通过测试后独立提交。
4. 对后端、两套前端、DocReader、迁移和核心定制回归，构建不可变候选产物；通过后快进ljc/custom并推送fork。生产切换另记录本轮涉及的迁移、运行任务、备份和验收。

本轮不追逐执行中出现的新官方提交；下次从本次SHA继续。此前旧勾选清单已被全量同步指令取代。日常前端唯一18182；禁止重建28182宿主端口，18180保留。

## 中断自检与手动核对

~~~bash
cd /home/ubuntu/RAG_Weknora-upstream-20261023
git status --short --branch
git rev-parse HEAD
git rev-parse -q --verify MERGE_HEAD || true
git diff --name-only --diff-filter=U
git worktree list
docker ps --filter label=ljc.operation=upstream-20261023
tail -40 runtime/upstream-20261023/*.log 2>/dev/null || true
~~~

先检查同一进程/容器/日志与结果，再恢复；测试/构建可重跑，迁移和发布不可盲目重放。不得以健康检查或局部测试代替定制功能验收。

## 冻结材料与回退

/home/ubuntu/weknora-fork-migration/20261023-upstream/pre-sync.bundle保存维护分支完整历史；runtime-images.txt记录本轮开始时镜像。恢复材料不进Git。此阶段未修改数据库或运行容器。

~~~bash
git bundle verify /home/ubuntu/weknora-fork-migration/20261023-upstream/pre-sync.bundle
# 单独新目录查看旧代码，不重置有数据或运行挂载的工作区：
git worktree add --detach /home/ubuntu/RAG_Weknora-review-baseline-20261023 2b7d874de33ca562405829eb3f58cd721e2c67f8
~~~

若merge仍在处理中且确认放弃本次未提交集成，可在本轮隔离工作区git merge --abort；不要对生产挂载目录执行reset/clean。若已涉及生产迁移，回退必须恢复匹配数据库，不能只换旧镜像。

后续测试命令、实际SHA、结果和产物映射在各阶段完成时补齐。
