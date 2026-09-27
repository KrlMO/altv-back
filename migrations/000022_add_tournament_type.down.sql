ALTER TABLE al_tournament
    DROP CONSTRAINT IF EXISTS fk_tournament_country;

ALTER TABLE al_tournament
    DROP CONSTRAINT IF EXISTS fk_tournament_region;

ALTER TABLE al_tournament
    DROP CONSTRAINT IF EXISTS fk_tournament_type;

ALTER TABLE al_tournament
    DROP COLUMN IF EXISTS organizer_type,
    DROP COLUMN IF EXISTS country_id,
    DROP COLUMN IF EXISTS type_id;

DROP TABLE IF EXISTS al_tournament_type;
