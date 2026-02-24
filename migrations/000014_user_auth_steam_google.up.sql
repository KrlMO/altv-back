CREATE TABLE al_user_auth (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR(20) NOT NULL, -- 'local', 'steam'
    provider_uid VARCHAR(100) NOT NULL, -- email o steamId
    password_hash TEXT, -- SOLO para local
    created_at TIMESTAMP DEFAULT NOW(),
    created_by BIGINT,
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by BIGINT,
    is_verified BOOLEAN DEFAULT FALSE,

    UNIQUE(provider, provider_uid),
    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id)
);
