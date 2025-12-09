-- ============================================
-- ORGANIZATIONS
-- ============================================

CREATE TABLE al_organization (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    logo_path TEXT,
    banner_path TEXT,
    created_by INT NOT NULL,
    active BOOLEAN NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_org_owner FOREIGN KEY (created_by) REFERENCES al_user(id)
);

-- ============================================
-- ORGANIZATION ROLES (OWNER, ADMIN, MODERATOR, STAFF, ANALYST, CREATOR)
-- ============================================

CREATE TABLE al_organization_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO al_organization_role (name) VALUES
('OWNER'),
('ADMIN'),
('MODERATOR'),
('STAFF'),
('ANALYST'),
('CREATOR');

-- ============================================
-- USER → ORGANIZATION → ROLE
-- ============================================

CREATE TABLE al_user_organization_role (
    user_id INT NOT NULL,
    organization_id INT NOT NULL,
    role_id INT NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by INT NOT NULL,
    active BOOLEAN NULL DEFAULT true,

    PRIMARY KEY (user_id, organization_id, role_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (organization_id) REFERENCES al_organization(id),
    FOREIGN KEY (role_id) REFERENCES al_organization_role(id),
    FOREIGN KEY (assigned_by) REFERENCES al_user(id)
);
