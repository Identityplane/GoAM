CREATE TABLE IF NOT EXISTS flows (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    id TEXT NOT NULL,
    route TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    definition_yaml TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant, realm, id),
    UNIQUE (tenant, realm, route)
); 