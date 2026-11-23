-- ============================================
-- TOURNAMENT REGISTRATION
-- Inscripción de un roster (team_game) a un torneo.
-- ============================================

CREATE TABLE al_tournament_registration (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    team_game_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    registered_by BIGINT NOT NULL,
    reviewed_by BIGINT NULL,
    reviewed_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    CONSTRAINT tournament_registration_status_check
        CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'CANCELLED')),
    CONSTRAINT tournament_registration_unique UNIQUE (tournament_id, team_game_id),

    FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
    FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
    FOREIGN KEY (registered_by) REFERENCES al_user(id),
    FOREIGN KEY (reviewed_by) REFERENCES al_user(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE INDEX tournament_registration_tournament_idx ON al_tournament_registration (tournament_id);
CREATE INDEX tournament_registration_team_game_idx ON al_tournament_registration (team_game_id);

-- ============================================
-- TOURNAMENT ROSTER (recreated)
-- Jugadores que compiten por una inscripción.
--
-- Cambios respecto a 000006:
--   * Cuelga de la inscripción (registration_id) con ON DELETE CASCADE:
--     rechazar/cancelar una inscripción limpia su roster automáticamente.
--   * UNIQUE (tournament_id, user_id): un usuario solo puede estar en un
--     roster por torneo, sin importar el equipo o el juego.
--   * FK compuesta (registration_id, tournament_id): la fila de roster no
--     puede apuntar a un torneo distinto al de su inscripción.
-- ============================================

DROP TABLE IF EXISTS al_tournament_roster;

CREATE TABLE al_tournament_roster (
    registration_id BIGINT NOT NULL,
    tournament_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    team_game_id BIGINT NOT NULL,
    player_image_id BIGINT NULL,
    is_starter BOOLEAN NOT NULL DEFAULT TRUE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    added_by BIGINT NOT NULL,

    PRIMARY KEY (registration_id, user_id),
    CONSTRAINT tournament_roster_unique_per_tournament UNIQUE (tournament_id, user_id),

    FOREIGN KEY (registration_id) REFERENCES al_tournament_registration (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
    FOREIGN KEY (player_image_id) REFERENCES al_team_game_player_image(id),
    FOREIGN KEY (added_by) REFERENCES al_user(id)
);

-- Unique constraint backing the composite foreign key below.
ALTER TABLE al_tournament_registration
    ADD CONSTRAINT tournament_registration_id_tournament_key UNIQUE (id, tournament_id);

-- The roster's tournament must match its registration's tournament.
ALTER TABLE al_tournament_roster
    ADD CONSTRAINT tournament_roster_registration_tournament_fk
    FOREIGN KEY (registration_id, tournament_id)
    REFERENCES al_tournament_registration (id, tournament_id);
