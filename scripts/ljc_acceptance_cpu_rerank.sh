#!/bin/bash
# Optional real-model acceptance. Existing cached weights, CPU only, bridge bind.
set -euo pipefail
cd "$(dirname "$0")/.."
gateway=$(docker network inspect -f '{{(index .IPAM.Config 0).Gateway}}' ljc-upstream-validate)
private=${LJC_ACCEPTANCE_DIR:-/home/ubuntu/weknora-fork-migration/20261020-upstream}
model=${LJC_ACCEPTANCE_RERANK_MODEL:-/home/ubuntu/RAG/runtime/models/BAAI/bge-reranker-v2-m3}
python=${LJC_ACCEPTANCE_MODEL_PYTHON:-/home/ubuntu/RAG/runtime/eval-venv/bin/python}
test -r "$model/config.json"
test -x "$python"
mkdir -p "$private"
exec 9>"$private/cpu-rerank.lock"
flock -n 9 || { echo "Candidate CPU rerank already running";exit 75; }
if curl -fsS --max-time 3 "http://$gateway:18083/health" 2>/dev/null | "$python" -c 'import json,sys; x=json.load(sys.stdin); assert x["device"]=="cpu" and x["model"]=="bge-reranker-v2-m3"' 2>/dev/null; then
 echo "Existing candidate CPU rerank is healthy at $gateway:18083";exit 0
fi
printf '%s\n' "$$" > "$private/cpu-rerank.pid"
export CUDA_VISIBLE_DEVICES= RERANK_DEVICE=cpu OMP_NUM_THREADS=4 MKL_NUM_THREADS=4
export HF_HUB_OFFLINE=1 TRANSFORMERS_OFFLINE=1 RERANK_MODEL="$model"
exec "$python" -m uvicorn adapters.rerank.server:app --host "$gateway" --port 18083
