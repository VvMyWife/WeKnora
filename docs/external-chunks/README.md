# 外部分块导入协议 v1

服务器源码为唯一交付：`/home/ubuntu/RAG`（`codex/week-04`）、`/home/ubuntu/RAG_Weknora` 与其独立 `frontend-custom/`（均为 `frontend-custom` 分支）。不新增 MD 分块入口，不覆盖已有知识或自动切换旧 RAG 活动索引。

## 固定协议

包是 ZIP，协议标识 `weknora-chunks-v1`。UTF-8，无 BOM。不得包含绝对路径、反斜线、`.`、`..`、符号链接、重复 ZIP 条目、未申报条目或加密条目。ZIP 文件名仅为建议，内容中的协议标识才是依据。

根目录必有 `manifest.json`、`documents.jsonl`、`chunks.jsonl`、`source_units.jsonl`，附件仅可在 `files/` 下。所有行是 JSON 对象，每行一条，不允许空 ID、重复 ID、无归属的 chunk/source unit。相对目录以 `/` 分隔。

`manifest.json`：

```json
{
  "format": "weknora-chunks-v1",
  "schema_sha256": "共享 manifest.schema.json 原始字节 SHA256",
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
8. 固定镜像、记录三个仓库提交和回退路径后上线。保留原版前端，不自动改变正式库暂停/运行状态。

每阶段测试通过后执行 `git diff --check`，核对暂存清单再提交；原件、数据、日志、ZIP、构建产物不入库。未完成的阶段不得表述为已上线。

## 手动运行与测试

### 1. 原 RAG：列出版本、生成新保护版本（可选）、导出

```bash
cd /home/ubuntu/RAG
.venv/bin/python -m pytest tests/test_chunker.py tests/test_table_protection.py tests/test_weknora_export.py -q
./scripts/ragctl.sh export-weknora --list
./scripts/ragctl.sh export-weknora --dataset week1-v1 --generation active --check-only
./scripts/ragctl.sh export-weknora --dataset week1-v1 --generation active
# 只有需要新表格保护版时执行；会生成向量，但不 promote：
./scripts/ragctl.sh generation build table-safe-v1 --dataset week1-v1 --strategy structure --context off
./scripts/ragctl.sh export-weknora --dataset week1-v1 --generation table-safe-v1
```

旧 generation 导出原样，不会自动修复旧 chunk。默认 ZIP 包含原件、MD；缺失原件默认失败，明确加 `--without-originals` 才省略。结果在 `runtime/exports/weknora/<时间>/`，异常原因在 `errors.jsonl`。仅输出单一数据集、单一完整版本；读取数据库使用只读一致性快照。ZIP 不覆盖，先完整校验再发布。

双方 Schema 字节必须一致：

```bash
sha256sum /home/ubuntu/RAG/contracts/weknora/manifest.schema.json
sha256sum /home/ubuntu/RAG_Weknora/internal/externalchunks/manifest.schema.json
sha256sum /home/ubuntu/RAG_Weknora/docs/external-chunks/manifest.schema.json
```

### 2. WeKnora：测试

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited sh -c 'go test ./internal/externalchunks ./internal/types ./internal/handler ./internal/router ./internal/database -count=1 && go test ./internal/application/service -run External -count=1'
/home/ubuntu/RAG/runtime/eval-venv/bin/python -m pytest adapters/bge_m3/test_input_limits.py -q
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm run type-check && npm test -- src/views/knowledge/externalChunks.test.ts src/views/knowledge/archiveMapping.test.ts src/views/knowledge/processingControl.test.ts'
```

