-- Original-file mapping is separate from parsing and retrieval IDs.
CREATE TABLE IF NOT EXISTS archive_mapping_batches (
 id VARCHAR(36) PRIMARY KEY, tenant_id BIGINT NOT NULL,
 knowledge_base_id VARCHAR(36) NOT NULL, parsed_root TEXT NOT NULL,
 created_by TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_batches_scope ON archive_mapping_batches(tenant_id, knowledge_base_id);
CREATE TABLE IF NOT EXISTS archive_mapping_items (
 id VARCHAR(36) PRIMARY KEY, batch_id VARCHAR(36) NOT NULL,
 relative_path TEXT NOT NULL, file_size BIGINT NOT NULL, candidates TEXT,
 knowledge_id VARCHAR(36), status TEXT NOT NULL, error TEXT,
 version_id VARCHAR(36), updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_items_batch ON archive_mapping_items(batch_id);
CREATE TABLE IF NOT EXISTS archive_mapping_versions (
 id VARCHAR(36) PRIMARY KEY, tenant_id BIGINT NOT NULL,
 knowledge_id VARCHAR(36) NOT NULL, file_name TEXT NOT NULL,
 relative_path TEXT NOT NULL, file_type TEXT NOT NULL, file_size BIGINT NOT NULL,
 sha256 VARCHAR(64) NOT NULL, storage_path TEXT NOT NULL,
 created_by TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_versions_knowledge ON archive_mapping_versions(tenant_id,knowledge_id);
CREATE TABLE IF NOT EXISTS archive_mapping_links (
 knowledge_id VARCHAR(36) PRIMARY KEY, tenant_id BIGINT NOT NULL,
 version_id VARCHAR(36) NOT NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
