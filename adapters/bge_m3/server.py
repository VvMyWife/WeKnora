import os
from typing import Any
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from .input_limits import input_counts

MODEL_PATH = os.environ.get("BGE_M3_MODEL", "/home/ubuntu/RAG/runtime/models/BAAI/bge-m3")
DEVICE = os.environ.get("BGE_DEVICE", "cuda")
app = FastAPI(title="WeKnora BGE-M3 Adapter", version="0.1.0")
model = SentenceTransformer(MODEL_PATH, device=DEVICE)

class EmbeddingRequest(BaseModel):
    input: str | list[str]
    model: str = "bge-m3"
    encoding_format: str | None = None

@app.get("/health")
def health():
    return {"status": "ok", "model": "bge-m3", "device": str(model.device)}

@app.get("/v1/models")
def models():
    return {"object": "list", "data": [{"id": "bge-m3", "object": "model", "owned_by": "BAAI"}]}

@app.post("/v1/embeddings")
def embeddings(req: EmbeddingRequest):
    texts = [req.input] if isinstance(req.input, str) else req.input
    if not texts or any(not isinstance(x, str) for x in texts):
        raise HTTPException(400, "input must be a string or list of strings")
    try:
        limits = input_counts(model, texts)
    except ValueError as exc:
        raise HTTPException(400, str(exc)) from exc
    if limits['over_limit_indices']:
        raise HTTPException(400, {'code': 'EMBEDDING_INPUT_TOO_LONG', **limits})
    vectors = model.encode(texts, normalize_embeddings=True, convert_to_numpy=True, batch_size=8, show_progress_bar=False)
    return {"object": "list", "data": [{"object": "embedding", "embedding": v.tolist(), "index": i} for i, v in enumerate(vectors)], "model": "bge-m3", "usage": {"prompt_tokens": 0, "total_tokens": 0}}

@app.post('/v1/token-counts')
def token_counts(req: EmbeddingRequest):
    texts = [req.input] if isinstance(req.input, str) else req.input
    try:
        return input_counts(model, texts)
    except ValueError as exc:
        raise HTTPException(400, str(exc)) from exc
