ALTER TABLE task_pending_ops ADD COLUMN IF NOT EXISTS available_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_task_pending_ready ON task_pending_ops(task_type,scope,scope_id,enqueued_at,id);
UPDATE task_pending_ops SET available_at=GREATEST(enqueued_at+INTERVAL '30 minutes',NOW()) WHERE task_type='wiki:ingest' AND fail_count>0 AND available_at IS NULL;

CREATE TABLE IF NOT EXISTS enrichment_retry_settings (id INTEGER PRIMARY KEY CHECK(id=1),enabled_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
INSERT INTO enrichment_retry_settings(id) VALUES(1) ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS enrichment_recoveries (
 id BIGSERIAL PRIMARY KEY, source_dead_letter_id BIGINT NOT NULL, logical_key TEXT NOT NULL UNIQUE,
 tenant_id BIGINT NOT NULL, knowledge_base_id VARCHAR(64) NOT NULL, knowledge_id VARCHAR(64) NOT NULL,
 attempt INTEGER NOT NULL, task_type VARCHAR(64) NOT NULL, payload JSONB NOT NULL,
 state VARCHAR(20) NOT NULL DEFAULT 'ready',attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 2),
 dispatch_no INTEGER NOT NULL DEFAULT 0,available_at TIMESTAMPTZ NOT NULL,queued_at TIMESTAMPTZ,lease_until TIMESTAMPTZ,
 slot_reserved BOOLEAN NOT NULL DEFAULT FALSE,last_error TEXT NOT NULL DEFAULT '',updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_enrichment_recovery_ready ON enrichment_recoveries(state,available_at,id);
CREATE TABLE IF NOT EXISTS enrichment_recovery_sources(dead_letter_id BIGINT PRIMARY KEY,recovery_id BIGINT NOT NULL REFERENCES enrichment_recoveries(id));
