# WeKnora upstream integration 2026-10-29

Custom baseline: 1f8564d64e7feb78afd5ea202f78205875c75e2f.
Fixed official target: e8e0d2ed4bd660c46313884d2f2b3b243d7c318d.
201 upstream commits and 702 changed files; no production deployment.
Production remains official128 + personal4. Never checkout/reset the production-mounted 20261020 worktree.

Stages: freeze; formal merge; custom compatibility; frontend adaptation; isolated runtime/migration acceptance; verified fork push.

Backups: /home/ubuntu/weknora-fork-migration/20261029-upstream/pre-sync.bundle and production-baseline.json.
Interrupted work must inspect Git, processes, container identities and test logs before resuming. Tests/builds may resume, migrations and deployment must never replay blindly.

## 已完成的代码阶段

- f61a5f2e0：冻结201个提交及702个文件。
- f8be70abe：正式merge，保留PDF完整页校验、资源限制拒绝和引用诊断；纯文本ref示例保留，真实来源引用仍需证据绑定。
- a7b2502f3：官方157个前端文件增量适配至frontend-custom，保留原件预览、档案Wiki/图谱和失败重试。
- 定制前端全量2400项：2399通过、1条件跳过；类型检查及Vite构建通过。新增官方批量重解析测试补充定制依赖夹具，没有删除或跳过断言。
- DocReader493项：12条件跳过，其余通过；后端、Asynq及原生AnyDoc测试和编译通过。独立TLS Docker守护进程完成两轮真实创建/读写/执行/超时/删除。
- 静态SSR测试禁用无用文件监听，避免共享宿主inotify耗尽；不修改生产系统限制。
