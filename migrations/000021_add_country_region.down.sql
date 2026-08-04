ALTER TABLE al_country
    DROP CONSTRAINT IF EXISTS fk_country_region;

ALTER TABLE al_country
    DROP COLUMN IF EXISTS region_id;
