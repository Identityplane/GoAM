-- Drop indexes
DROP INDEX IF EXISTS idx_users_tenant_realm;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_federated_id;

-- Drop table
DROP TABLE IF EXISTS users; 