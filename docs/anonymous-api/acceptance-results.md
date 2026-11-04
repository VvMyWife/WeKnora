# 独立验收记录（2026-11-04）

本记录区分隔离验收与生产切换。生产18182的免密钥API现已上线并通过25项客户端检查；完整浏览器问答回归仍待人工确认。正式知识库没有重新解析。应用代码源提交`86fde4b02`，镜像修订`db1eb36be`包含SQLite迁移编号整理；后续仅增加交付脚本与文档。

## 隔离阶段已验证

| 检查 | 结果与范围 |
| --- | --- |
| 八组 Go 回归 | database、session、anonymousapi、archivalcitations、models/api、logger、middleware、router 全部通过；SQLite 新库/旧库升级及重复迁移通过 |
| Go race | session、anonymousapi、logger 全部通过；覆盖异步事件、超时及生成失败 |
| 实际 Redis | 独立 Redis 7.0，两个客户端共享并发/IP/频次额度，取消拒绝，测试通过 |
| 实际 PostgreSQL | 独立 ParadeDB PostgreSQL17，增量表创建、回答归属与证据写入成功；重启后仍能查询 |
| 官方 Python SDK | OpenAI 2.54.0 + httpx 0.28.1；服务器及 Windows 真实内网客户端均通过 12 项检查 |
| 同步样例 | 最新 Windows 样例约 2.45 秒，返回 1 个有效引用及实际 usage；这是样例耗时，不是并发压测结论 |
| 流式样例 | SDK 读到 delta、最终校正版 final_content、引用和结束标记 |
| 多轮 | 本轮正确读到只存在于请求历史中的测试代号；没有重新执行历史问题 |
| 匿名引用范围 | 不带认证头读取本入口证据成功；私有网页 ID、未发布 scope 不能借用 |
| 真实来源拒绝 | 隔离 Docker 网络内白名单外的另一个真实 peer 返回403；伪造转发头不替换不可信 peer |
| 真实超时/断开 | 配置1秒超时返回504；断开后执行上下文收到取消，Redis占用归零 |
| 撤回与重启 | 改变 profile 后旧引用404；恢复配置并重启隔离 app 后原映射可查询 |
| 日志 | 抽查匿名问题/测试代号没有出现在候选日志，记录请求元数据；SQL隐私保护另有回归测试 |
| 失败输出 | Go 测试注入模型失败：同步502，已开始的SSE返回错误并DONE，不产生成功final_content |
| 发布准备 | 准备阶段只读核验并生成可审阅配置；获批准后执行apply，见下方生产记录 |

独立样例包含两块人工构造的验收资料，模型真实调用现有9B、embedding与rerank服务。返回的引用可能命中派生摘要，示例中保留 `source_kind=derived` 和 `location.status=missing`，没有伪造PDF页码。引言未被绑定时继续显示“引用待核验”；成功返回一条引用不代表整篇逐句覆盖。

## 可复现证据

服务器运行数据目录：`/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance/`。

- `sdk-server.json`：12项SDK检查、回答、用量、引用和原文快照。
- `edge-checks.json`：8项真实网络/超时/取消/撤回/重启检查。
- `runtime.json`、`fixture.json`：含隔离服务机密，仅服务器本地600权限，不分享、不提交。
- 构建和测试日志在 `/home/ubuntu/RAG_Weknora/runtime/api-release-verified.log`、`api-observability-build.log`、`api-race.log`、`api-database-regression.log`。
- Windows 证据：`E:\desktop\文档RAG\runtime\anonymous-api-windows-acceptance.json`。

README 提供重新执行各组检查的完整命令。不要对生产 Redis 执行独立 Redis 测试。

## 当前尚未完成

1. 浏览器完整问答回归：自动化安全检查停止后需人工复验。
2. 其他校园/VPN网段若要调用，须核实CIDR再加入；当前已开放10.40.0.0/22和10.20.0.0/22。
3. 未执行生产回退演练，保留旧镜像、旧Compose、备份和精确回退命令。

## 生产切换与验收补充（2026-11-04）

用户明确授权“可以切换”后，生产完成新鲜备份、归档校验、增量表迁移、专用Viewer创建与app重建。API已开放，来源限定已核实的两个网段。只重建app，PostgreSQL、Redis、Neo4j、docreader和两个前端保持原容器；应用二进制SHA256与隔离验收产物相同。

Windows对正式18182接口的25项检查通过；同步样例约15.08秒、2条有效引用，流式最终正文/证据和多轮历史均通过。不带认证头的HTTP访问正常，SDK占位字符串不作为凭据。某次模型输出没有有效引用，接口保留needs_review与未绑定数量，这是诚实状态，并非承诺逐句覆盖。

生产证据文件：`/home/ubuntu/RAG_Weknora/runtime/anonymous-api-production-acceptance.json` 和对应Windows同名runtime文件；运行镜像、容器对比、备份与配置见release目录的`postflight.json`。这些运行数据不提交Git。

网页完整问答回归仍待人工复验。自动化因为无法可靠确认浏览器URL而被安全检查停止，未绕过工具限制。页面HTTP200与前端未改动不表述为完整网页问答通过。未执行生产回退演练，旧镜像/Compose和回退命令已保留。
