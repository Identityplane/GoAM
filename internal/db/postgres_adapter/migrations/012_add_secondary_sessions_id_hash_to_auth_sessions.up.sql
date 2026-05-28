ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS secondary_session_id_hash VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_secondary_session_id_hash ON auth_sessions(secondary_session_id_hash);
