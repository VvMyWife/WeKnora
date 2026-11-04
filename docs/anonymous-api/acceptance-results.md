# 独立验收记录（2026-11-04）

本记录描述隔离验收，**不表示生产 18182 已上线**。正式应用仍为原运行镜像，正式知识库没有重新解析。候选二进制源提交 `86fde4b02`，交付脚本与文档随后单独提交。

## 已验证

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
| 发布准备 | 只读核验当前镜像、Compose、目标库和前端代理，生成可审阅部署配置；没有执行apply |

独立样例包含两块人工构造的验收资料，模型真实调用现有9B、embedding与rerank服务。返回的引用可能命中派生摘要，示例中保留 `source_kind=derived` 和 `location.status=missing`，没有伪造PDF页码。引言未被绑定时继续显示“引用待核验”；成功返回一条引用不代表整篇逐句覆盖。

## 可复现证据

服务器运行数据目录：`/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance/`。

- `sdk-server.json`：12项SDK检查、回答、用量、引用和原文快照。
- `edge-checks.json`：8项真实网络/超时/取消/撤回/重启检查。
- `runtime.json`、`fixture.json`：含隔离服务机密，仅服务器本地600权限，不分享、不提交。
- 构建和测试日志在 `/home/ubuntu/RAG_Weknora/runtime/api-release-verified.log`、`api-observability-build.log`、`api-race.log`、`api-database-regression.log`。
- Windows 证据：`E:\desktop\文档RAG\runtime\anonymous-api-windows-acceptance.json`。

README 提供重新执行各组检查的完整命令。不要对生产 Redis 执行独立 Redis 测试。

## 上线前仍需完成

1. 确认维护窗口，核实需要开放给其他同学的校园/VPN CIDR。已实测客户端仅10.40.0.34/32。
2. 使用release脚本apply；脚本在切换前重新备份、校验运行状态，保留旧镜像和配置。
3. 从真实客户端对18182复验同步、流式、历史、引用，并复验已有网页问答。仅健康通过不能代替这些检查。

生产切换、正式网页回归及生产回退演练未执行。正式知识库资料不会因隔离验收改变。
