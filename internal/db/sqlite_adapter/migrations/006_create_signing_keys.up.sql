CREATE TABLE IF NOT EXISTS signing_keys (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    kid TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    algorithm TEXT NOT NULL,
    implementation TEXT NOT NULL,
    signing_key_material TEXT NOT NULL,
    public_key_jwk TEXT NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    disabled TIMESTAMP,
    PRIMARY KEY (tenant, realm, kid)
);

CREATE INDEX IF NOT EXISTS idx_signing_keys_tenant_realm ON signing_keys(tenant, realm);
CREATE INDEX IF NOT EXISTS idx_signing_keys_active ON signing_keys(active); 