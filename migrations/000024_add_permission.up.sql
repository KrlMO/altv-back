CREATE TABLE al_permission (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code VARCHAR(80) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE TABLE al_membership_permission (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    membership_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    UNIQUE (membership_id, permission_id),
    FOREIGN KEY (membership_id) REFERENCES al_membership(id),
    FOREIGN KEY (permission_id) REFERENCES al_permission(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE TABLE al_user_membership (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    membership_id BIGINT NOT NULL,
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (membership_id) REFERENCES al_membership(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE TABLE al_team_membership (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id BIGINT NOT NULL,
    membership_id BIGINT NOT NULL,
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    FOREIGN KEY (team_id) REFERENCES al_team(id),
    FOREIGN KEY (membership_id) REFERENCES al_membership(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE TABLE al_organization_membership (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    membership_id BIGINT NOT NULL,
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    FOREIGN KEY (organization_id) REFERENCES al_organization(id),
    FOREIGN KEY (membership_id) REFERENCES al_membership(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);