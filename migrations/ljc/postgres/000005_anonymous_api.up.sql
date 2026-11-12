CREATE TABLE IF NOT EXISTS ljc_anonymous_api_responses (
 id varchar(64) PRIMARY KEY,
 tenant_id bigint NOT NULL,
 user_id varchar(36) NOT NULL,
 profile_id varchar(64) NOT NULL,
 profile_sha256 varchar(64) NOT NULL,
 session_id varchar(36) NOT NULL,
 message_id varchar(36) NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ljc_anon_response_message ON ljc_anonymous_api_responses(tenant_id,session_id,message_id);
