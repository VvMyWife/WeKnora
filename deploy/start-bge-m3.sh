#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p logs
if [ -f logs/bge-m3.pid ] && kill -0 "$(cat logs/bge-m3.pid)" 2>/dev/null; then
  echo "BGE-M3 already running: $(cat logs/bge-m3.pid)"; exit 0
fi
CUDA_VISIBLE_DEVICES=3 BGE_DEVICE=cuda BGE_M3_MODEL=/home/ubuntu/RAG/runtime/models/BAAI/bge-m3   nohup /home/ubuntu/RAG/runtime/eval-venv/bin/python -m uvicorn adapters.bge_m3.server:app   --host 0.0.0.0 --port 18082 > logs/bge-m3.log 2>&1 &
echo $! > logs/bge-m3.pid
sleep 2
curl -fsS http://127.0.0.1:18082/health
