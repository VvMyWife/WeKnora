# WeKnora 内网免密钥问答 API

本功能新增独立的 `/api/openai/v1` 入口。调用方不登录、不申请密钥；SDK 的 `not-required` 仅为占位字符串。原网页和管理 API 继续使用原认证。匿名调用者共享管理员开放的资料范围，不提供个人私密会话；知道回答 ID 的获准内网调用者可读取该回答的引用证据。

## 代码与运行基线

- 生产运行基线：`f45137d3fd784e076336fcc11b50d17ba5fe85c0`，未将其他工作树的未上线改动混入。
- 实施分支：`ljc/api-access`。
- 服务器工作树：`/home/ubuntu/RAG_Weknora/runtime/worktrees/api-access`。
- 隔离验收数据、构建日志和依赖都放在 `/home/ubuntu/RAG_Weknora/runtime/`，不提交 Git。
- 现有网页端口 18182、旧前端 18180 保留。生产后端切换需要维护窗口；隔离验收成功不等于已上线。

## 接口契约

| 方法和相对路径 | 用途 |
| --- | --- |
| `GET /models` | 当前可执行的开放问答别名 |
| `POST /chat/completions` | 同步或 SSE 问答 |
| `GET /responses/{response_id}/citations` | 本入口产生的回答的引用清单 |
| `GET /responses/{response_id}/citations/{citation_id}` | 证据快照与定位信息 |

只支持 `model`、`messages`、`stream`、`max_tokens`。`messages` 为非空、交替的纯文本 user/assistant 历史，从 user 开始并以 user 结束，最多 100 条，正文合计最多 256 KiB，请求体最多 512 KiB。不支持 system、工具调用、图片、附件、温度、结构化输出、`stream_options`、`max_completion_tokens` 等参数，传入返回 400。第一版没有上传接口。

每个请求创建独立内部会话，完整提供所需历史；不会自动关联上一次请求。旧问答对分别保存轮次，最后一轮只生成一次。历史回答中的引用不自动算作本轮证据。

同步结果包含 `id/object/created/model/choices/usage`，正文是 Markdown，引用位置显示 `[1]`、`[2]`。扩展顶层字段：

- `citations`：`id/number/title/source_kind/location/content_revision/content_sha256/archive_version_id/archive_file_name/detail_url`。
- `citation_status`：`state/unbound_count`；`needs_review` 等状态不表示事实已经被认证。
- `usage`：本轮可观测的所有模型调用用量之和，不仅是最终输出；无法统计时为 null，不伪造数值。

`location` 保留底层可信页码或缺失状态，不从文本页、逻辑页推算物理页码。`detail_url` 为相对地址，详情中的 `snapshot` 是引用当时的证据内容。来源删除或撤回开放范围后，详情不再可读。配置中的 profile 任一字段变化会使其旧回答的引用查询失效；这是一项保守撤回策略。没有任意知识库原文查询入口，内部网页消息 ID 无法在此查询。

SSE 为 Chat Completions chunk，`delta.content` 是即时草稿；结束前扩展 chunk 返回校验后的 `final_content/citations/citation_status/usage`，客户端必须用 `final_content` 替换草稿，以保证最终编号和重复引用一致。流结束为 `data: [DONE]`。生成错误在开始流之前返回 HTTP 错误；已开始则发 `error` 事件并结束，不发送成功的最终正文。

错误码：400 参数不兼容，403 来源或执行权限不足，404 未开放别名/不可查询回答或证据，429 限流并携带 `Retry-After: 60`，502 生成失败，503 配置/共享限制器/持久化不可用，504 超时。断开或超时会取消当前执行，不自动重做整轮答案。

## 服务端配置

环境变量 `WEKNORA_ANONYMOUS_API_CONFIG` 指向 JSON 文件。每个请求重新读取；建议挂载父目录并原子替换文件。默认未配置时禁用；启用但白名单为空时仅允许真实 loopback。

```json
{
  "enabled": true,
  "tenant_id": 10000,
  "user_id": "专用Viewer用户UUID",
  "allowed_cidrs": ["10.40.0.34/32"],
  "trusted_proxy_cidrs": ["172.25.0.2/32"],
  "global_concurrency": 4,
  "ip_concurrency": 2,
  "requests_per_minute": 10,
  "timeout_seconds": 300,
  "profiles": [{
    "id": "policy-qa",
    "agent_id": "builtin-quick-answer",
    "knowledge_base_ids": ["8575ca20-5e7e-48c4-9fc0-786d979fdf5d"],
    "max_tokens": 2048
  }]
}
```

