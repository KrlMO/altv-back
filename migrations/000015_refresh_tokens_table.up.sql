CREATE TABLE al_refresh_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id BIGINT NOT NULL,

    token_hash TEXT NOT NULL,

    device_info TEXT,
    ip_address TEXT,

    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_refresh_user
        FOREIGN KEY (user_id)
        REFERENCES al_user(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_refresh_token_hash
ON al_refresh_tokens(token_hash);

CREATE INDEX idx_refresh_user_id
ON al_refresh_tokens(user_id);

CREATE INDEX idx_refresh_user_revoked
ON al_refresh_tokens(user_id, revoked);