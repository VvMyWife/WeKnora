# 2026-11-13 WeKnora 主线集成

权威目录为 `/home/ubuntu/RAG_Weknora`。本批合入官方 `cbdcebabf48c5ee2d6868b370326aa8341703f62`，共 209 个提交、1,110 个文件。正式 merge 为 `dffde6e82`，前端适配为 `b440c52fc`，Wiki 中断续跑修复为 `f0693c65c`。保留 fork 已有的镜像发布工作流设置。

| 产物 | 源码 | 镜像 |
|---|---|---|
| 后端 | `f0693c65c` | `weknora-app:ljc-candidate-f0693c65c-bsk041`，ID `sha256:20025948c9bbfa8491cb8d1322a8c31b2525b6880de4df44b59ceed76500b1f9` |
| DocReader | `b440c52fc` | `weknora-docreader:ljc-candidate-b440c52fc`，ID `sha256:7887a23a38a75a179adb8070c8122f147c80b07a0c6be55df79a292da573ea6f` |
| 18182 候选 | `b440c52fc` | `weknora-ui-custom:ljc-candidate-b440c52fc`，ID `sha256:4b1d5e95102b1236a0eae240f98afedadf228ebe872f16e35c7eb86d7e83a667` |
| 升级前正式后端 | `61c41d0ee` | 原库保留，账本官方 134＋个人 5 |

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

后续正式部署在维护窗口执行。沿用 [生产运维流程](../production-workflow/README.md) 核对镜像和配置。发布时若升级数据库，回退须使用切换前数据库或恢复副本，切回旧镜像前要核对切换后新增的业务写入。当前线上仍以 18182 为正式入口。

## 18182 正式上线

北京时间 2026-11-13 14:23 完成切换。正式 18182 使用上表中的新版后端、DocReader 和定制前端，生产清单已更新到主目录 `runtime/production/compose.json`。运行库为 `weknora_release_20261113`，官方 136＋个人 5，均无 dirty 标记。正式数据流式恢复后升级，12 张业务表内容指纹一致，11 个检索索引有效。原库 `weknora_release_20261112` 保留。

18182 实际浏览器加载了新版资源并显示「档案智查」，控制台无错误。正式匿名问答通过，1 条引用明细可查。凭据库路由与表已就绪，匿名访问返回 401。登录后的管理界面仍以隔离验收记录为准。

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/production/release-20261113
cat "$p/state.json" "$p/database-proof.json" "$p/postflight.json"
docker compose -p ljc-weknora-production -f runtime/production/compose.json config --quiet
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18182/
```

回退清单为 `runtime/production/release-20261113/compose.before.json`。维护窗口内先暂停写入，保留并核对新版库的新业务记录，再使用原库和原镜像回退。确认数据处理方式后依次执行：

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/production/release-20261113
docker compose -p ljc-weknora-production -f "$p/compose.before.json" config --quiet
for service in docreader app frontend; do
  docker compose -p ljc-weknora-production -f "$p/compose.before.json" up -d --no-deps --pull never "$service" || break
done
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18182/
```

回退后复测问答和引用，并把主目录生产清单同步到实际回退版本。
