CREATE TABLE user_attributes (
    id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    index_value VARCHAR(255),
    type VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (user_id, tenant, realm) REFERENCES users(id, tenant, realm) ON DELETE CASCADE,
    UNIQUE (tenant, realm, type, index_value),
    PRIMARY KEY (id, tenant, realm)
);

CREATE INDEX idx_user_attributes_tenant_realm_user_id ON user_attributes(tenant, realm, user_id);
CREATE INDEX idx_user_attributes_tenant_realm_type ON user_attributes(tenant, realm, type);
CREATE INDEX idx_user_attributes_tenant_realm_type_index ON user_attributes(tenant, realm, type, index_value);
CREATE INDEX idx_user_attributes_value_gin ON user_attributes USING GIN (value); 