CREATE TABLE IF NOT EXISTS auth_sessions (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    run_id TEXT NOT NULL,
    session_id_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    session_information BLOB NOT NULL,
    PRIMARY KEY (tenant, realm, session_id_hash)
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_run_id ON auth_sessions(run_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at); 