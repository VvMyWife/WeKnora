# 18182 引用页预览

回答正文完成后，前端请求引用清单的 `preview=pages` 视图。服务端先查原件版本与被引用 chunk 的物理页锚点，然后沿 `next_chunk_id` 最多检查三个后续 chunk。相邻块仅用于定位。相同原件版本的同一页合并显示。点缩略图可以原地放大，缺页码和渲染失败时保留整份原件预览入口。

## 在权威目录测试

1. 检查当前分支、改动与测试端口。

   ```bash
   cd /home/ubuntu/RAG_Weknora
   git status --short --branch
   git log -5 --oneline
   ss -Hlt '( sport = :28281 or sport = :28282 )'
   ```

2. 跑后端引用测试，并检查前端单测和类型。

   ```bash
   docker run --rm --pull never --network none \
     -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" \
     -v weknora-go-build-cache:/root/.cache/go-build -w /src \
     weknora-builder:release-20261112-tests \
     go test ./internal/archivalcitations ./internal/handler -count=1 \
       -skip '^TestPutTenantParserConfigAdminPreservesRedactedSecrets$'
   bash runtime/chat-cited-page/test-ui.sh
   ```

   租户解析器的脱敏配置测试在上一个维护提交和本次提交中均返回 HTTP 400，预期为 200。单独定位时运行 `go test ./internal/handler -run '^TestPutTenantParserConfigAdminPreservesRedactedSecrets$' -count=1`。

3. 构建镜像。本次两套镜像的源码提交为 `a57f3fa11`。测试后新增的说明提交不参与镜像构建。

   ```bash
   p=runtime/chat-cited-page
   sha=$(cat "$p/build-source.sha")
   git merge-base --is-ancestor "$sha" HEAD
   git diff --exit-code "$sha" HEAD -- internal frontend-custom migrations
   bash "$p/build-app.sh" "$sha"
   bash "$p/build-ui.sh" "$sha"
   ```

4. 使用现有隔离队列和测试库验证接口。测试脚本只临时补入测试 chunk 的页码，在结束时恢复。该测试复用 2026-11-25 隔离验收设施。

   ```bash
   p=runtime/chat-cited-page
   docker start ljc-20261125-proof-pg
   docker compose -f "$p/compose.candidate.json" config --quiet
   docker compose -f "$p/compose.candidate.json" up -d --pull never
   curl -fsS --retry 15 --retry-all-errors --retry-delay 2 http://127.0.0.1:28281/health
   python3 "$p/candidate-api.py"
   ```

   在 28282 打开有引用页码的回答。确认同一 PDF 页只显示一次，点开图片时地址不变。再检查缺页码或越界的回答，确认旧原件入口还能打开 PDF。第 4 个后继块和版本不一致的情况由后端测试覆盖。

5. 关闭候选服务，检查生产入口。

   ```bash
   docker compose -f runtime/chat-cited-page/compose.candidate.json stop
   docker stop ljc-20261125-proof-pg
   curl -fsS http://127.0.0.1:18081/health
   curl -fsSI http://127.0.0.1:18182/
   ```

## 发布与回退

- 发布脚本使用已安装的镜像、权威分支和测试状态文件，切换生产 app 和 18182。新旧 Compose 和镜像记录保存在 `runtime/chat-cited-page/`，包含内部部署信息。执行之前查看 `deploy.log`、`deploy.status` 和生产容器镜像。已有 `compose.before.json` 时先核对之前是否完成切换。

  ```bash
  cd /home/ubuntu/RAG_Weknora
  git status --short --branch
  p=runtime/chat-cited-page
  nohup bash -c 'python3 -u runtime/chat-cited-page/deploy.py; echo $? > runtime/chat-cited-page/deploy.status' > "$p/deploy.log" 2>&1 < /dev/null &
  echo "发布 PID=$!"
  tail -f "$p/deploy.log"
  ```

- 中断后先检查运行状态。状态文件为零时，核对单独保存的产物记录。失败时用日志定位问题，新建验收步骤时复用同一目录。

  ```bash
  p=runtime/chat-cited-page
  docker ps -a --filter name=ljc-cited-page --format '{{.Names}} {{.Status}}'
  docker inspect WeKnora-app WeKnora-frontend-custom-preview --format '{{.Name}} {{.Image}}'
  cat "$p/deploy.status"
  tail -n 35 "$p/deploy.log"
  ```

- 如需回退镜像，在空闲窗口恢复部署前的 Compose。这个版本未更改数据库结构，回退沿用同一个生产库。

  ```bash
  cd /home/ubuntu/RAG_Weknora
  p=runtime/chat-cited-page
  docker compose -p ljc-weknora-production -f "$p/compose.before.json" config --quiet
  docker compose -p ljc-weknora-production -f "$p/compose.before.json" up -d --no-deps --pull never app frontend
  curl -fsS --retry 15 --retry-all-errors --retry-delay 2 http://127.0.0.1:18081/health
  curl -fsSI http://127.0.0.1:18182/
  install -m 600 "$p/compose.before.json" runtime/production/compose.json.rollback
  mv runtime/production/compose.json.rollback runtime/production/compose.json
  ```
