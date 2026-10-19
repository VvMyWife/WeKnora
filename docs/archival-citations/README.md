# 档案级回答引用：实施与验收

服务器源码：`/home/ubuntu/RAG_Weknora`；定制前端为其独立 `frontend-custom/` 仓库。两个仓库均沿用 `frontend-custom` 分支。所有新增提交作者与提交者为 `ljc <ljc@local>`。不推送，不重写历史，不覆盖原版 frontend。

## 实施前基线（2026-10-19）

- 后端：`077ee148`，应用 `weknora-app:graph-output-12288-20261016`，镜像 ID `f113613c6a5ae07f9752548a3b5ca3712e34b37fcb0da9627a7bd6985d49944f`。
- custom 前端：`8201116`，运行镜像 ID `28763a77429697caf8a3caa631cc7c24ab34e61617e0db518a6b72f694a3d00c`。
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

新增测试入口将在对应阶段创建；未创建前不能把命令存在当成测试通过。上线需要重新核对活动任务并确认后端重启窗口；不能沿用其他功能的重启授权。回退保留引用记录，不执行删证据的 down migration。
