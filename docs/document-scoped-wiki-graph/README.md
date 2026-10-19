# 按档案筛选 Wiki 图谱

## 实施基线

服务器 `/home/ubuntu/RAG_Weknora` 为唯一交付位置；`frontend-custom` 为独立前端仓库。两个仓库均沿用 `frontend-custom` 分支，提交身份为 `ljc <ljc@local>`，不重写历史。

2026-10-19 基线：后端 `0fe3fa5922ea0aa2bcc682f7cdd310742a7b256a`，前端 `fa6f28f`。本功能不修改原版 frontend，不调用模型，不重建 Wiki，不改变原件映射或文档处理任务。

发布前应用镜像：`weknora-app:archival-citations-20261019-145943-r2`；custom 镜像：`archive-intelligence-ui:archival-citations-20261019-145943-r3`。新功能需要重新取得后端重启窗口授权。

## 分阶段记录

每阶段执行：检查差异、通过对应测试、核对暂存文件、以 ljc 独立提交、记录提交编号。前后端分别提交。

1. 基线和使用说明。
2. 按档案范围、共享来源与已有跨来源链接的纯逻辑及测试。
3. 数据查询、参数验证、范围内搜索与权限边界及测试。
4. custom 前端档案选择、范围切换、关联高亮与竞态保护。
5. 联合验收、发布清单与回退记录。

## 范围与含义

按档案图谱仅使用已有 Wiki 页面及链接。共享节点表示其来源包含至少两份所选档案；跨来源链接表示端点可追溯到不同的所选档案，不代表模型证明了档案之间的业务关系。无关联时显示空结果，不回退为全库图谱。

选择档案使用 knowledge ID；重复名称不视为同一档案。页数与 PDF 版本定位沿用既有溯源功能，不推算历史页码。

## 测试与发布

所有命令在服务器执行：

```bash
cd /home/ubuntu/RAG_Weknora
git branch --show-current
git -C frontend-custom branch --show-current
git log -5 --format='%h %an <%ae> %cn <%ce> %s'
git -C frontend-custom log -5 --format='%h %an <%ae> %cn <%ce> %s'

docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" \
 -v weknora-go-build-cache:/root/.cache/go-build -w /src \
 weknora-builder:archival-citations sh -c \
 'go test ./internal/application/repository ./internal/application/service ./internal/handler ./internal/router -count=1'

docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c \
 'npm run type-check && npm test -- src/views/knowledge/wiki/documentGraphScope.test.ts && npm run build'
```

如需完整前端回归：

```bash
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c \
 'find src -name "*.test.ts" -print0 | xargs -0 npm test --'
```

2026-10-19 最终全量前端结果：416 项，414 通过；2 项为发布前 `fa6f28f` 已有的 `localeKeyAudit` 韩语、俄语进度条翻译键缺失。已用该基线 `git archive` 导出的原始 src 单独重跑，确认相同的 2 项失败。本功能不修改翻译文件；相关类型检查、3 项档案范围逻辑／显示层级回归测试和后端四包测试通过。原始日志在 `runtime/document-graph-*.log`。

## 使用与人工验收

1. 打开 custom 页面 `http://10.20.0.15:18182`，登录后进入知识库的「图谱」。
2. 点「按档案」→ 搜索解析文件名称 → 勾选 1～50 份 →「应用选择」。列表每页 50 份，覆盖所有文件夹，不受文档页当前目录限制。
3. 映射原件的文件显示 PDF 原件名称和路径，下面保留解析文件名称和目录。普通文档接口仅在 `include_archive_source=true` 时加载这些映射元数据，不影响普通列表开销。
4. 确认节点只来自所选 knowledge ID；可展开「已选档案」查看每份的可见／候选节点数。无 Wiki 来源的档案显示 0，不代表原文件不存在。
5. 多选时紫色外圈表示共享来源，紫色边表示已存在的 Wiki 跨来源链接。切换「只看档案间关联」；无关联时为空，不自动混入其他档案。
6. 搜索 Wiki 页面、双击展开、增量展开、改变类型筛选，均保持所选范围；切换知识库后清空选择。快速切换档案验证旧请求不会覆盖新结果。
7. 切回「全库概览」，原有概览、搜索和展开行为继续可用。原版 18180 不修改。

接口：`GET /api/v1/knowledgebase/{kb_id}/wiki/graph?scope=documents&knowledge_ids={id1}&knowledge_ids={id2}&view=union&limit=200`；`view=connections` 为关联视图。`mode=ego&center={slug}&depth=1` 仍限于选中范围。搜索同一路径下 `/search?q={keyword}`，附带相同 scope、knowledge_ids、view、types 参数；范围内搜索匹配页面标题或 slug。

空选、超过 50 份、非法 ID、已删除及其他知识库 ID 均拒绝。权限由既有 Wiki KB 读权限路由校验，不提供跨库回退。范围查询在数据库按 source_refs 精确匹配，仅读取图字段，不加载 Wiki 正文；兼容裸 ID 与 `ID|title`。没有新增数据库迁移。

## 准备、上线与回退

准备镜像不会重启服务；运行前工作树必须干净：

```bash
bash deploy/document-graph/prepare.sh
# 输出 Prepared only: /home/ubuntu/RAG_Weknora/runtime/document-graph/<发布编号>
release_dir=/home/ubuntu/RAG_Weknora/runtime/document-graph/<发布编号>
python3 deploy/document-graph/release.py "$release_dir" check
```

确认本次重启窗口后才执行：

