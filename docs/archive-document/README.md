# 完整档案显示、Wiki／图谱和失败重试

正式入口：`http://10.20.0.15:18182`。
权威代码：`/home/ubuntu/RAG_Weknora-upstream-20261020`，分支 `ljc/archive-document-workflow`。
原 RAG 的导出命令详见 `/home/ubuntu/RAG/docs/archive-document/README.md`。

## 使用和手工验收

1. 用原 RAG 的 `ingest --merge-folders --export-weknora` 生成 ZIP。每个直接子文件夹是一份档案。不要将页级旧库和档案新库混合导入。
2. 在 18182 新建知识库，配置 embedding、摘要模型，按需要启用 Wiki 和图谱。选择“导入外部分块”上传 ZIP，预检通过后开始。
3. 文档列表按原相对目录展示，一份档案一条记录，显示页数、内部切块数和实际附件大小。点击打开完整 PDF，“查看分块”展示绑定此档案的切块，并可定位 PDF 页。
4. Wiki 和图谱顶部入口默认列出档案，选择后展示当前档案。Wiki 先显示该档案概览；主题可跨档案共享，但返回的证据只包含当前档案切块。概览没有精确 chunk 引用时不编造页码。
5. 图谱查询同时限制 KB 和文档 namespace，关系通过实际实体 ID 连接。当前图中优先显示关联最多的 40 个实体及它们之间的关系；实体下展开证据可回到完整档案。此图是有界查询结果，不表示穷尽所有实体。
6. 可切换“全库主题视图”继续使用原 Wiki 浏览器。原有检索和 LLM 分批依然以 chunk 为内部工作单元，档案是用户视图和来源绑定单位。
7. 全库失败列表可以单条重试，也可以重试当前筛选阶段下全部失败档案；文档详情有“处理与重试”。服务端只针对最新处理 attempt 的失败摘要、图谱或 Wiki 补做；真正解析失败仍需重解析。已成功阶段保留，连续点击不产生重复任务。
8. 暂停知识库后点击重试：持久化记录保持 ready，不执行模型任务；恢复后再执行。自动次数耗尽后可人工重新提交一轮。批量接口同步接收各档案的重试请求，实际任务在后台持久队列执行，进度看全库状态；没有独立批量 job_id。

## 测试与构建

已有依赖使用持久缓存，构建保留 anydoc 解析能力。不下载新基础镜像；已有基础镜像先保留为回退标签。

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
docker run --rm --entrypoint sh \
  -v "$PWD:/app" -w /app \
  -v "$PWD/.go-modcache:/go/pkg/mod" \
  -v weknora-go-build-cache:/root/.cache/go-build \
  -e GOPROXY=https://goproxy.cn,direct \
  weknora-builder:upstream-20261020 -c \
  'go test ./internal/application/service ./internal/handler ./internal/router ./internal/application/repository/retriever/neo4j && make build-prod GO_BUILD_TAGS=anydoc && cp WeKnora runtime/archive-document/WeKnora'

docker run --rm --entrypoint sh \
  -v "$PWD/frontend-custom:/app" -w /app \
  -v ljc-upstream-custom-node-modules:/app/node_modules \
  -e npm_config_registry=https://registry.npmmirror.com \
  weknora-frontend-builder:upstream-20261020 -c \
  'npm run type-check && npm test -- src/views/knowledge/processingOverview.test.ts src/views/knowledge/processingControl.test.ts src/views/knowledge/wikiStatusRefresh.test.ts src/utils/knowledgeTrace.test.ts && npm run build'

docker build -f deploy/archive-document/Dockerfile.app \
  --build-arg SOURCE_REVISION="$(git rev-parse HEAD)" \
  -t weknora-app:archive-document-20261026 .
docker build -f deploy/archive-document/Dockerfile.frontend \
  --build-arg SOURCE_REVISION="$(git rev-parse HEAD)" \
  -t weknora-ui-custom:archive-document-20261026 .
```

Go 测试覆盖持久化摘要重试去重、图谱实体端点身份，以及现有服务/路由回归。前端类型检查、相关测试和实际浏览器验收需同时通过。

## 发布、状态自检及回退

本次用户已明确授权验收后替换并重启 app 和 18182 前端。下面脚本只适用于本次固定基线，并要求 app/UI 镜像的源码 revision 一致。`prepare` 已执行时不能重放；中断后先运行 `status` 并核对容器。

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 deploy/archive-document/release.py status
# 前端独立修复、验收且获授权后，只刷新前端，后端保持运行：
# python3 deploy/archive-document/release.py refresh-frontend
# 新一次交付尚未 prepare 时：
# python3 deploy/archive-document/release.py prepare
# 获得该次发布授权且验收通过后：
# python3 deploy/archive-document/release.py apply

curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18182/
docker inspect -f '{{.Name}} {{.Image}} {{.State.Status}}' \
  WeKnora-app WeKnora-frontend-custom-preview

# 回退本次发布，恢复已保存的两张旧镜像和 Compose：
python3 deploy/archive-document/release.py rollback
```

脚本不重启数据库、docreader、模型、18180，也不执行 schema migration；保存旧 Compose 和镜像 ID，错误时自动恢复。发布前发现无关容器 kabel 已自动重启 785 次：只校验其容器/镜像身份，不把该既有重启循环当成本次发布导致；其他保护容器还核对启动时间。本次不修复或操作 kabel。

证据在 `/home/ubuntu/RAG/runtime/archive-document-acceptance-20261026/`。`release/state.json` 记录发布前后镜像及保护容器；其中 Compose 和身份数据保存在权限 0700 的目录，不提交 Git。

## 实际验收记录

隔离环境使用单独 PostgreSQL 数据库与 Redis 队列，防止消费正式队列；共享模型与文件服务按独立 KB ID 写入。

- 001 样本：1 档案、45 页、95 正文切块；PDF 与 ZIP 字节完全一致，正文与来源锚点逐项相同。
- 图谱 95 段成功；1 篇档案概览、19 篇关联 Wiki。模型生成主题可以共享，不会伪称它们互相独立。
- 摘要失败夹具：暂停时 first accepted=1、duplicate=0，恢复后 recovered；正文切块/索引摘要指纹与 95 段成功图谱保持不变。
- 图谱失败夹具：自动预算耗尽后批量重新提交 accepted=1，重复=0；暂停时 attempts=0，恢复后 recovered。
- 原 RAG 全量 92 测试通过；完整目录 102 档案、2634 页、5998 切块。
- 正式 102 份档案逐项核验 PDF 字节、5998 正文切块和 OCR 来源锚点全部一致；102 份均已有一篇档案概览。
- PDF 页定位曾使用预览器不接受的标记，已修正并增加扫描 PDF 第 2 页定位回归；`mapping=exact` 仅表示归档页序确定，不宣称 OCR 框达到精确文字级映射。切换档案会清除上份档案的位置。

正式新库 ID：`a8589430-613a-4018-814c-0ddbab1117c6`。
发布的后端源码是 `f45137d3f`，前端修复源码是 `e5df8a440`；只有前端进行了后续独立刷新。发布脚本保存两者的镜像关系与原始回退版本。
截至本次交付记录，62 份全流程完成、40 份仍 finalizing，17 条图谱阶段失败记录仍在自动重试流程内，尚无当前新库的最终 dead-letter；这不是全部图谱完成。后续实际状态以页面实时数据为准。

正式新库使用“编目排序样本集-档案版”；旧 2634 文档库保留。正式处理进度以 `production-state.json` 和页面实时状态为准，不能把 ZIP 已导入或向量化结束写成 Wiki/图谱全部结束。
