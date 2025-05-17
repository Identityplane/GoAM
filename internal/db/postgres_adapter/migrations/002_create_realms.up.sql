-- migrations/002_create_realms.sql

-- Create realms table
CREATE TABLE IF NOT EXISTS realms (
    -- Organization Context
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    realm_name VARCHAR(255) NOT NULL,
    base_url VARCHAR(1024),

    -- Constraints
    PRIMARY KEY (tenant, realm)
); 