跨语言真实 ZIP 测试（替换实际包路径，不能使用旧草案 Schema 的测试包）：

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -v /home/ubuntu/RAG/runtime/exports/weknora/contract-v1-20261016/weknora-chunks-v1.zip:/fixture/export.zip:ro -e EXTERNAL_CHUNKS_TEST_ZIP=/fixture/export.zip -w /src weknora-builder:pdf-unlimited go test ./internal/externalchunks -count=1 -v
```

### 3. 网页导入和验收

1. 新建独立样例知识库，配置可用 embedding / 摘要模型；按需开启 Wiki、图谱。不要在已有正式库做首次导入试验。
2. 标题旁“导入外部分块”选择导出的 ZIP。上传完成后先查看预检结果，不会立即调用模型。预检校验每个成员原始字节、归属、数量、顺序、来源锚点和协议校验值。
3. “确认导入”以最多两份文档并行保存附件并提交后台处理。停止提交只停止后续发送；后台暂停使用全库进度面板。
4. 刷新或重新打开会恢复历史批次，不必重新选择 ZIP。已保存的分块用于所有索引重试与“重解析”，不交给内置分块器。
5. 打开文档，核对 chunk 原文和顺序；表格 chunk 可展开“表格结构片段”。附件原件可预览、下载，并保留 MD 入口。
6. 下载 ID 对照表，确认旧 document/chunk ID 对应新的 knowledge/chunk ID。相同包重复提交不会新增版本；同一来源 ID 的内容或解析附件变化会拒绝，不覆盖。
7. 超长不可拆表格必须出现明确失败，保留来源 payload，不静默截断；修正模型配置后显式重试。预检不保证所有后处理模型都能处理无限长输入。

### 4. 接口、额度和日志

接口前缀：`/api/v1/knowledge-bases/:id/knowledge/external-chunks/batches`。

- `POST`：multipart `file`，仅一个 ZIP，最多 4 GiB；返回预检批次与文档。
- `GET`：最近 100 个批次。
- `GET /:batch_id`：持久化文档状态、阶段、错误及当前处理状态。
- `POST /:batch_id/documents/:record_id/start`：执行/显式重试单份文档。
- `GET /:batch_id/id-map`：下载外部 ID 对照 JSON。

写入要求知识库所有者或管理员；读取要求对应库访问权限。所有查询和稳定 ID 均包含租户、知识库作用域。ZIP 不解压到任意路径；拒绝路径穿越、重复项、链接、加密、未申报成员及异常压缩比。展开总量上限 32 GiB、单 JSONL 上限 512 MiB、单行上限 64 MiB。

私有 ZIP 保存在应用持久化卷 `/data/files/external-imports/<tenant>/<kb>/`；不把这个物理路径返回浏览器。预检会占用持久化空间：保留 ZIP、来源 payload 和解析 MD 预留额度；原件存储通过档案映射另行计费。导入历史不会因删除知识条目而自动删除，防止破坏审计；因此这部分保留空间不会随文档删除立即释放。不得手工删包而保留数据库指针。

后台任务沿用现有全库进度和 Trace；导入事件在主机：

```bash
cd /home/ubuntu/RAG_Weknora
find data/processing-logs -path '*/external-chunks/*/events.jsonl' -print
docker logs --since 30m WeKnora-app 2>&1 | grep -E 'external|EXTERNAL_INDEX_FAILED'
tail -50 logs/bge-m3.log
```

`preparing` 的上传因进程退出中断时可在 30 分钟租约过期后再次提交。已经入队的解析/后处理任务由现有持久化队列与重试机制管理，不将其误认作尚未提交。

### 5. 上线状态和边界

部署与回退命令（会重启应用，应先确认允许中断当前任务）：

```bash
cd /home/ubuntu/RAG_Weknora
bash deploy/external-chunks/release.sh build
./frontendctl build custom
sudo systemctl restart weknora-bge-m3.service
/home/ubuntu/RAG/runtime/eval-venv/bin/python deploy/external-chunks/check_embedding.py
bash deploy/external-chunks/release.sh deploy
./frontendctl switch custom
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
# 后端回退；新增迁移和导入数据保留：
bash deploy/external-chunks/release.sh rollback
# custom 前端回退到本轮之前的镜像：
docker tag archive-intelligence-ui:before-external-chunks-20261016 archive-intelligence-ui:0.1.0
./frontendctl switch custom
```

可复现的独立验收包（不访问数据库、不调用模型）：

```bash
cd /home/ubuntu/RAG
PYTHONPATH=. .venv/bin/python scripts/weknora_acceptance_package.py runtime/exports/weknora/my-acceptance
```

v1 不支持直接克隆外部分块文档到另一知识库；请向目标库导入同一 ZIP，以生成正确作用域的 ID 映射。附件 TXT 用于隔离验收；实际导出保留各文档原 PDF/图片文件，不将 MD 冒充原件。

2026-10-16 已按用户允许重启的确认上线。应用镜像 `weknora-app:external-chunks-20261016`，custom 前端入口仍为 18180；原版 frontend 未改。首次启动发现重复服务装饰器注册，已先回退旧镜像，再修复并重新部署，最终应用健康，数据库为 93。正式库暂停/运行状态不由导入流程自动改变。

独立实机验收知识库：`4061d82c-401a-4f97-8b0b-96908030a705`。网页上传的样例包含两页 JSON 来源单元、多行表头和 rowspan 表格，共 3 个外部分块。数据库逐块核对 content/source_content、顺序、稳定 UUID、来源锚点均一致；向量化、摘要和全流程完成。再次“重解析”仍使用同一 payload，3 个分块与 ID 未变。原件预览显示独立 TXT，而不是 MD，分块界面可展开保留合并单元格的表格。此次样例未开启 Wiki/图谱：这两项复用现有后处理队列，不能把样例完成表述为新批次 Wiki/图谱全部完成。

```bash
cd /home/ubuntu/RAG_Weknora
python3 deploy/external-chunks/audit_import.py 45227aa4-6872-5f1a-b9ff-baacc2dd8378 \
 /home/ubuntu/RAG/runtime/exports/weknora/live-table-20261016/export/weknora-chunks-v1.zip
```

本次新增 PostgreSQL 93 / SQLite 14 迁移；回退保留新增表、导入 ZIP、原件和日志，不运行 down migration。embedding 的 `/v1/token-counts` 返回实际 tokenizer 长度，`/v1/embeddings` 拒绝超限，不依赖粗略字符估计。未提供可验证限额接口的 embedding provider 会被明确拒绝，不能声称通用 provider 已验收。
