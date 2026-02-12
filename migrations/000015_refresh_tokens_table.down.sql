-- ============================
-- DOWN: Drop al_refresh_tokens
-- ============================

DROP INDEX IF EXISTS idx_refresh_user_revoked;
DROP INDEX IF EXISTS idx_refresh_user_id;
DROP INDEX IF EXISTS idx_refresh_token_hash;

DROP TABLE IF EXISTS al_refresh_tokens;