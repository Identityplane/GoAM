CREATE TABLE IF NOT EXISTS applications (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    confidential BOOLEAN NOT NULL DEFAULT true,
    consent_required BOOLEAN NOT NULL DEFAULT false,
    description TEXT,
    allowed_scopes TEXT NOT NULL,
    allowed_flows TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant, realm, client_id)
); 