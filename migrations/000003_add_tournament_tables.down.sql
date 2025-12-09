-- ============================================
-- GAME
-- ============================================
CREATE TABLE al_game (
    
)



-- ============================================
-- TOURNAMENTS
-- ============================================

CREATE TABLE al_tournament (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'DRAFT',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (organization_id) REFERENCES al_organization(id)
);

-- ============================================
-- TOURNAMENT ROLES
-- ============================================

CREATE TABLE al_tournament_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO al_tournament_role (name) VALUES
('TOURNAMENT_OWNER'),
('TOURNAMENT_ADMIN'),
('MATCH_ADMIN'),
('REFEREE'),
('SUPPORT'),
('SCORE_KEEPER');

-- ============================================
-- USER → TOURNAMENT → ROLE
-- ============================================

CREATE TABLE al_user_tournament_role (
    user_id INT NOT NULL,
    tournament_id INT NOT NULL,
    role_id INT NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, tournament_id, role_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
    FOREIGN KEY (role_id) REFERENCES al_tournament_role(id)
);
