CREATE TABLE IF NOT EXISTS message_citation_manifests (
 tenant_id INTEGER NOT NULL, session_id TEXT NOT NULL, message_id TEXT NOT NULL,
 answer_sha256 TEXT NOT NULL, payload TEXT NOT NULL, created_at DATETIME NOT NULL,
 PRIMARY KEY(tenant_id,session_id,message_id)
);
CREATE TABLE IF NOT EXISTS message_citation_evidence (
 id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, session_id TEXT NOT NULL,
 message_id TEXT NOT NULL, knowledge_tenant_id INTEGER NOT NULL, knowledge_id TEXT NOT NULL,
 chunk_id TEXT NOT NULL, summary_json TEXT NOT NULL, snapshot TEXT NOT NULL, created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_message_citation_scope ON message_citation_evidence(tenant_id,session_id,message_id);
