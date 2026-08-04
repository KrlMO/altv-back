-- ============================================
-- TOURNAMENT TYPE
-- ============================================

CREATE TABLE al_tournament_type (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(80) NOT NULL UNIQUE,
    description TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO al_tournament_type (name) VALUES
('Abierto'),
('Invitación'),
('Clasificatorio'),
('Liga'),
('Copa'),
('Benéfico'),
('Amateur'),
('Profesional');

-- ============================================
-- TOURNAMENT FILTER FIELDS
-- ============================================

ALTER TABLE al_tournament
    ADD COLUMN type_id INT,
    ADD COLUMN region_id INT,
    ADD COLUMN country_id INT,
    ADD COLUMN organizer_type VARCHAR(20);

ALTER TABLE al_tournament
    ADD CONSTRAINT fk_tournament_type FOREIGN KEY (type_id) REFERENCES al_tournament_type(id) ON DELETE SET NULL;

ALTER TABLE al_tournament
    ADD CONSTRAINT fk_tournament_region FOREIGN KEY (region_id) REFERENCES al_region(id) ON DELETE SET NULL;

ALTER TABLE al_tournament
    ADD CONSTRAINT fk_tournament_country FOREIGN KEY (country_id) REFERENCES al_country(id) ON DELETE SET NULL;
