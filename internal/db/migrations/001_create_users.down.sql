-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_tenant_realm;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_federated_idp;
DROP INDEX IF EXISTS idx_users_federated_id;

-- Drop table
DROP TABLE IF EXISTS users; 