```bash
python3 deploy/document-graph/release.py "$release_dir" deploy
curl -f http://127.0.0.1:18081/health
curl -f http://127.0.0.1:18182/login >/dev/null
cat "$release_dir/release.json"
cat "$release_dir/deployed.json"
```

准备过程将当前运行镜像作为基础，仅替换后端二进制和 custom dist，保留 VLM 清洗、PDF、脚注等已有定制。发布检查现有镜像、环境及挂载，保留原 custom 容器供恢复；原版 frontend 不覆盖。任一步部署失败自动恢复原镜像与 custom 容器。日志和含配置的现场文件仅保存在权限 0700 的 runtime 发布目录，不提交用户数据与密码。

回退（新增引用、Wiki、原件及文档数据均保留，不执行删除数据的迁移）：

```bash
python3 deploy/document-graph/release.py "$release_dir" rollback
```

代码按对应阶段提交执行 `git revert <提交编号>`，再重新构建部署；不要使用 `git reset --hard`。

已提交：基线 `bdfa11325`，图谱契约与纯逻辑 `abdfcafd4`，查询／参数验证 `076ccb121`。后续提交及实际镜像对应关系以发布目录 `release.json`、`deployed.json` 为准，不能将 prepared_not_deployed 视为上线成功。

## 最终发布与现场验收（2026-10-19）

- 发布目录：`runtime/document-graph/20261019-195330`。
- 后端代码 `f50fd80d944dba313d0a0cfb37415fdd1026f4ea`；镜像 `weknora-app:document-graph-20261019-195330`。
- custom 代码 `30095372d73d168f659ed859ba770d4722ca696d`；镜像 `archive-intelligence-ui:document-graph-20261019-195330-r2`。
- 后端镜像 ID：`sha256:d2e2fd0b7893cf8604c04548818e4ed7756d51d63f4dad9e2620bc251ba0462c`。
- custom 镜像 ID：`sha256:abaab2b5e37fba000cf2472ef629fb200a75826f7a7d4e0cbf6440500e18b55b`。
- 正式入口仍为 18182；原版 18180 的镜像和启动时间未改变。数据库迁移版本仍为 `94`、`dirty=false`，未新增迁移。

真实只读验收：编目样本库选择 3 份档案，合并 30 个节点，关联 26 个节点、52 条跨来源链接；政策正式库选择 3 份档案，合并 77 个节点，关联 43 个节点、70 条跨来源链接。正式库选择列表首批 50 份均显示已映射 PDF 名称、原件路径和 MD 解析文件名。搜索、ego 展开、截取计数、空选／无效 ID 拒绝、来源 ID 对照均通过；全部验收请求为 GET，不调用模型，不创建处理任务。

真实页面验证了选择／应用、翻页保留选择、单档案 11 个节点、多档案 30 个节点、关联视图 26 个节点、单档案关联空结果、回到全库，以及切换知识库清空选择。正式 18182 已实际操作并显示政策档案 77 个节点；不是仅检查 health。

验收发现并修复的两个问题均有独立提交：历史 `source_refs=JSON null` 查询兼容（`f50fd80d9`）；空图覆盖选择窗口（`3009537`）。前者发现后回退旧版，再经真实 PostgreSQL 只读校验后重新发布；后者只恢复 custom 前端，在独立 18184 预览页验收后更新前端，未再重启后端。回退记录与前端候选变更记录保留在发布目录。临时预览验收完成后停止，不作为正式入口。

### 复跑浏览器只读接口验收

在已安装 Playwright CLI 的测试电脑上，先打开 18182，手动登录并进入至少有 2 份已生成 Wiki 文档的知识库，再运行仓库中的脚本（脚本不包含密码）：

```bash
playwright-cli -s=graph-acceptance open http://10.20.0.15:18182 --browser=msedge --headed
# 手动登录，进入目标知识库后：
playwright-cli -s=graph-acceptance snapshot
playwright-cli -s=graph-acceptance run-code --filename deploy/document-graph/live-api.js
```

`live-api.js` 仅访问当前已授权知识库；无 Wiki 来源的全新库不适合作为此脚本的正向样例。400 负向用例会在浏览器控制台显示对应资源拒绝，这是预期结果。

### 仅修复前端时的预览与发布

当前后端已通过验收而无需重启时，在同一发布目录登记新前端候选。先完成前端测试、构建、Git 提交，并用相同 Dockerfile 创建新镜像；不要覆盖运行镜像标签。独立预览保留现有网络和环境，只改测试容器名称、18184 端口及关闭自动重启：

```bash
release_dir=/home/ubuntu/RAG_Weknora/runtime/document-graph/20261019-195330
candidate_image=archive-intelligence-ui:你的新前端版本
python3 deploy/document-graph/preview.py "$release_dir" "$candidate_image"
# 在 http://10.20.0.15:18184 完成页面验收之后：
frontend_revision=$(git -C frontend-custom rev-parse HEAD)
python3 deploy/document-graph/release.py "$release_dir" candidate \
 --image "$candidate_image" --frontend-commit "$frontend_revision"
python3 deploy/document-graph/release.py "$release_dir" frontend
docker stop WeKnora-frontend-custom-document-graph-qa
```

预览容器名固定且已存在时脚本会拒绝覆盖；先核对是否仍有人使用，不要删除或覆盖不明容器。`frontend-deployed.json` 记录最终前端提交与 `backend_unchanged=true`；`release.json` 保存最终前后端镜像映射，`deployed.json` 保留最初整体验收部署记录。回退命令仍使用同一发布目录。
