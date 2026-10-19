# 档案级回答引用：实施与验收

服务器源码：`/home/ubuntu/RAG_Weknora`；定制前端为其独立 `frontend-custom/` 仓库。两个仓库均沿用 `frontend-custom` 分支。所有新增提交作者与提交者为 `ljc <ljc@local>`。不推送，不重写历史，不覆盖原版 frontend。

## 实施前基线（2026-10-19）

- 后端：`077ee148`，应用 `weknora-app:graph-output-12288-20261016`，镜像 ID `f113613c6a5ae07f9752548a3b5ca3712e34b37fcb0da9627a7bd6985d49944f`。
- custom 源码：`8201116`。重新核对运行容器：18182 的 `WeKnora-frontend-custom-preview` 镜像 ID 为 `e07b71a3cb129acafb0577d8f04f1e860537521deae88209ba5fab06c8398505`；18180 的 `WeKnora-frontend` 为 `28763a77429697caf8a3caa631cc7c24ab34e61617e0db518a6b72f694a3d00c`。本次只涉及 custom 源码及 18182，不替换 18180，也不修改原版 frontend 源码。
- 既有模型、图谱并发、摘要预算、切块与上传规则不在本次调整范围。

## 分阶段交付

0. Git 身份、基线、长期约定与本说明。
1. 引用契约及兼容测试。
2. 消息证据快照、权限接口和原件版本读取。
3. 普通问答与智能体引用验证、一次补引和未绑定提示。
4. 当前聊天内的 chunk 证据面板、复制与版本信息。
5. 本地 PDF.js 连续阅读、可信页定位和缺页码回退。
6. 按回答实际引用生成来源目录。
7. 独立样例库验收，固定镜像与回退清单。

每阶段独立测试、`git diff --check`、检查暂存清单后提交。提交后核对 `%an <%ae> | %cn <%ce>`。未完成的阶段不表述为已上线。

## 基本约束

保留引用时的 chunk 文本和版本，来源目录只列实际引用，不列全部检索结果。有可信物理页码才定位；无页码打开首页并标识。不得根据图片名、chunk 序号猜页码，不补跑历史 OCR、不重新切块。派生 Wiki/图谱不可冒充原始 PDF。快照和版本读取仍检查消息和知识库权限。

## 手动测试入口

```bash
cd /home/ubuntu/RAG_Weknora
git var GIT_AUTHOR_IDENT
git var GIT_COMMITTER_IDENT
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited sh -c 'go test ./internal/archivalcitations -count=1'
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm run type-check && npm test -- src/utils/archivalCitations.test.ts'
```

上线需要重新核对活动任务并确认后端重启窗口；不能沿用其他功能的重启授权。回退保留引用记录，不执行删证据的 down migration。2026-10-19 用户已为本功能单独授权立即重启，接受当前任务中断重试。

## 实际阶段提交（2026-10-19）

| 阶段 | 主仓库 | custom 仓库 | 状态 |
|---|---|---|---|
| 0 基线与身份 | 61848a25 | 不涉及 | 已保存长期约定 |
| 1 契约 | a0bcb622 | 836d000 | 兼容测试通过 |
| 2 快照、权限、指定原件版本 | 1fed9295 | 不涉及 | 测试通过；生产迁移已到干净 94 版本 |
| 3 引用验证与一次补引 | d5a180de | 2a40455 | 普通问答、智能体均接入 |
| 4 证据面板 | 不涉及 | b087d05 | 完整原文／纯文本复制 |
| 5 PDF.js | 不涉及 | dbde23e | 本地打包，无 CDN，连续滚动 |
| 6 实际引用目录 | ad84c394 | aa8d460 | 版本、页码、分块去重 |
| 7 联合检查修正 | 3f9d7384 | 8759477 | Wiki 关系、权限、流结束回归通过 |
| 发布修正 | f0a51dd8 | a990e45 | 迁移目录权限、PDF worker MIME、恢复流视图更新 |

