ALTER TABLE auth_sessions ADD COLUMN secondary_session_id_hash TEXT;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_secondary_session_id_hash ON auth_sessions(secondary_session_id_hash);