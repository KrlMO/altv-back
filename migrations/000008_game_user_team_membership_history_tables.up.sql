CREATE EXTENSION IF NOT EXISTS btree_gist;
ALTER TABLE al_team_game 
ADD CONSTRAINT uq_team_game_composite UNIQUE (id, game_id);
-- ============================================
-- USER - TEAM_ROLE BY TEAM
-- ============================================

CREATE TABLE al_user_team_role (
    user_id BIGINT NOT NULL,
    team_id INT NOT NULL,
    role_id INT NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by BIGINT NOT NULL,

    PRIMARY KEY (user_id, team_id, role_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (team_id) REFERENCES al_team(id),
    FOREIGN KEY (role_id) REFERENCES al_team_role(id),
    FOREIGN KEY (assigned_by) REFERENCES al_user(id)
);

-- ============================================
-- ROLE BY GAME
-- ============================================

CREATE TABLE al_game_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    game_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    ico_path TEXT,

    FOREIGN KEY (game_id) REFERENCES al_game(id)
);

-- ============================================
-- MEMBERSHIP HISTORY BY TEAM
-- ============================================

CREATE TABLE al_team_membership_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    team_game_id BIGINT NOT NULL,
    game_id INT NOT NULL,
    joined_at DATE NOT NULL DEFAULT CURRENT_DATE,
    left_at DATE NULL,
    added_by BIGINT NOT NULL,
    removed_by BIGINT,
    transfer_type VARCHAR(20) DEFAULT 'SIGNED',
    active BOOLEAN GENERATED ALWAYS AS (left_at IS NULL) STORED,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
    FOREIGN KEY (game_id) REFERENCES al_game(id),
    FOREIGN KEY (added_by) REFERENCES al_user(id),
    FOREIGN KEY (removed_by) REFERENCES al_user(id),
    CONSTRAINT prevent_overlapping_teams_per_game 
    EXCLUDE USING GIST (
        user_id WITH =,       -- Si es el mismo usuario...
        game_id WITH =,       -- Y es el mismo juego...
        daterange(joined_at, left_at, '[]') WITH && -- ¡Las fechas NO se pueden tocar!
    )
);

CREATE INDEX idx_history_user_game ON al_team_membership_history(user_id, game_id);
CREATE INDEX idx_history_active ON al_team_membership_history(team_game_id) WHERE left_at IS NULL;