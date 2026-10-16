# 外部分块导入协议 v1

服务器源码为唯一交付：`/home/ubuntu/RAG`（`codex/week-04`）、`/home/ubuntu/RAG_Weknora` 与其独立 `frontend-custom/`（均为 `frontend-custom` 分支）。不新增 MD 分块入口，不覆盖已有知识或自动切换旧 RAG 活动索引。

## 固定协议

包是 ZIP，协议标识 `weknora-chunks-v1`。UTF-8，无 BOM。不得包含绝对路径、反斜线、`.`、`..`、符号链接、重复 ZIP 条目、未申报条目或加密条目。ZIP 文件名仅为建议，内容中的协议标识才是依据。

根目录必有 `manifest.json`、`documents.jsonl`、`chunks.jsonl`、`source_units.jsonl`，附件仅可在 `files/` 下。所有行是 JSON 对象，每行一条，不允许空 ID、重复 ID、无归属的 chunk/source unit。相对目录以 `/` 分隔。

`manifest.json`：

```json
{
  "format": "weknora-chunks-v1",
  "package_id": "UUID",
  "source_system": "archive-rag",
  "dataset_id": "week1-v1",
  "generation_id": "table-safe-v1",
  "producer_commit": "git commit",
  "generation_config": {},
  "counts": {"documents": 1, "chunks": 2, "source_units": 3},
  "files": {"documents.jsonl": {"sha256": "64 lowercase hex", "size": 123}}
}
```

`files` 必须覆盖除 manifest 自身外的全部成员。校验值基于原始字节；包内容身份使用 manifest 原始字节的 SHA256，ID 不是内容校验值。包自身不保存凭据或主机绝对路径。`package_id` 可随机生成；同一来源标识的内容变化不能静默覆盖。

`documents.jsonl` 每条：`id`（旧文档 ID）、`title`、`relative_path`（逻辑 MD 相对路径）、`version`、`source_sha256`、`parsed_file`（附件路径）、`original_file`（附件路径或空字符串）。逻辑目录不能用同名文件模糊匹配。

`chunks.jsonl` 每条：`id`（旧 chunk ID）、`document_id`、`order`（每文档从 0 连续）、`content`（不可改写的证据文本）、`retrieval_text`、`context_text`、`token_count`（估计值）、`metadata`、`source_anchors`。内容完整保留；上下文不得覆盖 content。

`source_units.jsonl` 每条：`id`、`document_id`、`unit_id`、`page_index`（从 0）、`order`、`block_type`、`text`、`bounding_box`、`title_path`、`table`（无表格时 null，包含原始 HTML 和网格）。chunk 锚点 `unit_id` 必须引用本文件对应文档单元；表格片段额外提供 `row_indices`、`header_indices`，原文字符范围不可冒充表格行级精确定位。

## 表格规则

正文仍沿用旧长度与重叠逻辑。表格按完整行和跨行合并组切分；每片重复显式表头（thead/th），不猜测首行。保留片段 HTML、原始行号、rowspan/colspan 和完整来源单元。结构不可靠时保留整表并标记。目标长度为软限制，不是 embedding 的传输上限；真实模型超限必须报错，不允许静默截断。

## 分阶段提交与验收

1. WeKnora：协议、规范和兼容样例。
2. RAG：表格完整行/合并组保护及测试。
3. RAG：只读导出、预检、摘要和异常记录；不重跑 OCR、切块、向量化或模型。
4. WeKnora：安全预检、持久化批次、外部 ID 映射与幂等。
5. WeKnora：导入 chunk 原文直达向量与后处理，恢复只使用已保存分块；原件接现有档案映射。
6. custom：外部分块 ZIP 上传、预检、执行、进度、失败原因和 ID 对照下载。
7. 文档和独立端到端验收。
8. 固定镜像、记录三个仓库提交和回退路径后上线。保留原版前端与正式库暂停状态。

每阶段测试通过后执行 `git diff --check`，核对暂存清单再提交；原件、数据、日志、ZIP、构建产物不入库。未完成的阶段不得表述为已上线。
