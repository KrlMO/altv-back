DROP INDEX IF EXISTS idx_user_deleted_at;

ALTER TABLE al_user
    DROP COLUMN IF EXISTS deleted_at;