每次运行 `git show --format=fuller <编号>` 可查看 Author／Committer；新提交均为 `ljc <ljc@local>`。发布准备文档和脚本另作独立提交，不把“已构建”表述为“已上线”。

## 接口、版本与边界

- 默认关闭：`ARCHIVAL_CITATIONS_ENABLED=false`；本次在独立授权窗口开启，已部署至 18081 后端和 18182 custom 前端。
- `GET /api/v1/messages/{session}/{message}/citations`：实际引用清单、校验状态及当前不可访问的数量。
- `GET /api/v1/messages/{session}/{message}/citations/{citation}`：引用时快照、当前内容差异、固定原件版本。
- `GET .../citations/{citation}/origins/{knowledge}/chunks/{chunk}`：仅允许已记录的来源关系，仍逐次检查当前权限，返回 `current_source_not_snapshot`，不会替换引用时快照。
- `GET /api/v1/knowledge/{knowledge}/preview?source=archive&archive_version_id={version}`：严格读取指定版本；删除、无权限、版本不匹配不静默替换成最新版。
- PG 迁移 `000094`，SQLite 迁移 `000015`；只有新增证据表，无破坏性回退迁移。应用启动时从 `/app/migrations` 执行迁移。
- 快照固定于本轮模型调用之前，最多登记 256 个实际引用；单个快照上限 4 MiB。原件版本同时固定。消息正文后来被编辑时，原清单返回 `answer_changed`，不伪装为当前答案的有效依据。
- 一次补引只允许返回“已有证据 ID＋未绑定行号”，不能改写答案；补引输入按完整证据保守限制为 28,000 字节，超出则不截断表格、不强行调用。仍无依据的段落保留并标识。绑定校验不等于事实真实性审核。
- Wiki／图谱始终标为派生证据，来源关系可以展开查看；关系失效、源文件删除或权限撤销不绕过检查。无可信原始页关系时不能直接声称对应 PDF 页。
- 页码只采用外部分块包的可信 `source_anchors.page_index`（零基、转为物理页一基），并核对原件 SHA256。不从 chunk 顺序、图片文件名或正文页码猜测。旧数据无页码时打开首页并标识，不回填历史数据。
- PDF.js `6.4.299` 固定版本，worker、cmaps、fonts、wasm、iccs 均本地部署。仅按需渲染相邻页面，越界或错误版本会明确回退。

## 自己运行后端和前端测试

```bash
cd /home/ubuntu/RAG_Weknora
git branch --show-current
git -C frontend-custom branch --show-current
git var GIT_AUTHOR_IDENT
git var GIT_COMMITTER_IDENT
git -C frontend-custom var GIT_AUTHOR_IDENT
git -C frontend-custom var GIT_COMMITTER_IDENT
git diff --check
git -C frontend-custom diff --check
docker build -f deploy/archival-citations/Dockerfile.builder -t weknora-builder:archival-citations deploy/archival-citations
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:archival-citations sh -c 'go test ./internal/archivalcitations ./internal/modelcontext ./internal/application/service/chat_pipeline ./internal/agent ./internal/agent/tools ./internal/handler/session ./internal/handler ./internal/router ./internal/container ./internal/database ./internal/archivemapping -count=1'
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm run type-check && npm test -- --test-reporter=dot src/utils/archivalCitations.test.ts src/utils/chatMarkdownRenderer.test.ts src/utils/archivalStream.test.ts && npm run build'
```

GOPROXY／GOSUMDB 为构建镜像的全局默认，npm 使用 custom 内 `.npmrc` 的国内镜像。`node_modules` 是 custom 独立持久目录，不修改原版前端依赖。缺依赖时在服务器持久目录执行 `npm ci`（容器挂载同一仓库），不要放临时工作区。

