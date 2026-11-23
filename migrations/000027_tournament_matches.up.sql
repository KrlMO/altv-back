-- ============================================
-- TOURNAMENT MATCH
-- Partidos generados dentro de un stage. Los participantes son
-- inscripciones (al_tournament_registration) del torneo.
-- Igual que el roster: tournament_id denormalizado con FK compuesta para
-- que un match jamás apunte a un torneo distinto al de su stage.
-- ============================================

ALTER TABLE al_tournament_stage
    ADD CONSTRAINT tournament_stage_id_tournament_key UNIQUE (id, tournament_id);

CREATE TABLE al_tournament_match (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stage_id BIGINT NOT NULL,
    tournament_id BIGINT NOT NULL,
    round_number INT NOT NULL,
    match_number INT NOT NULL,
    team_a_registration_id BIGINT NULL,
    team_b_registration_id BIGINT NULL,
    best_of INT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    score_a INT NULL,
    score_b INT NULL,
    winner_registration_id BIGINT NULL,
    scheduled_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    CONSTRAINT tournament_match_status_check
        CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'WALKOVER')),
    CONSTRAINT tournament_match_unique UNIQUE (stage_id, round_number, match_number),

    FOREIGN KEY (stage_id) REFERENCES al_tournament_stage (id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

-- The match's tournament must match its stage's tournament.
ALTER TABLE al_tournament_match
    ADD CONSTRAINT tournament_match_stage_tournament_fk
    FOREIGN KEY (stage_id, tournament_id)
    REFERENCES al_tournament_stage (id, tournament_id);

-- Teams of a match must belong to the same tournament.
ALTER TABLE al_tournament_match
    ADD CONSTRAINT tournament_match_team_a_fk
    FOREIGN KEY (team_a_registration_id, tournament_id)
    REFERENCES al_tournament_registration (id, tournament_id);
ALTER TABLE al_tournament_match
    ADD CONSTRAINT tournament_match_team_b_fk
    FOREIGN KEY (team_b_registration_id, tournament_id)
    REFERENCES al_tournament_registration (id, tournament_id);

ALTER TABLE al_tournament_match
    ADD CONSTRAINT tournament_match_winner_fk
    FOREIGN KEY (winner_registration_id, tournament_id)
    REFERENCES al_tournament_registration (id, tournament_id);

CREATE INDEX tournament_match_stage_idx ON al_tournament_match (stage_id);
CREATE INDEX tournament_match_tournament_idx ON al_tournament_match (tournament_id);
