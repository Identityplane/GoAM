-- migrations/001_create_users.sql

CREATE TABLE IF NOT EXISTS users (
    -- Unique UUID for the user
    id VARCHAR(36) PRIMARY KEY,

    -- Organization Context
    tenant VARCHAR(255) NOT NULL,
    realm VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,

    -- User status
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Identity Information
    display_name VARCHAR(255),
    given_name VARCHAR(255),
    family_name VARCHAR(255),

    -- Profile Information
    profile_picture_uri VARCHAR(1024),

    -- Additional contact information
    email VARCHAR(255),
    phone VARCHAR(50),
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,

    -- Login Information
    login_identifier VARCHAR(255),

    -- Locale
    locale VARCHAR(10),

    -- Authentication credentials
    password_credential TEXT,
    webauthn_credential TEXT,
    mfa_credential TEXT,

    -- Credential lock status
    password_locked BOOLEAN DEFAULT FALSE,
    webauthn_locked BOOLEAN DEFAULT FALSE,
    mfa_locked BOOLEAN DEFAULT FALSE,

    -- Failed login attempts
    failed_login_attempts_password INTEGER DEFAULT 0,
    failed_login_attempts_webauthn INTEGER DEFAULT 0,
    failed_login_attempts_mfa INTEGER DEFAULT 0,

    -- User roles and groups (stored as JSON)
    roles JSONB DEFAULT '[]'::jsonb,
    groups JSONB DEFAULT '[]'::jsonb,
    entitlements JSONB DEFAULT '[]'::jsonb,
    consent JSONB DEFAULT '[]'::jsonb,

    -- Extensibility
    attributes JSONB DEFAULT '{}'::jsonb,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,

    -- Federation
    federated_idp VARCHAR(255),
    federated_id VARCHAR(255),

    -- Devices (stored as JSON)
    trusted_devices JSONB,

    -- Constraints
    CONSTRAINT unique_username_per_tenant_realm UNIQUE (tenant, realm, username)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_realm ON users(tenant, realm);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_federated_id ON users(federated_id);
CREATE INDEX IF NOT EXISTS idx_users_login_identifier ON users(login_identifier);

