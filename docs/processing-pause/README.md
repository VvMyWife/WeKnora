# 按知识库暂停和继续后台处理

服务器目录：`/home/ubuntu/RAG_Weknora`。此功能只控制选中的知识库，不暂停其他知识库或用户对话。需要 Redis/Asynq 任务后端，Lite 模式返回不支持。

## 页面操作

1. 打开知识库的“文档”页面，在“全库处理进度”右侧点击“暂停处理”。只有知识库所有者或有管理权限的用户可以操作。
2. 页面显示“正在暂停”和尚在运行的任务数。已启动的任务会完成其当前任务；排队任务不再进入业务处理，也不会扣减待完成子任务数或消耗失败重试次数。
3. 等待显示“已暂停”后，再修改模型服务参数。其他知识库仍可继续调用同一个服务，是否可重启共享模型仍需看其他使用者。
4. 修改后点击“继续处理”，保留的排队任务通常在约 30～45 秒内重新执行。已经归档的历史失败任务不会被自动重放。
5. 刷新或重新登录后仍能看到暂停状态。暂停标志没有自动过期时间；运行任务的计数使用持续续约的租约，应用退出后的失效记录会在约 45 秒内清理。

这属于任务级平稳暂停，已经发出的请求不被强行取消。长 Wiki 任务可能需要较久才能收尾，界面在收尾结束前不会声称“已暂停”。排队任务仍保留原来的模型 ID，暂停不会重写旧任务的模型选择。

## API

```text
GET /api/v1/knowledge-bases/:id/knowledge/processing-control
PUT /api/v1/knowledge-bases/:id/knowledge/processing-control
Content-Type: application/json
{"paused":true}
```

返回 `supported`、`requested_paused`、`state`（running/pausing/paused/unsupported）和 `active_tasks`。写入接口沿用知识库所有者/管理员、知识库写权限及 API Key 范围检查。只读共享用户不能修改。

## 运行测试

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go test ./internal/processingcontrol ./internal/router -run 'TestPause|TestDeadWorker|TestUnsupportedBackend|TestKnowledgePause|TestPausedFinal' -count=1
docker run --rm --network none -v "$PWD:/work" -w /work/frontend-custom node:22-bookworm npm test -- src/views/knowledge/processingControl.test.ts src/views/knowledge/processingOverview.test.ts
git diff --check
git -C frontend-custom diff --check
```

手动验收：用独立测试知识库验证暂停、刷新、继续；再观察暂停中的运行任务数归零、其他知识库继续推进。不要使用删除队列或清空知识库来验证暂停。

## 构建与部署

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go build -buildvcs=false -o runtime/WeKnora.processing-pause-20261015 ./cmd/server
docker build -f deploy/graph-output/Dockerfile.runtime --build-arg BINARY_FILE=WeKnora.processing-pause-20261015 -t weknora-app:processing-pause-20261015 runtime
./frontendctl build custom
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml config --quiet
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml up -d --no-deps --timeout 120 app
./frontendctl switch custom
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
docker exec WeKnora-app sha256sum /app/WeKnora
sha256sum runtime/WeKnora.processing-pause-20261015
```

本版同时包含图谱抽取输出上限 8192。部署后核对最新图谱阶段输入的 `max_output_tokens`，并观察新任务成功完成。重启前已经发出的请求仍使用旧请求参数。

## 日志与回退

## 2026-10-15 运行验收

- 后端镜像 `weknora-app:processing-pause-20261015` 健康，容器二进制与构建产物 SHA256 一致。
- custom 前端在 18180 上已验证暂停、刷新后保留暂停状态、继续处理；正式知识库始终保持 running。
- 独立样例排队期间未执行，恢复后进入处理。样例最终向量化失败是 `18082/v1/embeddings connection refused`，与暂停控制无关，未擅自调整模型服务。
- 新产生的图谱阶段输入已实际记录 `max_output_tokens: 8192`。

## 日志与回退命令

```bash
cd /home/ubuntu/RAG_Weknora
docker logs --since 30m WeKnora-app 2>&1 | grep 'KnowledgeProcessingControl'
```

控制操作输出 `[KnowledgeProcessingControl]`，包含知识库 ID、请求的暂停状态、运行任务数和实际状态。暂停本身不生成 `KnowledgeProcessFailure`。

回退前先在界面继续已暂停知识库，再恢复上一版后端镜像和 custom 前端镜像。上一版后端为 `weknora-app:diagnostics-20261014`，上一版前端在上线时标记为 `archive-intelligence-ui:before-processing-pause-20261015`。旧后端不会执行新暂停规则，因此回退前的恢复操作用于避免后续重新升级时意外保持暂停。不得清空 Redis、删除排队任务或回退数据库。
