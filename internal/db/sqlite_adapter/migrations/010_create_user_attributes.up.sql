CREATE TABLE user_attributes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    index_value TEXT NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (tenant, realm, type, index_value)
);

CREATE INDEX idx_user_attributes_tenant_realm_user_id ON user_attributes(tenant, realm, user_id);
CREATE INDEX idx_user_attributes_tenant_realm_type ON user_attributes(tenant, realm, type);
CREATE INDEX idx_user_attributes_tenant_realm_type_index ON user_attributes(tenant, realm, type, index_value); 