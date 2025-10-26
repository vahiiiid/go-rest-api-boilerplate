-- Migration: create_users_table (rollback)
-- Description: Drops users table and associated indexes

BEGIN;

DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;

COMMIT;

