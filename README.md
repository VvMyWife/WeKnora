# WeKnora v0.8.0 部署与测试说明

本目录是 WeKnora 的唯一项目边界：源代码、构建覆盖、OCR/BGE-M3 适配、评测数据、日志与报告都放在 `/home/ubuntu/RAG_Weknora`。现有 `/home/ubuntu/RAG`、OCR、vLLM 和其数据均不写入、不重启。

## 固定版本

- 上游：Tencent/WeKnora
- 版本：`v0.8.0`
- 固定提交：`1edcd54b43606d9079bb36650efe3f68707a79ea`
- 本地部署分支：`deploy/weknora-v0.8.0`

验证：

```bash
cd /home/ubuntu/RAG_Weknora
git status --short
git rev-parse HEAD
git describe --tags --exact-match
```

## GPU 规则

模型推理不得回退到 CPU。启动任何模型前执行：

```bash
nvidia-smi --query-gpu=index,memory.total,memory.used,utilization.gpu --format=csv,noheader
cat deploy/gpu-selection.json
```

优先使用 GPU 3；不足时按 GPU 2、1、0 尝试（始终以当次空闲显存为准）。WeKnora 的 Web、数据库和文档编排容器不属于模型推理；OCR 和 BGE-M3 必须使用 GPU。

## 首次构建（国内镜像）

`.env` 仅保存在本机，含数据库与 Redis 随机口令，禁止提交或打印。

```bash
cd /home/ubuntu/RAG_Weknora
docker compose --env-file .env config --quiet
docker run --rm -e npm_config_registry=https://registry.npmmirror.com   -v "$PWD/frontend:/app" -w /app node:22-bookworm   bash -lc 'npm ci && npm run build'
DOCKER_BUILDKIT=1 docker compose -f docker-compose.yml -f docker-compose.build-network.yml   --env-file .env build   --build-arg GOPROXY_ARG=https://goproxy.cn,direct   --build-arg GOSUMDB_ARG=off   --build-arg APK_MIRROR_ARG=mirrors.tuna.tsinghua.edu.cn   --build-arg APT_MIRROR=https://mirrors.tuna.tsinghua.edu.cn   app docreader frontend
```

`docker-compose.build-network.yml` 仅为构建网络覆盖。`packages/grpc_health_probe-linux-amd64` 为通过国内代理预取的固定健康检查二进制，避免构建时直连 GitHub。

## 启动、检查与停止

```bash
cd /home/ubuntu/RAG_Weknora
docker compose --env-file .env up -d
docker compose --env-file .env ps
curl -fsS http://127.0.0.1:18081/health
curl -fsSI http://127.0.0.1:18180/
docker compose --env-file .env logs --tail=150 app docreader
# 停止但保留 WeKnora 数据卷
docker compose --env-file .env stop
# 重新启动
docker compose --env-file .env up -d
```

服务端口固定为：前端 `18180`、API `18081`，避免与既有服务冲突。

## 复用能力的接入顺序

1. OCR：先只探测既有 `9960`、`9961` 健康接口；随后按 WeKnora 模型/API 页面配置外部解析服务，不复制或改写原始 OCR 结果。OCR 证据副本只放 `data/ocr-evidence/`。
2. BGE-M3：复用既有权重，适配服务放 `adapters/bge_m3/`，由现有 eval-venv 加载既有权重并固定 `CUDA_VISIBLE_DEVICES=3`，以 GPU 运行；WeKnora 中创建 OpenAI-compatible Embedding 与 Rerank 模型配置。
3. LLM：复用既有 `8101` 或 `8102` 的 OpenAI-compatible vLLM，不启动重复大模型。
4. 评测：先运行 `bash evaluation/run_smoke.sh` 验证 API、OCR、vLLM、BGE-M3；测试集放 `evaluation/datasets/`，每次运行放 `evaluation/runs/<run-id>/`，报告放 `reports/`；比较 WeKnora 默认能力与接入 OCR/BGE-M3 后的结果。

## 故障恢复

```bash
cd /home/ubuntu/RAG_Weknora
# 仅观察，不修改既有服务
docker compose --env-file .env ps
tail -200 logs/build-docreader.log
docker compose --env-file .env logs --tail=200 app docreader postgres redis
# 只重建 WeKnora 指定镜像
docker compose -f docker-compose.yml -f docker-compose.build-network.yml --env-file .env build docreader
```

不要执行 `down -v`，除非明确要删除 WeKnora 自己的数据库和文件卷；该命令不可恢复。

## BGE-M3 GPU 服务

```bash
cd /home/ubuntu/RAG_Weknora
bash deploy/start-bge-m3.sh
curl -fsS http://127.0.0.1:18082/health
bash evaluation/run_smoke.sh
# 停止适配服务
bash deploy/stop-bge-m3.sh
```

该适配服务固定使用宿主 GPU 3（进程内显示为 `cuda:0`），模型权重只读复用 `/home/ubuntu/RAG/runtime/models/BAAI/bge-m3`；不下载第二份模型，不回退 CPU。

## 网页配置出现 SSRF validation failed

如果网页版测试内网模型地址提示 SSRF validation failed，确认 `.env` 中包含 `SSRF_WHITELIST_EXTRA=10.20.0.15,...`，然后执行 `docker compose --env-file .env up -d --force-recreate app`。不要关闭 SSRF 防护，也不要把白名单设为 `*`。

## 全库文档处理进度（开发与验收）

知识库「文档」页的进度面板统计整个知识库，包含所有子文件夹，不受当前文件夹和列表分页影响。总完成率为 `completed / 全部文档`；解析完成率为 `(finalizing + completed) / 全部文档`；摘要完成率只统计已经进入摘要任务的文档；后处理完成率为 `completed / 已解析文档`。失败与取消单独列出，不算完成。单个文档在「优化中」时可能正在等待摘要、问题生成或图谱等后续任务，并非精确的单文件百分比。

在服务器上逐步验证代码和接口（不要在大量文档仍在处理时重启 `app`）：

```bash
cd /home/ubuntu/RAG_Weknora
docker volume create weknora-go-mod-cache
docker volume create weknora-go-build-cache
docker run --rm -v "$PWD:/src" -v weknora-go-mod-cache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build -w /src \
  -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=off \
  golang:1.26-bookworm go test ./internal/application/repository ./internal/handler ./internal/application/service ./internal/router
docker run --rm --network none -v "$PWD:/work" -w /work/frontend-custom \
  node:22-bookworm sh -lc 'npm test -- src/views/knowledge/processingOverview.test.ts && npm run type-check && npm run build'
```

待后台处理批次完成后，使用仓库当前的构建配置重建并部署应用及前端；先确认 `docker compose --env-file .env ps`，再在浏览器登录、进入任意有子文件夹的知识库，核对面板总数与目录树根节点总数一致，切换子文件夹后总数保持不变，并观察后台处理时数值自动更新。接口可用已登录浏览器请求 `GET /api/v1/knowledge-bases/<知识库ID>/knowledge/overview` 验证。浏览器刷新后进度仍由服务端重新计算，不依赖本地缓存。
