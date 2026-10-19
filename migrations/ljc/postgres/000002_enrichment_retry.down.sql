DROP TABLE IF EXISTS enrichment_recovery_sources;
DROP TABLE IF EXISTS enrichment_recoveries;
DROP TABLE IF EXISTS enrichment_retry_settings;
DROP INDEX IF EXISTS idx_task_pending_ready;
ALTER TABLE task_pending_ops DROP COLUMN IF EXISTS available_at;
