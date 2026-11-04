# 内网免密钥问答 API

基线：生产运行提交 f45137d3fd784e076336fcc11b50d17ba5fe85c0。分支 ljc/api-access，隔离工作树 /home/ubuntu/RAG_Weknora/runtime/worktrees/api-access。保留18182网页、18180旧前端和现有认证。

实施阶段：0契约；1受限匿名入口与资源范围；2同步/流式问答；3引用；4SDK验收和部署说明。各阶段独立测试和提交。未部署不能表述为已上线。

契约：/api/openai/v1/models、/chat/completions、/responses/{id}/citations、/responses/{id}/citations/{citation_id}。仅支持model/messages/stream/max_tokens，user/assistant纯文本历史，最后一条user。无密钥；SDK使用not-required占位。匿名调用者共享开放范围，不提供个人私密会话。

默认全局4并发、单IP2并发/每分钟10次、超时300秒。缺少网络配置仅允许loopback，客户端不能伪造转发头绕过。新增入口不能开放管理、浏览器、终端、MCP写入或任意文件查询。临时文件后续独立实施。
