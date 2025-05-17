-- migrations/006_create_signing_keys.sql

CREATE TABLE IF NOT EXISTS signing_keys (
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    kid VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    algorithm VARCHAR(50) NOT NULL,
    implementation VARCHAR(50) NOT NULL,
    signing_key_material TEXT NOT NULL,
    public_key_jwk TEXT NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    disabled TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (tenant, realm, kid)
);

CREATE INDEX IF NOT EXISTS idx_signing_keys_tenant_realm ON signing_keys(tenant, realm);
CREATE INDEX IF NOT EXISTS idx_signing_keys_active ON signing_keys(active); 