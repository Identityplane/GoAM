-- Add debug_allowed column to flows table
ALTER TABLE flows ADD COLUMN debug_allowed BOOLEAN NOT NULL DEFAULT FALSE; 