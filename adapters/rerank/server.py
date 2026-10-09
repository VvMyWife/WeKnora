import os, uuid
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import CrossEncoder

MODEL_PATH = os.environ.get("RERANK_MODEL", "/home/ubuntu/RAG/runtime/models/BAAI/bge-reranker-v2-m3")
DEVICE = os.environ.get("RERANK_DEVICE", "cuda")
app = FastAPI(title="WeKnora BGE Rerank Adapter", version="0.1.0")
model = CrossEncoder(MODEL_PATH, device=DEVICE, max_length=512)

class RerankRequest(BaseModel):
    model: str = "bge-reranker-v2-m3"
    query: str
    documents: list[str]

@app.get("/health")
def health():
    return {"status": "ok", "model": "bge-reranker-v2-m3", "device": str(model.device)}

@app.post("/rerank")
def rerank(req: RerankRequest):
    if not req.documents:
        return {"id": str(uuid.uuid4()), "model": req.model, "usage": {"total_tokens": 0}, "results": []}
    pairs = [[req.query, d] for d in req.documents]
    scores = model.predict(pairs, batch_size=8, show_progress_bar=False)
    results = [{"index": i, "document": {"text": req.documents[i]}, "relevance_score": float(scores[i])} for i in range(len(req.documents))]
    results.sort(key=lambda x: x["relevance_score"], reverse=True)
    return {"id": str(uuid.uuid4()), "model": req.model, "usage": {"total_tokens": 0}, "results": results}
