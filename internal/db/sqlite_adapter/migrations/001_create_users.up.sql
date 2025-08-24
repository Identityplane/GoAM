-- migrations/001_create_users.sql

CREATE TABLE IF NOT EXISTS users (
    -- Unique UUID for the user
    id TEXT PRIMARY KEY,

    -- Organization Context
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,

    -- User status
    status TEXT NOT NULL DEFAULT 'active',

    -- Audit
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    last_login_at TEXT,

    -- Constraints
    CONSTRAINT unique_user_per_tenant_realm UNIQUE (tenant, realm, id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_realm ON users(tenant, realm);
CREATE INDEX IF NOT EXISTS idx_users_id ON users(id);

