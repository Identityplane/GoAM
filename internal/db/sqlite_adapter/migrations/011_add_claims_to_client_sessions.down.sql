-- migrations/011_add_claims_to_client_sessions.down.sql

ALTER TABLE client_sessions DROP COLUMN claims;

