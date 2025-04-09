-- migrations/001_create_users.sql

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,                          -- UUID or other unique ID
  username TEXT UNIQUE NOT NULL,                -- login name (LDAP: uid)
  password_hash TEXT NOT NULL,                  -- bcrypt or argon2 hash

  display_name TEXT,                            -- full name
  given_name TEXT,                              -- LDAP: givenName
  family_name TEXT,                             -- LDAP: sn
  email TEXT,                            -- LDAP: mail
  phone TEXT,                                   -- LDAP: telephoneNumber

  email_verified BOOLEAN DEFAULT FALSE,
  phone_verified BOOLEAN DEFAULT FALSE,

  roles TEXT[],                                 -- array of roles
  groups TEXT[],                                -- optional group tags
  attributes TEXT[],                            -- custom key-values

  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  last_login_at TIMESTAMP,

  -- Federation
  federated_idp TEXT,                           -- e.g. "google", "azuread"
  federated_id TEXT,                            -- IDP user ID

  -- Devices (linked elsewhere, but track here too)
  trusted_devices TEXT[],

  -- Login security
  failed_login_attempts INTEGER DEFAULT 0,
  last_failed_login_at TIMESTAMP,
  account_locked BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_federated ON users(federated_idp, federated_id);