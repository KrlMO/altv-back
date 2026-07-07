ALTER TABLE al_team_game
    ADD COLUMN region_id INT,
    ADD COLUMN country_id INT;

ALTER TABLE al_team_game
    ADD CONSTRAINT fk_team_game_region FOREIGN KEY (region_id) REFERENCES al_region(id) ON DELETE SET NULL;

ALTER TABLE al_team_game
    ADD CONSTRAINT fk_team_game_country FOREIGN KEY (country_id) REFERENCES al_country(id) ON DELETE SET NULL;
