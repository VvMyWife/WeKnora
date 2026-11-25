# 聊天图片引用预览

遇到 `weknora://doc/<id>/attachment/0/image/48.png` 这类内部地址，聊天会按本条回答已有的文本引用匹配文档。来源唯一时可打开对应原件。来源无法确认时会提示图片附件不可用。

在服务器权威目录测试和构建。

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm --pull never --network none -v "$PWD:/src" -v ljc-weknora-20261113-ui-deps:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-node:24.16.0-npmmirror bash -ec '
npm test -- src/utils/relativeImageSource.test.ts src/utils/chatMarkdownRenderer.test.ts
npm run type-check
'
p=runtime/chat-document-image
bash "$p/build.sh" "$(cat "$p/source.sha")"
curl -fsS http://127.0.0.1:18182/ -o /dev/null
```

登录 18182，刷新原会话并点击「预览对应原件」。遇到部署中断，先核对镜像标签和构建状态，再决定从哪个阶段继续。

```bash
cat "$p/build.status"
docker inspect WeKnora-frontend-custom-preview --format '{{.Config.Image}}'
git status -sb
```

前端回退使用切换前保存的 Compose 文件。

```bash
docker compose -p ljc-weknora-production -f "$p/compose.before.json" up -d --no-deps --pull never frontend
curl -fsS http://127.0.0.1:18182/ -o /dev/null
install -m 600 "$p/compose.before.json" runtime/production/compose.json
```
