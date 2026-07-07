ALTER TABLE al_user
    ADD COLUMN deleted_at TIMESTAMP NULL;

CREATE INDEX idx_user_deleted_at ON al_user (deleted_at) WHERE deleted_at IS NOT NULL;
