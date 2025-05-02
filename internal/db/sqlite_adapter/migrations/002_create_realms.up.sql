-- migrations/002_create_realms.sql

-- Create realms table
CREATE TABLE IF NOT EXISTS realms (
    -- Organization Context
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    realm_name TEXT NOT NULL,
    base_url TEXT,

    -- Constraints
    PRIMARY KEY (tenant, realm)
);
