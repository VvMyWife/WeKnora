# WeKnora 官方主线同步 · 2026-11-25

权威目录为 `/home/ubuntu/RAG_Weknora`。目标是官方 `main@71c32af9273482e0dad013d0e3bea2f4d719516c`，包含 v0.9.1 和该发行版之后的聊天文件选择弹层定位修复。相对上一轮官方基线 `cbdcebab`，共有 174 个提交、548 个文件变化。阶段提交作者和提交者均为 `ljc <ljc@local>`。 [官方比较](https://github.com/Tencent/WeKnora/compare/cbdcebabf48c5ee2d6868b370326aa8341703f62...71c32af9273482e0dad013d0e3bea2f4d719516c) · [v0.9.1](https://github.com/Tencent/WeKnora/releases/tag/v0.9.1)

- `d552b6a42` 保留历史合并官方主线，适配外部分块、解析重试和引用流事件结算。
- `ca8d4a6d2` 将官方前端变化移植到 18182 的定制前端，保留档案智查品牌、原件预览、档案与队列界面。
- `0bc6a849e` 适配外部分块重试，沿用同一次解析编号，阻止过期任务把新任务标为失败。
- [官方提交清单](official-commits.tsv) · [官方变化文件](official-files.tsv)
- 候选后端源码为 `0bc6a849e`，DocReader 与 18182 前端源码为 `ca8d4a6d2`。生产运行状态另记于发布目录。

## 重跑源码验证与候选构建

在权威目录执行。状态文件为 0 表示对应命令成功。先检查既有容器、日志和账本，按失败阶段恢复。后端按 `candidate.sha` 构建，DocReader 和前端按 `frontend.sha` 构建。运行最后两条构建命令前，确认与 `frontend.sha` 的差异只涉及后端和验收记录。

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/upstream-integration/20261125
git status --short
git worktree list
git merge-base --is-ancestor 71c32af9273482e0dad013d0e3bea2f4d719516c HEAD
git diff --name-only "$(cat "$p/candidate.sha")" HEAD
docker ps -a --filter label=ljc.operation=integration-20261125
bash "$p/backend-tests.sh"
bash "$p/external-backend-test.sh"
bash "$p/docreader-test.sh"
bash "$p/frontend-test.sh"
bash "$p/anydoc.sh"
install -m 644 "$p/native/third_party/anydoc-go/lib/linux_amd64_gnu/libanydoc_go.a" third_party/anydoc-go/lib/linux_amd64_gnu/libanydoc_go.a
sha=$(cat "$p/candidate.sha")
bash "$p/candidate-backend.sh" "$sha"
ui_sha=$(cat "$p/frontend.sha")
bash "$p/candidate-docreader.sh" "$ui_sha"
bash "$p/candidate-ui.sh" "$ui_sha"
```

前端测试复用持久 npm 缓存。缓存依赖缺失时用下列命令补齐。后端复用 Go 缓存和已安装的 DuckDB 扩展。候选使用 GPUStack 的模型实例。隔离 BGE 适配服务将 token 计数交给同一模型的 tokenizer，过长输入会被拒绝，向量计算仍由 GPU 完成。

```bash
docker run --rm --pull never -v "$PWD:/src" -v weknora-npm-cache:/npm-cache -v ljc-weknora-20261113-ui-deps:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-node:24.16.0-npmmirror npm ci --no-audit --no-fund
```

## 隔离数据库与浏览器验收

生产库快照流式恢复到隔离容器。PG 账本从 136＋5 升至 137＋5，迁移重复执行两次，业务表指纹一致。后端测试包含 SQLite 官方与个人双账本。首次运行数据库副本验证前确认隔离库为空。副本已经恢复、仍是 136＋5 时，运行 `db-proof-resume.sh` 从迁移阶段继续。已通过的阶段核对状态与日志。

```bash
cat "$p/db-proof.status"
cat "$p/database.before.fingerprints" "$p/database.after.fingerprints"
cmp "$p/database.before.fingerprints" "$p/database.after.fingerprints"
docker compose -f "$p/compose.acceptance.json" config --quiet
docker start ljc-20261125-proof-pg
docker compose -f "$p/compose.acceptance.json" up -d --pull never
curl -fsS http://127.0.0.1:28281/health
curl -fsS -o /dev/null http://127.0.0.1:28282/
bash "$p/run-acceptance.sh"
docker compose -f "$p/compose.acceptance.json" stop
docker stop -t 120 ljc-20261125-proof-pg
```

浏览器使用验收账号登录 28282，依次查看档案智查首页、知识库、任务队列和资料处理进度。在聊天中验证引用卡显示来源知识库，原件预览能打开映射后的 PDF 。拥有可信物理页码的样本用于验证自动定位，普通 Markdown 样本用于验证映射后的 PDF 阅读器。正式 18182 上验证部署资源版本、匿名问答引用明细与管理接口鉴权。

## 正式切换与恢复

仅在验收通过且正式队列空闲时执行。正式切换脚本预检镜像与生产数据库，沿用生产配置挂载，合并官方智能体提示词的表格核对规则。切换涉及生产数据，按发布状态恢复，避免重放初始化步骤。

```bash
r=runtime/production/release-20261125
python3 "$r/deploy.py" prepare
nohup python3 -u "$r/deploy.py" cutover > "$r/cutover.log" 2>&1 </dev/null &
cat "$r/state.json"
curl -fsS http://127.0.0.1:18081/health
curl -fsS -o /dev/null http://127.0.0.1:18182/
WEKNORA_SMOKE_DB=weknora_release_20261125 python3 "$p/production-smoke.py"
```

回退前暂停写入并核对新版库中新增的业务数据。原库尚在时，使用 `compose.before.json` 恢复原镜像与原库配置。若原库已退役，先从唯一的 `runtime/production/database-backup/latest.dump` 恢复并核对账本和业务记录，再启动旧镜像。提示词回退文件在发布目录 `agent_system_prompt.before.yaml`。

```bash
install -m 664 "$r/agent_system_prompt.before.yaml" runtime/production/config/prompt_templates/agent_system_prompt.yaml
docker compose -p ljc-weknora-production -f "$r/compose.before.json" config --quiet
for service in docreader app frontend; do
  docker compose -p ljc-weknora-production -f "$r/compose.before.json" up -d --no-deps --pull never "$service" || break
done
curl -fsS -o /dev/null http://127.0.0.1:18182/
```
