# 档案映射

服务器唯一源码：`/home/ubuntu/RAG_Weknora`。本仓库与 `frontend-custom/` 分别提交。原版 `frontend/` 不修改。原件映射不生成知识条目、不改变 knowledge/chunk ID，不调用模型。

## 分阶段提交

1. 独立数据表、相对路径匹配与版本幂等测试。
2. 批次清单、原件存储、版本管理和日志。
3. 原件预览/下载解析，以及复制、删除生命周期。
4. custom 前端映射管理、刷新恢复、每批并发 2。
5. custom 预览/引用关联原件，保留解析文本入口。
6. 完整验收后固定部署镜像；每阶段先测试，再在对应仓库提交。

## 手动测试命令

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go test ./internal/archivemapping -count=1
git status --short
git diff --check
git log -6 --oneline
git -C frontend-custom status --short
git -C frontend-custom log -6 --oneline
```

新增表迁移只在部署新版后生效；部署前的测试使用独立内存数据库，不修改正在处理的正式库。回退程序镜像时保留映射表与文件，不执行删除数据表的 down migration。

## 使用步骤

1. 将处理完成的 MD 按目录正常导入 WeKnora；本功能不负责外部分块导入。
2. 在目标知识库标题右侧选择“档案映射”，选择原 PDF/图片/Office 文件所在目录。浏览器忽略最外层原件目录名称，其下相对目录必须与 MD 匹配。
3. “MD 匹配根目录”填现有目录树中 MD 的公共前缀，例如 `政策标准库_MD`。若 MD 直接位于根目录则留空。点击“新建匹配预览”。新建清单本身不上传二进制原件。
4. 检查每行相对路径与目标。唯一匹配可批量上传；未匹配或多目标用下拉列表手动选择 MD。不同目录的同名文件不会互相覆盖。已有原件必须明确允许替换，旧版本不会删除。
5. 上传期间可停止发送后续文件，当前两个上传收尾。刷新后历史批次及结果仍在，重新选择原件目录即可继续剩余项。失败原因在对应行显示并导出日志。
6. 文件详情/聊天引用仍使用原 knowledge ID；自动预览和下载优先原件，保留“解析文档”和分块。只保证文件级溯源，不从 MD 行号猜测 PDF 页码。未关联原件时保持原有行为。不支持内嵌展示的格式可按原权限下载。
7. 在版本区域恢复旧版本或解除关联。解除关联保留全部历史文件，占用空间会显示。复制文档会独立复制原件版本；删除知识条目时清理对应原件；物理清理失败保留元数据并记入应用日志，不能伪报文件已释放。

## 接口与日志

- `GET/POST /api/v1/knowledge-bases/:id/knowledge/archive-mapping/batches`
- `GET .../batches/:batch_id` 查看持久化匹配结果。
- `POST .../batches/:batch_id/items/:item_id/file`：multipart `file`、`knowledge_id`、`replace`。SHA256 由服务端计算；重复同一批次项幂等，变更内容或目标需新批次。
- `PUT .../items/:item_id/skip` 跳过待处理项。
- `GET/PUT /api/v1/knowledge/:id/archive-source`：读取版本；PUT `{"version_id":"..."}` 切换版本，空字符串解除关联。
- 原 `/preview` 和 `/download` 支持 `source=auto|parsed|archive`。内部重解析始终读取原来的 MD，不受浏览器源选择影响。

写入仍要求知识库创建者或管理员，读取与预览沿用既有知识库权限，下载沿用既有更严格权限。API 不返回物理路径。批次写入 `events.jsonl` 和可读 `status.txt`，保存在服务器：

```text
/home/ubuntu/RAG_Weknora/data/processing-logs/<知识库ID>/archive-mapping/<批次ID>/
```

```bash
cd /home/ubuntu/RAG_Weknora
find data/processing-logs -path '*/archive-mapping/*/status.txt' -print
docker logs --since 30m WeKnora-app 2>&1 | grep 'ArchiveMapping'
# 替换为实际路径查看逐项结果（不要直接照抄尖括号）
less data/processing-logs/知识库ID/archive-mapping/批次ID/status.txt
```

## 构建、上线与回退

只在允许重启或现有任务结束后执行上线。构建不影响正在运行的应用。原件映射不自动重处理已有文档。

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited sh -lc 'go test ./internal/archivemapping ./internal/handler ./internal/router -count=1 && go test ./internal/application/service -run "Archive|Clone|DeleteKnowledge" -count=1 && go build -buildvcs=false -o runtime/WeKnora.archive-mapping-20261015 ./cmd/server'
mkdir -p runtime/archive-migrations/versioned runtime/archive-migrations/sqlite
cp migrations/versioned/000091_archive_mapping.* runtime/archive-migrations/versioned/
cp migrations/sqlite/000013_archive_mapping.* runtime/archive-migrations/sqlite/
docker build --progress=plain -f deploy/archive-mapping/Dockerfile.runtime -t weknora-app:archive-mapping-20261015 runtime
docker tag archive-intelligence-ui:0.1.0 archive-intelligence-ui:before-archive-mapping-20261015
./frontendctl build custom
mkdir -p data/processing-logs
# 日志目录继承当前目录组，便于 ubuntu 从主机读取容器日志
chmod g+s data/processing-logs
# 将 deploy/pdf-unlimited/compose.yaml 中 app.image 改为新镜像后：
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml config --quiet
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml up -d --no-deps --timeout 120 app
./frontendctl switch custom
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
docker exec WeKnora-app sha256sum /app/WeKnora
sha256sum runtime/WeKnora.archive-mapping-20261015
```

新版镜像继承 `weknora-app:processing-pause-20261015`，携带新增 PostgreSQL 91 / SQLite 13 迁移，不重建或丢弃现有 VLM、PDF、暂停及日志定制。上线后检查新表存在、后台任务继续运行、样例原件和解析 MD 分别可读、替换与解除关联有效，且知识数量与 chunk ID 未改变。最后记录验收结果并分别提交部署与 custom 前端变更。

回退：保留映射表、日志和已存原件，恢复 compose 的 app 镜像为 `weknora-app:processing-pause-20261015`；把 `archive-intelligence-ui:before-archive-mapping-20261015` 重新标记为 `archive-intelligence-ui:0.1.0` 后切换 custom。不要执行 down migration、清空数据库或删除原版前端。所有源码回退使用对应提交的 `git revert`，先核对工作树。
