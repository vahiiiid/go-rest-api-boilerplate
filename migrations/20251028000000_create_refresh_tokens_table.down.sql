-- Migration: create_refresh_tokens_table (rollback)
-- Description: Drops refresh_tokens table

BEGIN;

DROP TABLE IF EXISTS refresh_tokens;

COMMIT;
