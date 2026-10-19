CREATE TABLE IF NOT EXISTS message_citation_manifests (
 tenant_id BIGINT NOT NULL, session_id VARCHAR(36) NOT NULL, message_id VARCHAR(36) NOT NULL,
 answer_sha256 VARCHAR(64) NOT NULL, payload TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL,
 PRIMARY KEY(tenant_id,session_id,message_id)
);
CREATE TABLE IF NOT EXISTS message_citation_evidence (
 id VARCHAR(64) PRIMARY KEY, tenant_id BIGINT NOT NULL, session_id VARCHAR(36) NOT NULL,
 message_id VARCHAR(36) NOT NULL, knowledge_tenant_id BIGINT NOT NULL, knowledge_id VARCHAR(36) NOT NULL,
 chunk_id VARCHAR(36) NOT NULL, summary_json TEXT NOT NULL, snapshot TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_message_citation_scope ON message_citation_evidence(tenant_id,session_id,message_id);
