-- migrations/001_create_users.sql

CREATE TABLE IF NOT EXISTS users (

    -- Unique UUID for the user
    id TEXT PRIMARY KEY,

    -- Organization Context
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    username TEXT NOT NULL,

    -- User status
    status TEXT NOT NULL DEFAULT 'active',

    -- Identity Information
    display_name TEXT,
    given_name TEXT,
    family_name TEXT,

    -- Profile Information
    profile_picture_uri TEXT,

    -- Additional contact information
    email TEXT,
    phone TEXT,
    email_verified BOOLEAN DEFAULT false,
    phone_verified BOOLEAN DEFAULT false,

    -- Login Information
    login_identifier TEXT,

    -- Locale
    locale TEXT,

    -- Authentication credentials (stored as encrypted JSON)
    password_credential TEXT,
    webauthn_credential TEXT,
    mfa_credential TEXT,

    -- Credential lock status
    password_locked BOOLEAN DEFAULT false,
    webauthn_locked BOOLEAN DEFAULT false,
    mfa_locked BOOLEAN DEFAULT false,

    -- Failed login attempts
    failed_login_attempts_password INTEGER DEFAULT 0,
    failed_login_attempts_webauthn INTEGER DEFAULT 0,
    failed_login_attempts_mfa INTEGER DEFAULT 0,

    -- User roles and groups (stored as JSON)
    roles TEXT DEFAULT '[]',
    groups TEXT DEFAULT '[]',
    entitlements TEXT DEFAULT '[]',
    consent TEXT DEFAULT '[]',

    -- Extensibility
    attributes TEXT DEFAULT '{}',

    -- Audit
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    last_login_at TEXT,

    -- Federation
    federated_idp TEXT,
    federated_id TEXT,

    -- Devices (stored as JSON)
    trusted_devices TEXT,

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

