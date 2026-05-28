-- Increase id column length in flows and users tables
ALTER TABLE flows ALTER COLUMN id TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN id TYPE VARCHAR(255);
ALTER TABLE user_attributes ALTER COLUMN user_id TYPE VARCHAR(255);
ALTER TABLE auth_sessions ALTER COLUMN run_id TYPE VARCHAR(255);
ALTER TABLE client_sessions ALTER COLUMN user_id TYPE VARCHAR(255);
