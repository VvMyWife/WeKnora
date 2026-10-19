# 档案智查前端

这是 WeKnora 的独立品牌前端版本，源代码位于：

```text
/home/ubuntu/RAG_Weknora/frontend-custom
```

## 一、构建并自检

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl build custom
```

该命令依次执行 Vue/TypeScript 类型检查、生产构建、离线 Docker 镜像构建和 HTTP 冒烟测试。构建过程使用 Node.js 22 容器，不修改服务器系统 Node.js。

项目已在 `.npmrc` 中固定国内 npm 镜像：

```text
https://registry.npmmirror.com
```

## 二、切换版本

启用档案智查：

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl switch custom
```

切回原版：

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl switch original
```

查看当前状态：

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl status
```

正式访问地址始终为：

```text
http://10.20.0.15:18180
```

## 三、卸载定制版

仅卸载定制镜像并恢复原版，保留源代码：

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl uninstall custom
```

恢复原版后，同时删除定制源代码：

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl uninstall custom --purge-source
```

卸载流程会先确认原版前端已启动，再清理定制镜像或源代码。

## 四、开发改动后的完整验证

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl build custom
./frontendctl switch custom
./frontendctl status
curl -fsS http://127.0.0.1:18180/ | grep '<title>档案智查</title>'
```

然后在浏览器检查登录、知识库列表、知识库详情、对话页和设置页。

## 知识库大文件上传上限与自检

当前单文件上限由项目根目录 `.env` 中的 `MAX_FILE_SIZE_MB=256` 控制。修改该值后，前端、应用与文档解析服务都需要重新创建；不要在知识库仍有 `processing` / `finalizing` 任务时重启应用，以免中断正在处理的文档。`frontendctl` 会把同一个值传给网页和 Nginx。

在服务器上查看配置和运行值：

```bash
cd /home/ubuntu/RAG_Weknora
grep '^MAX_FILE_SIZE_MB=' .env
curl -fsS http://127.0.0.1:18180/config.js | grep MAX_FILE_SIZE_MB
docker exec WeKnora-frontend nginx -T 2>/dev/null | grep client_max_body_size
docker inspect WeKnora-app --format '{{range .Config.Env}}{{println .}}{{end}}' | grep '^MAX_FILE_SIZE_MB='
docker inspect WeKnora-docreader --format '{{range .Config.Env}}{{println .}}{{end}}' | grep '^DOCREADER_GRPC_MAX_FILE_SIZE_MB='
```

确认知识库没有正在处理的文件后，按顺序使配置生效：

```bash
cd /home/ubuntu/RAG_Weknora
docker compose -f docker-compose.yml -f docker-compose.vlm-clean.yml up -d --no-deps app docreader
./frontendctl switch custom
```

验证方法：先上传一份原先超过 50 MB、但小于 256 MB 的 PDF；确认上传请求成功，再等待解析状态变为成功。旧式 `.doc` 文件即使上传成功，也可能因为解析器不支持而失败；可另行转换为 `.docx` 或 PDF 后上传。批量上传期间不要关闭上传页；网页会过滤不支持的扩展名和超限文件，需查看上传结果而非仅看知识库当前文档数量。

## 批量上传进度与复测

从知识库的“文档”页选择文件夹并确认上传后，页面会显示已处理、成功、失败数量和当前文件；每处理 10 份会刷新一次文档列表。上传仍由当前浏览器页逐份发送请求，因此应等待进度条完成后再刷新或关闭页面。浏览器刷新时会提示上传仍在进行。

手动复测步骤：

1. 登录 `http://10.20.0.15:18180`，进入测试知识库的“文档”页，点击添加文档并选择一个含多份文件的文件夹。
2. 确认上传，观察进度条从 `0 / 总数` 增长；每 10 份检查目录/文档数是否更新。等待最终的成功与失败提示。
3. 如需验证刷新提醒，在另一个小文件夹的上传中按浏览器刷新键；浏览器应提示上传尚未完成。选择“取消”后，进度应继续增加。
4. 在服务器上核对对应知识库的接收请求（把 `<知识库ID>` 换成浏览器地址中的 ID）：

```bash
docker logs --since 30m WeKnora-frontend 2>&1 | grep 'POST /api/v1/knowledge-bases/<知识库ID>/knowledge/file'
```

日志中的 HTTP 200 表示文件已接收，文档是否解析成功还要查看知识库里的状态。若再次出现 HTTP 400，记录当时是否发生页面刷新，并检查 400 后是否紧跟 `GET /platform/knowledge-bases/`；该序列说明浏览器在请求期间重新加载了页面。

## 空知识库提示与列表分页复测

单文件上限由服务器 `.env` 的 `MAX_FILE_SIZE_MB` 控制；空知识库和拖放区域的提示读取同一运行时配置，不再写死旧的 10 MB/200 KB。更改服务器额度后需要重新创建前端容器，让 `config.js` 更新。

```bash
cd /home/ubuntu/RAG_Weknora
./frontendctl build custom
./frontendctl switch custom
docker exec WeKnora-frontend sh -lc 'grep MAX_FILE_SIZE_MB /usr/share/nginx/html/config.js'
./frontendctl status
```

在浏览器打开一个空知识库，确认提示与 `config.js` 一致；再打开含超过 35 份文档的文件夹，确认首屏加载后自动补取下一页，右侧文档区出现滚动条；滚动到底部应继续加载，直到已加载数等于总数。切换文件夹或筛选条件后，列表应重新从第一页加载。

