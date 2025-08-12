-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
-- This is a simplified approach - in production you might want to use a more sophisticated migration strategy

-- Create new table with the desired schema
CREATE TABLE flows_new (
    tenant TEXT NOT NULL,
    realm TEXT NOT NULL,
    id TEXT NOT NULL,
    route TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    definition_yaml TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant, realm, id),
    UNIQUE (tenant, realm, route)
);

-- Copy data from old table to new table
INSERT INTO flows_new SELECT tenant, realm, id, route, active, definition_yaml, created_at, updated_at FROM flows;

-- Drop old table
DROP TABLE flows;

-- Rename new table to original name
ALTER TABLE flows_new RENAME TO flows; 