上面的地址是这次核实的客户端/代理示例，不代表已经确认全校 CIDR。管理员需加入实际校园/VPN网段。只信任明确配置的代理地址，从转发链右侧剥离可信代理，不能信任调用方自报 IP。代理容器重建换 IP 后需更新。部署前检查 Nginx 的转发规则，并使用真实客户端测试。

专用执行用户必须活跃、不是系统管理员、不能跨所有租户，且在目标租户拥有 active Viewer 成员资格。允许知识库还必须在该身份可读取范围内。请求复制智能体配置并强制只读、只用开放知识库、关闭记忆、MCP、skills、浏览器、网页抓取、沙箱和业务修改工具；内部最多 6 轮，每次模型调用最多 120 秒，整轮最多 300 秒。`max_tokens` 不得超过 profile 上限，合法配置范围 1–12288，并继续受模型窗口约束。

Redis 在所有应用实例间原子限制生成并发与 IP 频次。默认全体 4、每 IP 2、每 IP 60 秒计数窗口 10 次；超过立即 429，不排无界队列。共享 NAT 会共享 IP 限额。Redis 故障拒绝放行；进程意外终止的占用最迟在请求期限加 30 秒后失效。纯 SQLite/lite 且没有 Redis 的部署不能启用该入口。

## 调用步骤

curl 无需 Authorization：

```bash
BASE=http://10.20.0.15:18182/api/openai/v1
curl -fsS "$BASE/models"
curl -fsS "$BASE/chat/completions" -H 'Content-Type: application/json' \
  -d '{"model":"policy-qa","messages":[{"role":"user","content":"第三级灾难恢复的数据备份要求是什么？"}],"max_tokens":1024}'
curl -N "$BASE/chat/completions" -H 'Content-Type: application/json' \
  -d '{"model":"policy-qa","messages":[{"role":"user","content":"请说明依据"}],"stream":true}'
curl -fsS "$BASE/responses/替换为返回的回答ID/citations"
curl -fsS "$BASE/responses/替换为回答ID/citations/替换为引用ID"
```

Python 依赖安装到持久目录，项目配置使用清华镜像。Windows PowerShell：

```powershell
cd E:\desktop\文档RAG
python -m venv runtime/anonymous-api-sdk
@'
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
timeout = 60
'@ | Set-Content -Encoding ascii runtime/anonymous-api-sdk/pip.ini
runtime/anonymous-api-sdk/Scripts/python.exe -m pip install openai==2.54.0 httpx==0.28.1
```

```python
from openai import OpenAI
client = OpenAI(base_url="http://10.20.0.15:18182/api/openai/v1",
                api_key="not-required", timeout=310, max_retries=0)
messages = [{"role": "user", "content": "第三级灾难恢复的数据备份要求是什么？"}]
r = client.chat.completions.create(model="policy-qa", messages=messages)
print(r.choices[0].message.content)
print(r.model_extra.get("citations", []))
messages += [{"role": "assistant", "content": r.choices[0].message.content},
             {"role": "user", "content": "具体依据是什么？"}]
r = client.chat.completions.create(model="policy-qa", messages=messages)
```

服务器提供完整示例 `examples/anonymous-api/client.py`，支持 `--stream`；不是完整的 OpenAI API 替代服务，只兼容上述接口和参数。SDK 示例关闭自动重试，避免 429/故障时重新生成整轮。

## 构建与自动测试（服务器 ubuntu 账号）

```bash
cd /home/ubuntu/RAG_Weknora/runtime/worktrees/api-access
docker run --rm -v "$PWD:/src" \
  -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build -w /src \
  -e GOMAXPROCS=4 -e GOPROXY=https://goproxy.cn,direct \
  weknora-builder:upstream-20261020 sh -c '
  go test ./internal/database ./internal/handler/session ./internal/anonymousapi ./internal/archivalcitations ./internal/models/api ./internal/logger ./internal/middleware ./internal/router -count=1 &&
  go test -race ./internal/handler/session ./internal/anonymousapi ./internal/logger -count=1 &&
  go build -buildvcs=false -o /src/runtime/WeKnora.anonymous-api ./cmd/server'
```

依赖缓存复用持久目录；容器基础镜像复用生产实际镜像，不额外拉取或升级生产依赖。服务器 SDK：

