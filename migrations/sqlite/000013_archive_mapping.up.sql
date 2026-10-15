CREATE TABLE IF NOT EXISTS archive_mapping_batches (
 id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL,
 knowledge_base_id TEXT NOT NULL, parsed_root TEXT NOT NULL,
 created_by TEXT NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_batches_scope ON archive_mapping_batches(tenant_id,knowledge_base_id);
CREATE TABLE IF NOT EXISTS archive_mapping_items (
 id TEXT PRIMARY KEY, batch_id TEXT NOT NULL, relative_path TEXT NOT NULL,
 file_size INTEGER NOT NULL, candidates TEXT, knowledge_id TEXT,
 status TEXT NOT NULL, error TEXT, version_id TEXT,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_items_batch ON archive_mapping_items(batch_id);
CREATE TABLE IF NOT EXISTS archive_mapping_versions (
 id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL, knowledge_id TEXT NOT NULL,
 file_name TEXT NOT NULL, relative_path TEXT NOT NULL, file_type TEXT NOT NULL,
 file_size INTEGER NOT NULL, sha256 TEXT NOT NULL, storage_path TEXT NOT NULL,
 created_by TEXT NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_archive_mapping_versions_knowledge ON archive_mapping_versions(tenant_id,knowledge_id);
CREATE TABLE IF NOT EXISTS archive_mapping_links (
 knowledge_id TEXT PRIMARY KEY, tenant_id INTEGER NOT NULL,
 version_id TEXT NOT NULL, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
