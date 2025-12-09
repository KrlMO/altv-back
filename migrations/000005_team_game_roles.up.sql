-- ============================================
-- ROLES POR EQUIPO-POR-JUEGO
-- ============================================

CREATE TABLE al_team_game_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Roles oficiales
INSERT INTO al_team_game_role (name) VALUES
('PLAYER'),
('SUBSTITUTE'),
('COACH'),
('ANALYST'),
('MANAGER'),
('STAFF');

-- ============================================
-- ASIGNACIÓN: USER → TEAM_GAME → ROLE
-- ============================================

CREATE TABLE al_user_team_game_role (
    user_id INT NOT NULL,
    team_game_id INT NOT NULL,
    role_id INT NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by INT NOT NULL,

    PRIMARY KEY (user_id, team_game_id, role_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
    FOREIGN KEY (role_id) REFERENCES al_team_game_role(id),
    FOREIGN KEY (assigned_by) REFERENCES al_user(id)
);