## 全库文档处理进度复测

此面板依赖 WeKnora 应用接口 `/api/v1/knowledge-bases/<知识库ID>/knowledge/overview`，先确认后端已部署对应版本。它展示整个知识库（含子文件夹）的总完成率、文档解析率、已安排任务的摘要完成率，以及已解析文档的后处理完成率；失败和取消不计入完成。

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm --network none -v "$PWD:/work" -w /work/frontend-custom node:22-bookworm npm test -- src/views/knowledge/processingOverview.test.ts
./frontendctl build custom
./frontendctl switch custom
./frontendctl status
```

登录网页，打开“政策知识库”的文档页，观察“全库处理进度”总数；切换不同子文件夹，面板总数应保持一致；后台仍有任务时约 10 秒自动更新，结束后约 60 秒刷新一次。点击面板刷新按钮可立即重新取数。若后端接口尚未部署，页面只显示“全库进度暂时无法获取”，不会伪造百分比。
# 按知识库暂停后台处理

“全库处理进度”右侧提供“暂停处理／继续处理”。暂停后先等待运行任务收尾，显示“已暂停”后再修改参数。排队任务保留，其他知识库继续运行；恢复后通常约 30～45 秒继续调度。

后端接口、手动验收和完整构建步骤见服务器 `/home/ubuntu/RAG_Weknora/docs/processing-pause/README.md`。

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm --network none -v "$PWD:/work" -w /work/frontend-custom node:22-bookworm npm test -- src/views/knowledge/processingControl.test.ts src/views/knowledge/processingOverview.test.ts
./frontendctl build custom
./frontendctl switch custom
```

## 档案映射

在知识库标题栏点击“档案映射”，选择原件目录，确认“MD 匹配根目录”，再点击“新建匹配预览”。检查相对路径和目标 MD 后，点击“上传已匹配原件”。未匹配、多目标和已有原件需手动处理，不会按同名文件跨目录猜测绑定。原件不进入 OCR、摘要、Wiki 或图谱队列。

浏览器每批并发上传 2 个原件。刷新后打开历史批次并重新选择相同原件目录，已关联项会保留，不重复存储。替换必须明确勾选，旧版本保留并计入存储空间；在“原件版本与回退”中可恢复旧版本或解除关联。文件详情默认查看原件，也可选择“解析文档”或“分块”。聊天里的旧引用按 knowledge ID 动态解析原件。

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm --network none -v "$PWD:/work" -w /work/frontend-custom node:22-bookworm sh -lc 'npm run type-check && npm test -- src/views/knowledge/archiveMapping.test.ts src/views/knowledge/processingControl.test.ts'
./frontendctl build custom
# 后端已部署对应版本后再切换
./frontendctl switch custom
```

后端迁移、日志、完整验收及回退命令见 `/home/ubuntu/RAG_Weknora/docs/archive-mapping/README.md`。只有原件映射上传接口解除解析文件大小限制，其他上传规则不变；仍受租户存储额度和服务器剩余空间约束。

## 外部分块导入

知识库标题旁“导入外部分块”接受原 RAG 的标准 ZIP；先预检，再执行，不再交给内置分块器。支持历史批次恢复、停止后续提交、逐项错误、显式重试和 ID 对照下载。表格 chunk 可展开只含安全表格标签的结构片段，原件通过既有档案映射预览。ZIP 专用上传上限 4 GiB，展开和记录数量另有后端安全上限。

完整导出、测试、导入、日志和部署说明：服务器 `/home/ubuntu/RAG_Weknora/docs/external-chunks/README.md`。本地阶段测试：`npm test -- src/views/knowledge/externalChunks.test.ts`；完整类型检查使用 Node 22 执行 `npm run type-check`。
# 对话中 MD 相对图片与原件溯源

Markdown 中的 `images/page_9_image_1.jpg` 不是可访问的服务器 URL。仅上传 MD 或映射 PDF 不会生成这张 JPG。本版对话渲染器不再把它请求到 `/platform/chat/images/…`：从本轮检索证据中，按完整图片相对路径唯一确认 knowledge ID 后，显示“预览对应原件”，点击在当前对话右侧打开已映射原件。多个文档同路径或缺失证据时明确提示，不猜测；正常 HTTP、resource 和真实沙箱附件链接不变。历史回答重新打开即适用，不改写消息或 chunk。

服务器测试和构建命令：

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm run type-check && npm test -- src/utils/relativeImageSource.test.ts src/utils/chatMarkdownRenderer.test.ts src/utils/referenceSources.test.mjs'
docker tag archive-intelligence-ui:0.1.0 archive-intelligence-ui:before-chat-source-20261019
./frontendctl build custom
./frontendctl switch custom
```

手动验收：打开含相对图片的历史回答，点击“预览对应原件”，应在当前聊天右侧显示原件预览，地址栏不变。核对文档身份和已映射 PDF；未映射时回退解析文件，不能把提示理解为 JPG 已恢复。

同时覆盖模型把知识库图片误写为 `sandbox:page_4_image_0.jpg` 的情况：真实消息附件优先；只有当前消息检索证据能唯一确认来源才转成原件预览，冲突时明确提示而不猜测。测试命令额外加入 `src/utils/sandboxArtifactRefs.test.ts`。本轮侧栏改动前镜像为 `archive-intelligence-ui:before-chat-inline-source-20261019`；回退时重新标记为 `archive-intelligence-ui:0.1.0` 后执行 `./frontendctl switch custom`。只切换 custom 前端，不重启后端和模型。
