CREATE TABLE al_user_verification_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_auth_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    FOREIGN KEY (user_auth_id) REFERENCES al_user_auth(id)
);

CREATE UNIQUE INDEX ux_user_active_verification
ON al_user_verification_tokens(user_auth_id)
WHERE is_active = true;