已通过：11 个后端包、前端类型检查、47 项针对性前端测试及打包。浏览器隔离验收覆盖：脚注点击、完整 HTML 表格及合并单元格复制、纯文本复制、快照／当前文本差异、403 后旧证据清除、历史无记录、Wiki 来源关系、PDF 第 2／3／16 页真实画面、滚动、缩放、全屏、无页码与越界回退、目录点击到指定页。测试脚本见 `frontend-custom/qa/README.md`。

隔离浏览器使用合成接口响应验证边界情况，不冒充真实知识库全链路验收。上线后另建“档案脚注验收-20261019-150903”样例库，已用真实模型完成普通问答与智能体回答、快照及原件哈希、匿名拒绝、历史重载、实际 PDF 页显示／滚动／来源目录检查。Wiki／图谱派生关系、撤销共享权限、原件换版和删除等边界由回归测试覆盖，未针对正式业务库执行破坏性演练；真实大文件及业务档案抽查仍需人工复核。详见 `acceptance-20261019.md`。

验收脚本使用 `requests`（服务器或本地持久环境），密码交互输入，不写入仓库。脚本会创建独立样例库，只使用指定状态文件中的样例 ID，不修改正式库。先备好至少 16 页测试 PDF：

```bash
cd /home/ubuntu/RAG_Weknora
python3 deploy/archival-citations/acceptance.py setup --email 你的邮箱 --state runtime/new-citation-acceptance.json --pdf runtime/pdf-fixtures/sixteen.pdf --schema internal/externalchunks/manifest.schema.json
python3 deploy/archival-citations/acceptance.py inspect --email 你的邮箱 --state runtime/new-citation-acceptance.json
# 等 inspect 显示文档 completed 后，顺序运行，避免并行覆盖同一个报告。
python3 deploy/archival-citations/acceptance.py normal --email 你的邮箱 --state runtime/new-citation-acceptance.json
python3 deploy/archival-citations/acceptance.py agent --email 你的邮箱 --state runtime/new-citation-acceptance.json
python3 deploy/archival-citations/acceptance.py verify --email 你的邮箱 --state runtime/new-citation-acceptance.json
```

`setup` 若状态文件已存在会拒绝重复创建；失败后检查该文件／样例库，不直接反复 setup。`normal`／`agent` 会产生新的测试会话。`verify` 只读核对 SSE 最终答案与历史一致、快照内容 SHA、指定原件 SHA 和匿名访问拒绝，不声称验证模型事实正确。样例只用于功能验证，不混入正式政策知识库。

## 只准备镜像，不影响线上

先提交本阶段；脚本要求两个仓库的已跟踪文件干净，不会修改现有容器或应用数据库：

```bash
cd /home/ubuntu/RAG_Weknora
bash -n deploy/archival-citations/prepare-release.sh
bash deploy/archival-citations/prepare-release.sh
ls -lt runtime/archival-citations/
```

输出目录 `runtime/archival-citations/<时间>/` 内有 `release.json`、构建日志、二进制及发布资产。`app-before.json`／`frontend-before.json` 包含完整回退配置，可能含敏感环境变量；目录权限 700、文件 600，不提交 Git、不对外分享。镜像直接在当前运行镜像之上叠加二进制／SQL／custom dist，不丢失既有 VLM-clean、全页 PDF、上传配置或 nginx 定制。镜像内迁移目录须为 755、SQL 为 644；私有发布目录的 700 权限不能直接成为应用用户不可读的镜像目录。构建结束额外以 appuser 验证读取，部署不能只看 `/health`，还要确认 `schema_migrations=94,false`。

`runtime-release.py` 提供明确的 `check / backup / app / frontend / rollback` 操作，固定回退环境，比较现有环境和挂载，避免 Compose 源配置与实际运行配置不一致。先运行 `check`，核对后在授权窗口顺序执行 `backup`、`app`、`frontend`。已有完整备份不能重复 `backup` 覆盖；发布目录固定为本次脚本输出。它不代替真实样例验收。

