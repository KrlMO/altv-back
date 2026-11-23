-- Restore the pre-000027 state.

DROP TABLE IF EXISTS al_tournament_match;

ALTER TABLE al_tournament_stage
    DROP CONSTRAINT IF EXISTS tournament_stage_id_tournament_key;
