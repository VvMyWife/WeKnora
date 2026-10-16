# WeKnora embedding 服务

接口 `http://10.20.0.15:18082/v1/embeddings`，模型名 `bge-m3`，输出 1024 维。代码位于本项目 `adapters/bge_m3/server.py`；复用 `/home/ubuntu/RAG/runtime/` 下已有的 Python 环境和模型文件，不修改旧 RAG 项目源码，不重新下载权重。

2026-10-16 排查时 18082 未监听，旧日志末尾为 Uvicorn 正常退出。此前只有 nohup + PID 文件方式，未发现 systemd 服务：服务器重启后不会自动恢复，旧 PID 还可能被其他进程复用。现有证据不支持把连接拒绝归因于显存不足。

## 首次安装与启动

```bash
cd /home/ubuntu/RAG_Weknora
nvidia-smi --query-gpu=index,name,memory.used,memory.total --format=csv
test -x /home/ubuntu/RAG/runtime/eval-venv/bin/python
test -r /home/ubuntu/RAG/runtime/models/BAAI/bge-m3/config.json
mkdir -p logs
sudo install -m 0644 deploy/embedding/weknora-bge-m3.service /etc/systemd/system/weknora-bge-m3.service
sudo systemd-analyze verify /etc/systemd/system/weknora-bge-m3.service
sudo systemctl daemon-reload
sudo systemctl enable --now weknora-bge-m3.service
bash deploy/start-bge-m3.sh
```

固定使用 GPU 3，保持原有部署习惯；先确认容量。开机自动启动、意外退出自动重启；手动 systemctl stop 不会触发自动重启。只加载本地权重，HF 默认国内镜像且离线，不触发下载。

## 真实向量化验收

不能只检查 health：

```bash
curl -fsS http://127.0.0.1:18082/health
curl -fsS http://127.0.0.1:18082/v1/embeddings -H 'Content-Type: application/json' -d '{"model":"bge-m3","input":["档案映射验收","独立向量化样例"]}' | /home/ubuntu/RAG/runtime/eval-venv/bin/python -c 'import sys,json,math; rows=json.load(sys.stdin)["data"]; assert len(rows)==2; assert all(len(x["embedding"])==1024 and all(math.isfinite(v) for v in x["embedding"]) for x in rows); print("embedding OK: 2 x 1024")'
sudo systemctl is-enabled weknora-bge-m3.service
sudo systemctl status weknora-bge-m3.service --no-pager
tail -50 logs/bge-m3.log
```

还需在独立知识库上传小 MD，确认 WeKnora 的请求成功、向量化阶段成功且有分块。不得擅自恢复用户已暂停的正式库，不能重处理整库来验收。

## 停止、重启与回退

```bash
cd /home/ubuntu/RAG_Weknora
bash deploy/stop-bge-m3.sh
sudo systemctl restart weknora-bge-m3.service
bash deploy/start-bge-m3.sh
# 仅在明确决定停用 embedding 时执行：
sudo systemctl disable --now weknora-bge-m3.service
```

应用日志仍为 `logs/bge-m3.log`，系统级原因可用 `journalctl -u weknora-bge-m3.service` 查看。无需修改 WeKnora 已有模型 ID、base URL 或向量维度。

## 2026-10-16 验收记录

服务已安装到 `/etc/systemd/system/weknora-bge-m3.service`，为 enabled / active。GPU 3 加载后约使用 2707 MiB；真实请求返回 2 条 1024 维向量。独立档案映射样例库的两份 MD 重解析后向量化均成功，应用容器请求在 `logs/bge-m3.log` 中返回 200。正式库保持暂停，没有重处理整库。自动重启策略已配置，未通过重启整台服务器或故意杀死模型进程进行破坏性验收。
