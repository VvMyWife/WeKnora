# 档案脚注隔离浏览器验收

这是开发用页面，不进入生产 Vite 的 index.html 入口。它验证组件，不代替真实模型／知识库验收。源代码位于服务器 `/home/ubuntu/RAG_Weknora/frontend-custom`，本目录随 custom 仓库独立提交，作者／提交者 `ljc`。

## 服务器启动

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm run type-check && npm test -- --test-reporter=dot src/utils/archivalCitations.test.ts src/utils/chatMarkdownRenderer.test.ts src/utils/archivalStream.test.ts'
docker run -d --name weknora-citations-qa-ui -p 18185:5173 -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm npm run dev -- --host 0.0.0.0 --port 5173 --strictPort
```

容器已存在时不要重复创建，先 `docker inspect weknora-citations-qa-ui`；已有容器停止后用 `docker start`。它与正式 18182 分开，不要重启正式应用。

## Windows 浏览器逐项操作

playwright-cli 已持久安装到 `E:/download/Claude-Code/npm-global`，npm registry 默认国内镜像，浏览器使用本机 Edge。以下在本目录父项目的本地测试副本执行；若路径不同，替换 filename。

```powershell
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa open http://10.20.0.15:18185/qa/archival.html --browser=msedge
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa run-code --filename qa/boot.js
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa snapshot
```

点击最新 snapshot 中的脚注 ref，检查完整原文、表格 rowspan、复制、切换脚注及目录。ref 每次重新 snapshot 获取，不能沿用旧编号。

PDF：把一个至少 16 页、页面内容可辨的测试 PDF 放在 `qa/pdf.js` 指定的持久本地路径（或修改此文件的 fixture 路径）。执行：

```powershell
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa run-code --filename qa/pdf.js
```

已测试：脚注 1 在第 2 页，滚轮到第 3 页，页码框到第 16 页；缩放／全屏；脚注 2 无页码明确回首页。执行 `qa/page-error.js` 后重开脚注 1，999 页须警告并回首页，而非伪装成准确定位。

执行 `qa/deny.js` 模拟权限撤销，重新打开脚注时不能残留旧原文；执行 `qa/legacy.js` 验证历史无记录提示。执行 `qa/wiki.js` 后检查派生来源关系展开、当前来源分块与引用快照区分、没有虚假的 PDF 定位按钮。

```powershell
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa run-code --filename qa/wiki.js
& E:/download/Claude-Code/npm-global/playwright-cli.cmd -s=archival-qa snapshot
```

每个脚本使用路由模拟，最新添加路由优先。回到标准样例时重新运行 `qa/boot.js`。真实上线验收必须使用主仓库 `docs/archival-citations/README.md` 的授权窗口和独立样例库流程。
