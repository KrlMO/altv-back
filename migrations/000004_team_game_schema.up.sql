-- ============================================
-- TEAM 
-- ============================================

CREATE TABLE al_team (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    tag VARCHAR(15) NOT NULL UNIQUE,
    description TEXT,
    logo_path TEXT,
    banner_path TEXT,
    country_id INT NOT NULL,
    active BOOLEAN DEFAULT true,
    is_virtual BOOLEAN DEFAULT false,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,

    FOREIGN KEY (country_id) REFERENCES al_country(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id)
);

-- ============================================
-- TEAM GAME: equipo por juego (ej: Team Spirit CS2)
-- ============================================
CREATE TABLE al_team_game (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id BIGINT NOT NULL,
    game_id INT NOT NULL,
    
    active BOOLEAN DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,

    UNIQUE(team_id, game_id),

    FOREIGN KEY (team_id) REFERENCES al_team(id),
    FOREIGN KEY (game_id) REFERENCES al_game(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id)
);
