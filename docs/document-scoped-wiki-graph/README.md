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

2026-10-19 全量前端结果：415 项，413 通过；2 项为发布前 `fa6f28f` 已有的 `localeKeyAudit` 韩语、俄语进度条翻译键缺失。已用该基线 `git archive` 导出的原始 src 单独重跑，确认相同的 2 项失败。本功能不修改翻译文件；相关类型检查、档案范围逻辑测试和后端四包测试通过。原始日志在 `runtime/document-graph-*.log`。

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
