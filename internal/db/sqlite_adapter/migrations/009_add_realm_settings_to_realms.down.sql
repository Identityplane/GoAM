-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
-- This is a simplified approach - in production you might want to use a more sophisticated migration strategy

-- Create new table with the desired schema
CREATE TABLE realms_new (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    realm_name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    PRIMARY KEY (tenant, realm)
);

-- Copy data from old table to new table
INSERT INTO realms_new SELECT tenant, realm, realm_name, base_url FROM realms;

-- Drop old table
DROP TABLE realms;

-- Rename new table to original name
ALTER TABLE realms_new RENAME TO realms; 