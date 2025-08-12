-- Add realm_settings column to realms table
ALTER TABLE realms ADD COLUMN realm_settings JSONB NOT NULL DEFAULT '{}'; 