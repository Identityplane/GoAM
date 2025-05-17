-- migrations/003_create_flows.sql

CREATE TABLE IF NOT EXISTS flows (
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    id VARCHAR(36) NOT NULL,
    route VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    definition_yaml TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant, realm, id),
    UNIQUE (tenant, realm, route)
); 