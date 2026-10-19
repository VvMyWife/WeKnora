CREATE TABLE IF NOT EXISTS external_chunk_batches (
 id VARCHAR(36) PRIMARY KEY, tenant_id BIGINT NOT NULL, knowledge_base_id VARCHAR(36) NOT NULL,
 manifest_hash VARCHAR(64) NOT NULL, manifest_json TEXT NOT NULL, package_path TEXT NOT NULL,
 status VARCHAR(24) NOT NULL, created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_external_batches_scope ON external_chunk_batches(tenant_id,knowledge_base_id);
CREATE TABLE IF NOT EXISTS external_chunk_records (
 id VARCHAR(64) PRIMARY KEY, tenant_id BIGINT NOT NULL, knowledge_base_id VARCHAR(36) NOT NULL,
 batch_id VARCHAR(36) NOT NULL, external_document_id TEXT NOT NULL, knowledge_id VARCHAR(36) NOT NULL,
 relative_path TEXT NOT NULL, content_hash VARCHAR(64) NOT NULL, payload TEXT NOT NULL,
 status VARCHAR(24) NOT NULL, stage VARCHAR(40) NOT NULL, error TEXT NOT NULL, updated_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_external_records_knowledge ON external_chunk_records(tenant_id,knowledge_id);
CREATE TABLE IF NOT EXISTS external_chunk_memberships (
 batch_id VARCHAR(36) NOT NULL, record_id VARCHAR(64) NOT NULL, PRIMARY KEY(batch_id,record_id)
);
