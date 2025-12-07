-- migrations/011_add_claims_to_client_sessions.up.sql

ALTER TABLE client_sessions ADD COLUMN claims TEXT;

