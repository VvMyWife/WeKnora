# 图谱抽取输出上限

服务器源码目录：`/home/ubuntu/RAG_Weknora`。

后台图谱分块抽取请求的 `max_tokens` 从 4096 调整为 8192；对话查询实体抽取仍为 4096。模型服务的 `max_model_len=49152` 不变。输入与预留输出仍需同时容纳在模型上下文内。

调整针对生成被截断的风险，不会修复模型生成的非法 JSON，也不会自动重放已归档失败任务。模型格式约束、原始响应和结束原因记录属于独立改进。

## 自己测试

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go test ./internal/application/service/chat_pipeline -run 'TestExtractionCompletionBudgets|TestGraphExtractorRejectsInvalidJSON|TestFormater_ParseGraph_FenceVariants|TestExtractJSONLike' -count=1
git diff --check
```

测试会捕获实际抽取请求参数，验证后台图谱为 8192、查询实体为 4096，并验证错误 JSON 仍会报告失败。

## 构建和上线

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go build -buildvcs=false -o runtime/WeKnora.graph-output-8192-20261015 ./cmd/server
docker build -f deploy/graph-output/Dockerfile.runtime --build-arg BINARY_FILE=WeKnora.graph-output-8192-20261015 -t weknora-app:graph-output-8192-20261015 runtime
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml config --quiet
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml up -d --no-deps --timeout 120 app
docker inspect WeKnora-app --format '{{.Config.Image}} {{.State.Health.Status}}'
docker exec WeKnora-app sha256sum /app/WeKnora
sha256sum runtime/WeKnora.graph-output-8192-20261015
docker logs --since 5m WeKnora-app 2>&1 | grep -E 'KnowledgeProcessFailure|graph|extract'
```

部署后核对镜像与二进制哈希，并观察真实后台图谱任务继续推进。新图谱阶段的输入记录中包含 `max_output_tokens: 8192`，可以在文档阶段诊断或数据库轨迹中核对。重启前已发出的模型请求不会改变输出预算，新应用发出的请求使用 8192。

## 回退

前一镜像为 `weknora-app:diagnostics-20261014`。将 `deploy/pdf-unlimited/compose.yaml` 中 app 镜像恢复为该值，再执行上述 Compose 上线命令。通过 Git 反向提交记录回退；不删除数据库或处理任务。
