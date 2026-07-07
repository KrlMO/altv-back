ALTER TABLE al_team_game
    DROP CONSTRAINT IF EXISTS fk_team_game_country;

ALTER TABLE al_team_game
    DROP CONSTRAINT IF EXISTS fk_team_game_region;

ALTER TABLE al_team_game
    DROP COLUMN IF EXISTS country_id;

ALTER TABLE al_team_game
    DROP COLUMN IF EXISTS region_id;
