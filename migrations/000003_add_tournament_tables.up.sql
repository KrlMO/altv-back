-- ============================================
-- GAME
-- ============================================

CREATE TABLE al_game (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT NULL,
    logo_path TEXT NULL,
    banner_path TEXT NULL,
    year_founded INT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    active BOOLEAN DEFAULT true,
    color_hex VARCHAR(7) NULL,
    popularity_rank INT NULL
);

-- ============================================
-- PLATFORM
-- ============================================

CREATE TABLE al_platform (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    logo_path TEXT NULL,
    active BOOLEAN DEFAULT true
);

-- ============================================
-- GENRE_E_SPORTS
-- ============================================

CREATE TABLE al_genre_e_sports (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL
);

-- ============================================
-- TOURNAMENTS
-- ============================================

CREATE TABLE al_tournament (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    game_id INT NOT NULL,
    platform_id INT NULL,
    genre_id INT NULL,
    name TEXT NOT NULL,
    description TEXT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'DRAFT',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    color_hex VARCHAR(7) NULL,
    banner_path TEXT NULL,
    logo_path TEXT NULL,

    FOREIGN KEY (organization_id) REFERENCES al_organization(id),
    FOREIGN KEY (game_id) REFERENCES al_game(id),
    FOREIGN KEY (platform_id) REFERENCES al_platform(id),
    FOREIGN KEY (genre_id) REFERENCES al_genre_e_sports(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id)
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
    user_id BIGINT NOT NULL,
    tournament_id BIGINT NOT NULL,
    role_id INT NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    active BOOLEAN NULL DEFAULT true,
    assigned_by INT NOT NULL,

    PRIMARY KEY (user_id, tournament_id, role_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
    FOREIGN KEY (role_id) REFERENCES al_tournament_role(id),
    FOREIGN KEY (assigned_by) REFERENCES al_user(id)
);