## 上线窗口（必须重新获准后执行）

1. 先确认没有需要保留的处理中任务，或者用户明确接受中断重试。
2. 保存数据库备份与迁移版本；检查 `release.json` 提交、镜像 ID 和构建日志。
3. 只更新 app 和 18182 custom；不运行整个 Compose 栈的 down/up。

只读检查：

```bash
cd /home/ubuntu/RAG_Weknora
docker inspect WeKnora-app WeKnora-docreader --format '{{.Name}} {{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{end}}'
docker exec WeKnora-postgres sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "select parse_status,summary_status,count(*) from knowledges where deleted_at is null group by 1,2 order by 1,2;"'
docker exec WeKnora-postgres sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "select * from schema_migrations;"'
```

以下命令会重建 app，必须在授权窗口执行。将 `ARCHIVAL_APP_IMAGE` 替换为 release.json 中实际镜像：

```bash
export ARCHIVAL_APP_IMAGE=weknora-app:archival-citations-实际时间
export ARCHIVAL_CITATIONS_ENABLED=true
docker compose -p weknora_v080 -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml -f deploy/archival-citations/compose.yaml up -d --no-deps app
docker logs --since 5m WeKnora-app
curl -fsS http://127.0.0.1:18081/health
```

custom 切换前核对 `frontend-before.json` 的网络、端口、环境、挂载；若与下列已核对参数不同，不盲用命令。保留旧容器便于一键恢复；不要在没有释放 18182 时重复创建：

```bash
docker stop WeKnora-frontend-custom-preview
docker rename WeKnora-frontend-custom-preview WeKnora-frontend-custom-before-citations
docker run -d --name WeKnora-frontend-custom-preview --restart unless-stopped --network weknora_v080_WeKnora-network -p 18182:80 -e DEFAULT_LOCALE=zh-CN -e APP_HOST=app -e APP_PORT=8080 -e APP_SCHEME=http -e MAX_FILE_SIZE_MB=50 -e MAX_SKILL_BUNDLE_SIZE_MB=256 archive-intelligence-ui:archival-citations-实际时间
curl -fsS http://127.0.0.1:18182/login >/dev/null
```

登录后创建独立样例库。使用带可信页 anchors 的外部分块包及对应 PDF 原件，普通问答和智能体各问一次：段落、列表、跨页表格及多文件；目录不得包含仅检索未引用资料。脚注复制原文与记录 SHA 一致；物理页定位后可继续上下浏览。再验证原件换版保留旧引用、删除原件／分块、撤销共享权限、刷新历史对话、快速切换与关闭重开。无页码首页回退不算“精确定位通过”。不要重跑正式库的历史 OCR 或切块。

## 关闭功能与回退

仅关闭：设 `ARCHIVAL_CITATIONS_ENABLED=false`，使用相同 Compose 命令只重建 app（同样需要重启授权）。

二进制／前端回退保留新增证据，不执行 down migration、不重置 Git：

```bash
export ARCHIVAL_APP_IMAGE=weknora-app:archival-rollback-对应发布时间
export ARCHIVAL_CITATIONS_ENABLED=false
docker compose -p weknora_v080 -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml -f deploy/archival-citations/compose.yaml up -d --no-deps app
docker stop WeKnora-frontend-custom-preview
docker rename WeKnora-frontend-custom-preview WeKnora-frontend-custom-citations-stopped
docker rename WeKnora-frontend-custom-before-citations WeKnora-frontend-custom-preview
docker start WeKnora-frontend-custom-preview
```

如需二次发布，先检查上述保留容器名称，不能覆盖或随意删除已有回退资产。回退镜像保留旧二进制并携带新增迁移源文件，使旧迁移器仍能识别 94 版本；不是直接用缺少新迁移文件的旧镜像盲目启动。以启动日志验证，禁止 `force` 版本或删除证据表来掩盖错误。
