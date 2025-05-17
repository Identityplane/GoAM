-- migrations/004_create_applications.sql

CREATE TABLE IF NOT EXISTS applications (
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    confidential BOOLEAN NOT NULL,
    consent_required BOOLEAN NOT NULL,
    description TEXT,
    allowed_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_grants JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_authentication_flows JSONB NOT NULL DEFAULT '[]'::jsonb,
    access_token_lifetime INTEGER NOT NULL,
    refresh_token_lifetime INTEGER NOT NULL,
    id_token_lifetime INTEGER NOT NULL,
    access_token_type VARCHAR(50) NOT NULL,
    access_token_algorithm VARCHAR(50),
    access_token_mapping JSONB,
    id_token_algorithm VARCHAR(50),
    id_token_mapping JSONB,
    redirect_uris JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant, realm, client_id)
); 