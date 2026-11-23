-- Restore the original 000006 roster table.

DROP TABLE IF EXISTS al_tournament_roster;

ALTER TABLE al_tournament_registration
    DROP CONSTRAINT IF EXISTS tournament_roster_registration_tournament_fk;
ALTER TABLE al_tournament_registration
    DROP CONSTRAINT IF EXISTS tournament_registration_id_tournament_key;

DROP TABLE IF EXISTS al_tournament_registration;

CREATE TABLE al_tournament_roster (
    user_id BIGINT NOT NULL,
    tournament_id BIGINT NOT NULL,
    team_game_id BIGINT NOT NULL,
    player_image_id BIGINT NOT NULL,
    is_starter BOOLEAN NOT NULL DEFAULT TRUE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    added_by BIGINT NOT NULL,

    PRIMARY KEY (user_id, tournament_id),

    FOREIGN KEY (user_id) REFERENCES al_user(id),
    FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
    FOREIGN KEY (team_game_id) REFERENCES al_team_game(id),
    FOREIGN KEY (player_image_id) REFERENCES al_team_game_player_image(id),
    FOREIGN KEY (added_by) REFERENCES al_user(id)
);
