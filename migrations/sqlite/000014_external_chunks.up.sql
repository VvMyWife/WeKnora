CREATE TABLE IF NOT EXISTS external_chunk_batches (
 id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, knowledge_base_id TEXT NOT NULL,
 manifest_hash TEXT NOT NULL, manifest_json TEXT NOT NULL, package_path TEXT NOT NULL,
 status TEXT NOT NULL, created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_external_batches_scope ON external_chunk_batches(tenant_id,knowledge_base_id);
CREATE TABLE IF NOT EXISTS external_chunk_records (
 id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, knowledge_base_id TEXT NOT NULL,
 batch_id TEXT NOT NULL, external_document_id TEXT NOT NULL, knowledge_id TEXT NOT NULL,
 relative_path TEXT NOT NULL, content_hash TEXT NOT NULL, payload TEXT NOT NULL,
 status TEXT NOT NULL, stage TEXT NOT NULL, error TEXT NOT NULL, updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_external_records_knowledge ON external_chunk_records(tenant_id,knowledge_id);
CREATE TABLE IF NOT EXISTS external_chunk_memberships (
 batch_id TEXT NOT NULL, record_id TEXT NOT NULL, PRIMARY KEY(batch_id,record_id)
);
