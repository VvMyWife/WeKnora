# WeKnora 功能开发、合并和 GitHub 交付

`ljc/custom` 是最新经过验证的交付分支。大改造在独立的 `ljc/功能名` 分支和工作目录开发；代码、回归及必要的隔离运行验收通过后再合入 custom，然后明确推送 GitHub。暂不配置 GitLab。

文件修改、commit、push、部署是四件不同的事。分支跟踪关系不会自动推送；custom 分支也不通过定时器或 hook 自动上传半成品。Git 中最新可用源码与服务器实际运行镜像分别记录。合并不会自动切换生产服务，有迁移或重启的发布仍需相应授权和验收。

## 1. 开始一个大的功能

下列示例中的功能名和目录需换成本次实际名称。创建前检查名称和路径是否已存在；中断后复用原工作目录，不重复创建。

```bash
cd /home/ubuntu/RAG_Weknora
git worktree list
git status --short --branch
git ls-remote --heads origin ljc/custom
git worktree add -b ljc/example-feature \
  /home/ubuntu/RAG_Weknora-example-feature ljc/custom
cd /home/ubuntu/RAG_Weknora-example-feature
git status --short --branch
```

不要在生产挂载的旧工作目录执行 checkout/reset/clean。当前生产挂载目录仍是 `/home/ubuntu/RAG_Weknora-upstream-20261020`。custom 的工作目录是 `/home/ubuntu/RAG_Weknora-upstream-20261023`，Git 共享仓库位于 `/home/ubuntu/RAG_Weknora`。

## 2. 分阶段开发和回归

每个阶段范围明确、测试通过、独立提交。作者和提交者均用 `ljc <ljc@local>`。国内镜像、持久依赖缓存及 anydoc 原生库复用方式见 `docs/upstream-sync-20261023/README.md`。

```bash
git diff --check
git status --short
# 精确添加本阶段文件，避免纳入密钥、数据、日志或构建产物：
# git add path/to/changed-file
# git -c user.name=ljc -c user.email=ljc@local commit -m '本阶段的具体变更'

# 沿用现有可重复测试脚本，在本次功能工作目录执行：
bash scripts/ljc_sync_20261023.sh backend
bash scripts/ljc_sync_20261023.sh custom-ui
```

前述命令包含后端全量相关包、Asynq 无限超时、anydoc、前端全量测试、类型检查和构建。若变更官方前端或 DocReader，还要运行对应 `official-ui`／`docreader` 阶段；涉及迁移时在独立副本证明，不能仅凭测试通过就迁移正式库。需要浏览器及真实接口验收的功能须完成相应隔离验收。

## 3. 合入 custom，再推送 GitHub

若 custom 在开发期间有其他更新，先在功能工作目录合并 custom 并重新验收；不能用 reset 或强推覆盖另一条分支。

```bash
cd /home/ubuntu/RAG_Weknora-example-feature
git merge --no-ff ljc/custom
# 有冲突时解决后重新运行对应回归，再提交。

cd /home/ubuntu/RAG_Weknora-upstream-20261023
git status --short --branch
git merge --ff-only ljc/example-feature
git log -3 --oneline
git ls-remote --heads origin ljc/custom
git push origin ljc/custom:ljc/custom
git ls-remote --heads origin ljc/custom
```

最后一次 `ls-remote` 的 SHA 应等于 `git rev-parse ljc/custom`。如果远程在此期间被其他人更新而拒绝推送，应获取远程更新，在功能分支整合和重测；禁止直接 force push。保留 `push.default=nothing`，确保每次都明确指定要推送的分支。

### 当前实际推送通道

服务器 origin 的 fetch 地址是 HTTPS，push 地址是 SSH；当前 Ubuntu 没有 GitHub 推送私钥。此次在服务器完成合并，再把 Git bundle 下载到 Windows，使用本机已有的 Git Credential Manager 登录凭据，经 HTTPS 正常推送。没有复制私钥、配置自动推送或关闭 SSH 主机验证。

若服务器尚未配置自己的 GitHub 凭据，可继续用以下方法。先在服务器生成 bundle：

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261023
git bundle create runtime/custom-delivery.bundle ljc/custom
```

在 Windows PowerShell 下载并推送。`github-push.git` 是本次已建立的持久裸仓库；后续复用它，不对现有项目工作目录执行 reset：

```powershell
scp ubuntu@10.20.0.15:/home/ubuntu/RAG_Weknora-upstream-20261023/runtime/custom-delivery.bundle E:\desktop\文档RAG\reports\custom-delivery.bundle
git -C E:\desktop\文档RAG\reports\generated\custom-delivery-20261026\github-push.git fetch E:\desktop\文档RAG\reports\custom-delivery.bundle refs/heads/ljc/custom:refs/heads/custom
git -C E:\desktop\文档RAG\reports\generated\custom-delivery-20261026\github-push.git push https://github.com/VvMyWife/WeKnora.git refs/heads/custom:refs/heads/ljc/custom
git ls-remote --heads https://github.com/VvMyWife/WeKnora.git ljc/custom
```

本机需要有该仓库的 GitHub 写权限和有效登录；不将 token 写进命令或仓库。智能体连接服务器仍默认 Paramiko 密码方式，以上 scp 是给人工操作的标准下载命令。

## 4. 本次档案功能整合

整合起点：custom `5a1f74c9c`（已含官方 bc1ade36），档案功能 `58500f746`。二者分别有 44、13 个独有提交。独立工作目录 `/home/ubuntu/RAG_Weknora-custom-integration-20261026`、分支 `ljc/merge-archive-into-custom-20261026`；采用正式 merge 保留双方历史，不 cherry-pick 丢失分支关系，不覆盖 custom。

合并没有文本冲突。合并后的最终测试、提交和 GitHub SHA 记录在本目录的 `acceptance.json`。此次 Git 合并不执行生产 migration，也不切换正式镜像；生产后端 f45137d3f、前端 e5df8a440 的既有档案版交付记录仍见 `docs/archive-document/README.md`。custom 另外包含此前已验收、尚未发布的官方133迁移等更新，不能把推送完成描述成生产部署完成。

## 5. 中断自检和恢复

```bash
cd /home/ubuntu/RAG_Weknora-custom-integration-20261026
git status --short --branch
git rev-parse -q --verify MERGE_HEAD || true
git diff --name-only --diff-filter=U
git worktree list
bash scripts/ljc_sync_20261023.sh status
tail -40 runtime/custom-integration-20261026/*.log
```

先确认同名测试容器或进程是否仍在运行、状态文件是否成功，再决定重跑。失败就继续修复并重测，custom 保持上一个已验证提交。合并后若尚未推送需要撤回，先在隔离目录核对恢复方案；已推送的共享历史使用明确 revert 提交，不改写分支历史。
