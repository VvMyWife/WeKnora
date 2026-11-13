# 2026-11-13 WeKnora 主线集成

权威目录为 `/home/ubuntu/RAG_Weknora`。本批合入官方 `cbdcebabf48c5ee2d6868b370326aa8341703f62`，共 209 个提交、1,110 个文件。正式 merge 为 `dffde6e82`，前端适配为 `b440c52fc`，Wiki 中断续跑修复为 `f0693c65c`。保留 fork 已有的镜像发布工作流设置。

| 产物 | 源码 | 镜像 |
|---|---|---|
| 后端 | `f0693c65c` | `weknora-app:ljc-candidate-f0693c65c-bsk041`，ID `sha256:20025948c9bbfa8491cb8d1322a8c31b2525b6880de4df44b59ceed76500b1f9` |
| DocReader | `b440c52fc` | `weknora-docreader:ljc-candidate-b440c52fc`，ID `sha256:7887a23a38a75a179adb8070c8122f147c80b07a0c6be55df79a292da573ea66f` |
| 18182 候选 | `b440c52fc` | `weknora-ui-custom:ljc-candidate-b440c52fc`，ID `sha256:4b1d5e95102b1236a0eae240f98afedadf228ebe872f16e35c7eb86d7e83a6677` |
| 正式后端 | `61c41d0ee` | 当前生产仍运行此源码，账本官方 134＋个人 5 |

后端沿用已验收运行基础层，BrowserSkill 守护程序和扩展使用与官方清单同源的 0.4.1 官方发布资产。DocReader 复用上一批依赖层，本批 `pyproject.toml` 和 `uv.lock` 无变化。

## 验收记录

- Go、Asynq、AnyDoc 通过。定制前端 2,739 项通过、1 项跳过，类型检查和生产构建通过。DocReader 运行 566 项，12 项跳过。
- PostgreSQL 从生产库只读取得数据流，在独立库验证 134＋5 升至 136＋5，重复执行成功，12 张业务表指纹一致。SQLite 51＋4 升至 52＋4，重复执行成功。凭据库创建、查询、存储和日志脱敏及删除均通过。
- 真实文件入库、CSV／Excel／普通 PPTX／扫描 PDF 解析、扫描 PDF OCR 分块、混合检索、CUDA 重排、外部分块去重、原件双版本、Wiki／图谱范围隔离、重试去重、暂停／恢复／取消、跨空间权限及匿名同步／流式问答均通过。匿名同步和流式各 1 条引用明细。浏览器实测登录及知识库页面，控制台无错误。
- 待专项验收包括浏览器内上传／聊天／临时失败后的引用预览及实际 PDF 页码交互、PPTX 矢量媒体实文件、BrowserSkill 会话、独立沙箱和数据库失败恢复演练。引用预览失败恢复和 PDF 页码已有测试覆盖。

## 复现与回退

以下在服务器执行。运行材料和原始日志位于 Git 忽略的私有目录。重新构建前确认工作区包含上述产物源码，并核对运行配置。既有候选运行数据库和存储仅供本批验收。

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/upstream-integration/20261113
git status -sb
git merge-base --is-ancestor cbdcebab HEAD
git diff --exit-code f0693c65c -- . ':!docs/upstream-sync-20261113'
bash "$p/backend-full.sh"
bash "$p/frontend-full.sh"
bash "$p/docreader-test.sh"
bash "$p/candidate-backend.sh" "$(git rev-parse f0693c65c)"
docker build --pull=false -t weknora-app:ljc-candidate-f0693c65c-bsk041 "$p/bsk-image"
bash "$p/candidate-docreader.sh" "$(git rev-parse b440c52fc)"
bash "$p/candidate-ui.sh" "$(git rev-parse b440c52fc)"
```

启动前检查 28281／28282 空闲，Compose 仅绑定回环地址。用同一批隔离容器和测试数据库重新运行验收，验收后停止。查看 `*.status` 和对应日志，已完成的迁移或恢复阶段先检查现有账本和容器，再决定从哪个阶段继续。

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/upstream-integration/20261113
docker start ljc-20261113-proof-pg
docker compose -f "$p/compose.acceptance.json" config --quiet
docker compose -f "$p/compose.acceptance.json" up -d --pull never
curl -fsS http://127.0.0.1:28281/health
python3 "$p/acceptance-realfile.py"
python3 "$p/acceptance-gpu-rerank.py"
python3 "$p/anon-smoke.py"
docker compose -f "$p/compose.acceptance.json" stop
docker stop ljc-20261113-proof-pg
```

正式部署走独立维护窗口。沿用 [生产运维流程](../production-workflow/README.md) 核对镜像和配置。发布时若升级数据库，回退须使用切换前数据库或恢复副本，切回旧镜像前要核对切换后新增的业务写入。当前线上仍以 18182 为正式入口。
