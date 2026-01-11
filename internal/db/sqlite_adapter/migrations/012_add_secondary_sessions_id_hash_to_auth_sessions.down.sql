DROP INDEX IF EXISTS idx_auth_sessions_secondary_session_id_hash;

ALTER TABLE auth_sessions DROP COLUMN secondary_session_id_hash;


