# WeKnora 知识库处理日志

唯一代码目录：`/home/ubuntu/RAG_Weknora`。本功能不改解析、判重、摘要或图谱任务，也不重启 WeKnora 应用；每 5 分钟从数据库和当前应用容器日志导出一次可读文件。

## 查看

```bash
cd /home/ubuntu/RAG_Weknora
cat data/processing-logs/INDEX.txt
less data/processing-logs/<知识库ID>/status.txt
tail -f data/processing-logs/<知识库ID>/events.log
```

`status.txt` 包含文档数量、解析/摘要状态、解析/后处理/Wiki/图谱等阶段计数、当前失败文档和最近一次处理尝试的失败步骤；每轮原子更新。软删除知识库标记“已删除”。`events.log` 只追加上线后采集到的上传/处理失败事件；没有新失败时文件可能尚不存在。`exporter.log` 保存定时导出程序本身的执行结果和错误。目录默认仅属主和同组可读，不应对外分享原始错误详情。

Docker 原始日志仍可用 `docker logs WeKnora-app` 查看，数据库 `knowledge_processing_spans` 仍是完整处理轨迹的权威来源。此处按**知识库**归档；上传流程目前没有可靠的持久化批次 ID，因此不虚构按批次归档。

## 自己测试与运行

```bash
cd /home/ubuntu/RAG_Weknora
python3 -m unittest discover -s scripts -p 'test_export_processing_logs.py' -v
python3 scripts/export_processing_logs.py --once
python3 scripts/install_processing_log_cron.py
cat data/processing-logs/INDEX.txt
crontab -l | grep export_processing_logs
```

选择 `INDEX.txt` 中的知识库 ID，打开对应 `status.txt`，将解析/摘要计数与网页进度面板核对。若你再次上传一份内容重复的文件，等下一次导出后在该知识库的 `events.log` 查找 `KnowledgeUploadFailure` 与 `duplicate_of`；处理阶段发生新失败时查找 `KnowledgeProcessFailure`。若失败发生在没有知识库 ID 的早期请求校验阶段，它仍只能在 Docker 原始日志中查到。

定时任务每 5 分钟运行一次，程序使用文件锁避免重叠；中断或数据库暂时不可用时返回错误，保留上一份报告，下次定时运行自动重试。手动检查故障：

```bash
tail -n 50 data/processing-logs/exporter.log
docker ps --format '{{.Names}} {{.Status}}' | grep -E 'WeKnora-app|WeKnora-postgres'
python3 scripts/export_processing_logs.py --once
```

停用定时任务：`python3 scripts/install_processing_log_cron.py --remove`。它仅删除带本项目标记的那一行，不会删除已生成的文件。
