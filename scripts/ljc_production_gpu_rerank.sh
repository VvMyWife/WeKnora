#!/bin/bash
# Production GPU service: cached weights only, never evict another GPU worker.
set -euo pipefail
cd /home/ubuntu/RAG_Weknora-upstream-20261020
export CUDA_VISIBLE_DEVICES=GPU-918824a7-2144-fe49-5183-7646ab974946
free=$(nvidia-smi --id="$CUDA_VISIBLE_DEVICES" --query-gpu=memory.free --format=csv,noheader,nounits)
test "$free" -ge 4096 || { echo 'Selected GPU has less than4GiB free'; exit 1; }
export RERANK_DEVICE=cuda HF_HUB_OFFLINE=1 TRANSFORMERS_OFFLINE=1 OMP_NUM_THREADS=4 MKL_NUM_THREADS=4
export RERANK_MODEL=/home/ubuntu/RAG/runtime/models/BAAI/bge-reranker-v2-m3
test -r "$RERANK_MODEL/config.json"
exec /home/ubuntu/RAG/runtime/eval-venv/bin/python -m uvicorn adapters.rerank.server:app --host 10.20.0.15 --port 18083
