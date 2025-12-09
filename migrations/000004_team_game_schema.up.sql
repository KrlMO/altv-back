-- ============================================
-- TEAM 
-- ============================================

CREATE TABLE al_team (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
    created_by INT NOT NULL,
    update_by INT NOT NULL,

    FOREIGN KEY (country_id) REFERENCES al_country(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (update_by) REFERENCES al_user(id)
);

-- ============================================
-- TEAM GAME: equipo por juego (ej: Team Spirit CS2)
-- ============================================
CREATE TABLE al_team_game (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_id INT NOT NULL,
    game_id INT NOT NULL,

    name VARCHAR(150) NOT NULL,
    logo_path TEXT,
    banner_path TEXT,
    active BOOLEAN DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by INT NOT NULL,
    update_by INT NOT NULL,

    UNIQUE(team_id, game_id),

    FOREIGN KEY (team_id) REFERENCES al_team(id),
    FOREIGN KEY (game_id) REFERENCES al_game(id)
);
