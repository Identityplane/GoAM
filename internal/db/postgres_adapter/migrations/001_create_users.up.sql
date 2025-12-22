-- migrations/001_create_users.sql

CREATE TABLE IF NOT EXISTS users (
    -- Unique UUID for the user
    id VARCHAR(36) NOT NULL,

    -- Organization Context
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,

    -- User status
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    PRIMARY KEY (id, tenant, realm)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_realm ON users(tenant, realm);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

