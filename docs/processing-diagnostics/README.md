# WeKnora 上传与处理故障定位

适用服务器：`10.20.0.15`，代码目录：`/home/ubuntu/RAG_Weknora`。

本次修改不改变“同知识库、同文件类型、同内容哈希”的判重规则。重复上传仍返回 HTTP 409，但响应会指出已存在文档的路径；应用日志会同时记下请求路径、重复对象路径、文档 ID 和知识库 ID。解析、摘要、Wiki、图谱等处理阶段调用 `FailSpan` 时，会输出一条带文档 ID、尝试次数、阶段、错误码和原因的 `[KnowledgeProcessFailure]` 日志。完整错误明细仍保存在 PostgreSQL 的 `knowledge_processing_spans` 表。

Wiki 模型输出请求上限从 32768 降为 8192，以给 32768-token 上下文中的输入提示留空间。这不会自动恢复以前进入死信队列的任务，也不保证超长提示或超长 JSON 输出一定成功。

## 自己验证

在服务器运行：

```bash
cd /home/ubuntu/RAG_Weknora
git diff --check
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go test ./internal/handler ./internal/application/service ./internal/application/repository
```

构建与上线（确认当前批处理允许短暂重启后执行）：

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go build -buildvcs=false -o runtime/WeKnora.diagnostics-20261014 ./cmd/server
BUILDKIT_PROGRESS=plain docker build -f deploy/processing-diagnostics/Dockerfile.runtime --build-arg BINARY_FILE=WeKnora.diagnostics-20261014 -t weknora-app:diagnostics-20261014 runtime
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml config --services
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml up -d --no-deps app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
```

旧运行容器的二进制备份在 `runtime/WeKnora.pre-diagnostics-20261014`，回退镜像是 `weknora-app:pre-diagnostics-20261014`。如需回退，先将 `deploy/pdf-unlimited/compose.yaml` 中的 app 镜像改为该回退标签，再执行同一条 `docker compose ... up -d --no-deps app`。不要删除回退镜像或二进制，直至新版本验证完成。重启可能使正在执行的 Asynq 子任务中断并重新入队；先观察任务数是否继续推进。

重传一份内容与已存在文档相同、但文件名或文件夹不同的 `.md`。应收到 HTTP 409；响应 `code` 为 `duplicate_file`，`message` 包含已存在文档路径。按文档 ID 查原始处理轨迹：

```bash
docker logs --since 30m WeKnora-app 2>&1 | grep -E '\[KnowledgeUploadFailure\]|\[KnowledgeProcessFailure\]'
docker exec WeKnora-postgres psql -U weknora -d weknora -P pager=off -c "SELECT attempt,name,status,error_code,error_message,error_detail,finished_at FROM knowledge_processing_spans WHERE knowledge_id = '替换为文档ID' AND status = 'failed' ORDER BY id DESC LIMIT 30;"
```

后台日志是 Docker 容器日志，不另存用户文档内容或模型提示。阶段失败日志最多保留 1024 字节的原因摘要；数据库轨迹中保留较完整的错误明细。若日志里的 `knowledge_id` 不直观，可查询路径：

```bash
docker exec WeKnora-postgres psql -U weknora -d weknora -P pager=off -c "SELECT id,folder_path,file_name,parse_status,summary_status,error_message FROM knowledges WHERE id = '替换为文档ID';"
```

本次源目录 `E:\desktop\政策标准库_MD` 有 651 份 Markdown，目标知识库曾登记 634 份。差出的 17 份逐一与库中另外 17 份文档具有相同 MD5 内容哈希；这是当前判重规则的预期结果，不代表文档内容完全丢失，但这些源路径没有单独入库。
