# WeKnora 本地私有 Git 使用说明

本服务器保留两个独立仓库：`/home/ubuntu/RAG_Weknora` 记录 WeKnora 后端、部署和配置；`/home/ubuntu/RAG_Weknora/frontend-custom` 记录定制前端。两者都只在服务器本地保存，不自动推送。原始前端仍由既有切换方案保留；Git 提交不会自动改变运行中的容器。

## 每次修改后的检查与提交

先在对应目录操作。例如主仓库：

```bash
cd /home/ubuntu/RAG_Weknora
git status --short
git diff --check
git add -- <实际修改的源码或配置路径>
git diff --cached --name-only
git diff --cached --check
git commit -m "fix: 简要说明本次修改"
git log -3 --oneline
```

定制前端使用同样流程，但先运行 `cd /home/ubuntu/RAG_Weknora/frontend-custom`。不要使用 `git add -A` 或 `git add .` 将未知文件一并加入暂存区。提交前确认暂存区没有 `.env`、个人数据、备份、模型文件、`dist/` 或 `node_modules/`。

## 验证最近的记录与文件

```bash
git -C /home/ubuntu/RAG_Weknora log -5 --oneline
git -C /home/ubuntu/RAG_Weknora status --short
git -C /home/ubuntu/RAG_Weknora/frontend-custom log -5 --oneline
git -C /home/ubuntu/RAG_Weknora/frontend-custom status --short
git -C /home/ubuntu/RAG_Weknora show --stat --oneline HEAD
git -C /home/ubuntu/RAG_Weknora/frontend-custom show --stat --oneline HEAD
```

## 出问题时回退

先记下要撤销的提交 ID，并检查是否有未提交修改。下例仅展示本地源码回退，执行前把 `<提交ID>` 换成实际值：

```bash
cd /home/ubuntu/RAG_Weknora
git status --short
git show --stat <提交ID>
git revert <提交ID>
```

前端提交则在 `frontend-custom` 目录执行相同命令。`git revert` 会产生一条新的、可再次撤销的提交；如果当前有相同文件的未提交修改，应先处理这些修改，不要强行覆盖。代码回退后，按现有部署流程构建并切换容器，最后用 `git log` 和页面/接口测试验证。不要对数据库卷使用 `docker compose down -v`。

## 误提交敏感内容时

立即停止推送和共享仓库；即使随后删除文件，旧提交仍会保留内容。应先轮换相关凭据，再评估是否需要重写私有历史。不要把密码写在提交信息中。
