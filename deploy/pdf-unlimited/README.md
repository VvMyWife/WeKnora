# WeKnora PDF 全页解析与全库进度

此目录用于构建并部署 PDF 全页解析版本。PaddleOCR-VL 不再限制只接收第一页；WeKnora 会对比源 PDF 页数与返回页数，页数不一致时记录失败，不把不完整结果当作成功。已有知识库不会自动重建，历史文档的页数校验状态为“未验证”。

## 自行核对现状

在服务器上执行：

```bash
cd /home/ubuntu/RAG_Weknora
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml --env-file .env ps
curl -fsS http://127.0.0.1:9960/full-page-status
curl -fsS http://127.0.0.1:18081/health
curl -I http://127.0.0.1:18180/login
```

`full-page-status` 中的 `max_num_input_imgs` 应为 `null`。打开网页并进入知识库的“文档”页，查看全库处理总进度、PDF 页数校验以及失败文件列表。切换文件夹后全库总数不应变化。对一个新上传的多页 PDF，处理完成后点开“查看分块”，应能看到后续页内容；校验状态应显示“已验证”。

## 构建与部署

构建会使用 `deploy/pdf-unlimited/Dockerfile.builder` 中的国内 Debian 和 Go 镜像配置，前端依赖由仓库现有 `node_modules` 提供。构建脚本会先运行相关测试，再产出后端和定制前端镜像：

```bash
cd /home/ubuntu/RAG_Weknora
bash deploy/pdf-unlimited/build.sh
cat runtime/full-build.status
bash deploy/pdf-unlimited/deploy.sh
cat runtime/deploy-unlimited.status
```

两个状态文件的 `0` 表示对应脚本正常结束。部署脚本带文件锁；若中途断开，可再次执行同一脚本。Paddle 服务由 `paddleocr-vl.service` 管理，部署前先确认其已加载不限页配置并响应上面的状态接口。该脚本保留 `before-pdf-unlimited` 镜像标签供回退。定制前端使用 `deploy/frontend/frontendctl switch custom`，不修改原版前端源码。

## 出错时核对

```bash
cd /home/ubuntu/RAG_Weknora
docker logs --tail 100 WeKnora-app
docker logs --tail 100 WeKnora-docreader
systemctl status paddleocr-vl.service --no-pager
journalctl -u paddleocr-vl.service -n 100 --no-pager
```

优先在网页的失败列表查看文件名、处理阶段、原因和重试状态。若页数不一致，先比对源 PDF 与 Paddle 返回页数；不要对当前示例批次自动重处理。构建或部署中断时，先查看状态文件、容器状态和健康接口，再重跑相应脚本，避免重复启动多个作业。
