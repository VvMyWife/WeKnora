#!/bin/bash
# Candidate-only GPU rerank using existing cached weights. No CPU fallback.
set -euo pipefail
cd "$(dirname "$0")/.."
gateway=$(docker network inspect -f '{{(index .IPAM.Config 0).Gateway}}' ljc-upstream-validate)
private=${LJC_ACCEPTANCE_DIR:-/home/ubuntu/weknora-fork-migration/20261020-upstream}
model=${LJC_ACCEPTANCE_RERANK_MODEL:-/home/ubuntu/RAG/runtime/models/BAAI/bge-reranker-v2-m3}
python=${LJC_ACCEPTANCE_MODEL_PYTHON:-/home/ubuntu/RAG/runtime/eval-venv/bin/python}
: "${LJC_ACCEPTANCE_GPU_UUID:?Set the explicitly selected GPU UUID}"
free=$(nvidia-smi --id="$LJC_ACCEPTANCE_GPU_UUID" --query-gpu=memory.free --format=csv,noheader,nounits)
test "$free" -ge 4096 || { echo "Selected GPU has less than 4GiB free; refusing to displace existing services";exit 1; }
test -r "$model/config.json"
test -x "$python"
mkdir -p "$private"
exec 9>"$private/gpu-rerank.lock"
flock -n 9 || { echo "Candidate GPU rerank already running";exit 75; }
if curl -fsS --max-time 3 "http://$gateway:18085/health" 2>/dev/null | "$python" -c 'import json,sys;x=json.load(sys.stdin);assert x["device"].startswith("cuda") and x["model"]=="bge-reranker-v2-m3"' 2>/dev/null; then
 echo "Existing candidate GPU rerank healthy at $gateway:18085";exit 0
fi
printf '%s\n' "$$" > "$private/gpu-rerank.pid"
export CUDA_VISIBLE_DEVICES="$LJC_ACCEPTANCE_GPU_UUID" RERANK_DEVICE=cuda
export OMP_NUM_THREADS=4 MKL_NUM_THREADS=4 HF_HUB_OFFLINE=1 TRANSFORMERS_OFFLINE=1 RERANK_MODEL="$model"
exec "$python" -m uvicorn adapters.rerank.server:app --host "$gateway" --port 18085
