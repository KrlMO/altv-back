-- Indices
DROP INDEX IF EXISTS idx_history_active;
DROP INDEX IF EXISTS idx_history_user_game;

-- Tables (en orden inverso de creación)
DROP TABLE IF EXISTS al_team_membership_history;
DROP TABLE IF EXISTS al_game_role;
DROP TABLE IF EXISTS al_user_team_role;

-- Constraint adicional
ALTER TABLE al_team_game 
DROP CONSTRAINT IF EXISTS uq_team_game_composite;

-- Extension (opcional, solo si esta migración la creó)
DROP EXTENSION IF EXISTS btree_gist;
