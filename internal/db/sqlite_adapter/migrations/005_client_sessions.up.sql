CREATE TABLE IF NOT EXISTS client_sessions (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    client_session_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    grant_type TEXT NOT NULL,
    access_token_hash TEXT,
    refresh_token_hash TEXT,
    auth_code_hash TEXT,
    user_id TEXT,
    scope TEXT,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expire TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant, realm, client_session_id)
);

CREATE INDEX IF NOT EXISTS idx_client_sessions_access_token ON client_sessions(access_token_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_refresh_token ON client_sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_auth_code ON client_sessions(auth_code_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_client_id ON client_sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_client_sessions_user_id ON client_sessions(user_id); 