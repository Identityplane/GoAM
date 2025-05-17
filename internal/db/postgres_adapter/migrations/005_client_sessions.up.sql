-- migrations/005_client_sessions.sql

CREATE TABLE IF NOT EXISTS client_sessions (
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    client_session_id VARCHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    grant_type VARCHAR(50) NOT NULL,
    access_token_hash VARCHAR(255),
    refresh_token_hash VARCHAR(255),
    auth_code_hash VARCHAR(255),
    user_id VARCHAR(36),
    scope TEXT DEFAULT '',
    login_session_state_json TEXT,
    code_challenge VARCHAR(255),
    code_challenge_method VARCHAR(50),
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expire TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (tenant, realm, client_session_id)
);

CREATE INDEX IF NOT EXISTS idx_client_sessions_access_token ON client_sessions(access_token_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_refresh_token ON client_sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_auth_code ON client_sessions(auth_code_hash);
CREATE INDEX IF NOT EXISTS idx_client_sessions_client_id ON client_sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_client_sessions_user_id ON client_sessions(user_id); 