```bash
/home/ubuntu/anaconda3/bin/python -m venv examples/anonymous-api/.sdk-venv
cp examples/anonymous-api/pip.conf examples/anonymous-api/.sdk-venv/pip.conf
examples/anonymous-api/.sdk-venv/bin/python -m pip install -r examples/anonymous-api/requirements.txt
python3 deploy/anonymous-api/isolation.py
python3 deploy/anonymous-api/fixture.py
examples/anonymous-api/.sdk-venv/bin/python deploy/anonymous-api/acceptance.py \
  --out /home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance/sdk-server.json
docker run --rm --network weknora-api-test-net -v "$PWD:/src" \
  -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod \
  -v weknora-go-build-cache:/root/.cache/go-build -w /src -e GOMAXPROCS=4 \
  -e ANON_TEST_REDIS=api-redis:6379 weknora-builder:upstream-20261020 \
  go test ./internal/anonymousapi -run TestAnonymousRealRedis -v -count=1
```

真实 Redis 测试会清理其专属 anonymous-api 测试键，**只能指向独立 Redis**。隔离实例为新建 PostgreSQL/Redis/存储，不共享生产队列，不重新解析正式库；只复用模型服务执行少量样例调用。样例文件独立生成，没有 Wiki/图谱。

从真实内网客户端验收：

```bash
python3 deploy/anonymous-api/acceptance_proxy.py --client-ip 10.40.0.34
```

该代理只暴露隔离实例的匿名路由，其他路径 404。电脑执行（先将 acceptance.py 复制到电脑）：

```powershell
runtime/anonymous-api-sdk/Scripts/python.exe -X utf8 runtime/worktrees/api-access/deploy/anonymous-api/acceptance.py --base http://10.20.0.15:28482/api/openai/v1 --out runtime/anonymous-api-windows-acceptance.json
```

`sdk-server.json` 和 Windows 报告只在所有检查通过后生成；不要把失败伪造为通过。检查覆盖同步、流式、行内编号、证据快照、历史、用量、私有消息隔离、不支持参数和管理入口隔离。另需核验故障、取消、超时、白名单外拒绝、限流和撤回范围，详见 `acceptance-results.md`。

## 维护窗口部署与回退

只重建 app，不重建数据库、模型、docreader 或前端。先只准备可审阅产物：

```bash
python3 deploy/anonymous-api/release.py prepare --cidr 10.40.0.34/32
# 记录输出的 prepared 绝对目录；检查 config.proposed.json 与 compose.api.json。
# 如需开放其他同学，先核实其校园/VPN CIDR，再重新 prepare，逐个 --cidr 添加。
```

进入已获批准的维护窗口后执行，替换目录：

```bash
RELEASE=/home/ubuntu/RAG_Weknora/runtime/anonymous-api-releases/替换为准备目录
python3 deploy/anonymous-api/release.py apply --release-dir "$RELEASE" --maintenance-approved
curl -fsS http://127.0.0.1:18081/health
docker inspect WeKnora-app --format '{{.Image}}'
docker logs --since 10m WeKnora-app 2>&1 | grep anonymous_api
```

脚本重新核验运行镜像、候选镜像、代理地址和 LJC 版本4；生成并验证新鲜 `pg_dump -Fc` 备份，事务性增加 migration5 的回答归属表，建立禁用交互登录的 Viewer，先禁用新入口再切换 app，健康通过后才发布开放配置。旧镜像与旧 Compose 保存在 release 目录。配置/备份可能含服务机密，目录 700、文件 600，不提交或对外分发。

从获准内网客户端对 **18182** 运行示例，并复验网页登录、已有会话和问答。检查实际镜像、无认证 curl、流式、历史和脚注详情均成功后才报告上线。网页问题不通过不得以健康检查替代。

回退仅撤回新入口和应用镜像，保留新增表、证据和业务数据，不做数据库恢复覆盖：

```bash
python3 deploy/anonymous-api/release.py rollback --release-dir "$RELEASE" --maintenance-approved
curl -fsS http://127.0.0.1:18081/health
docker inspect WeKnora-app --format '{{.Image}}'
```

中断后先读 `$RELEASE/state.json`，检查容器镜像、健康状态、迁移表版本、配置 enabled 值及备份归档。只有 `prepared` 允许 apply；遇到 `schema_applied/cutover_started` 必须检查并回退或明确恢复，脚本拒绝盲目重复插入身份和切换。不要重复创建正式知识库或重新解析文档。

验收结束后可停止隔离服务（保留数据便于复验）：

```bash
docker stop weknora-api-test-proxy weknora-api-acceptance weknora-api-test-redis weknora-api-test-pg
```

后续临时附件另开 `ljc/` 分支设计归属与访问凭据，本版不开放匿名上传。
