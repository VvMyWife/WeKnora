# 档案映射

服务器唯一源码：`/home/ubuntu/RAG_Weknora`。本仓库与 `frontend-custom/` 分别提交。原版 `frontend/` 不修改。原件映射不生成知识条目、不改变 knowledge/chunk ID，不调用模型。

## 分阶段提交

1. 独立数据表、相对路径匹配与版本幂等测试。
2. 批次清单、原件存储、版本管理和日志。
3. 原件预览/下载解析，以及复制、删除生命周期。
4. custom 前端映射管理、刷新恢复、每批并发 2。
5. custom 预览/引用关联原件，保留解析文本入口。
6. 完整验收后固定部署镜像；每阶段先测试，再在对应仓库提交。

## 手动测试命令

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited go test ./internal/archivemapping -count=1
git status --short
git diff --check
git log -6 --oneline
git -C frontend-custom status --short
git -C frontend-custom log -6 --oneline
```

新增表迁移只在部署新版后生效；部署前的测试使用独立内存数据库，不修改正在处理的正式库。回退程序镜像时保留映射表与文件，不执行删除数据表的 down migration。
