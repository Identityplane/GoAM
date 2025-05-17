-- migrations/007_auth_sessions.sql

CREATE TABLE IF NOT EXISTS auth_sessions (
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    run_id VARCHAR(36) NOT NULL,
    session_id_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    session_information BYTEA NOT NULL,
    PRIMARY KEY (tenant, realm, session_id_hash)
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_run_id ON auth_sessions(run_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at); 