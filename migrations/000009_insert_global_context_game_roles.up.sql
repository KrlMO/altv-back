-- ============================================
-- GLOBAL ROLES
-- ============================================

INSERT INTO al_global_role (name) VALUES
('SUPER_ADMIN'),
('STAFF'),
('USER'),
('BANNED');

-- ============================================
-- TEAM ROLES
-- ============================================

INSERT INTO al_team_role (name) VALUES
('TEAM_OWNER'),
('TEAM_MANAGER'),
('TEAM_STAFF');