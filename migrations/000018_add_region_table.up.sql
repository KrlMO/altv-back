CREATE TABLE al_region (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(80) NOT NULL,
    iso_code VARCHAR(2) NOT NULL,
    flag_path TEXT NULL,

    UNIQUE(name),
    UNIQUE(iso_code)
);


INSERT INTO al_region (name, iso_code) VALUES
('Europe', 'EU'),
('North America', 'NA'),
('South America', 'SA'),
('Asia', 'AS'),
('Africa', 'AF'),
('Oceania', 'OC');