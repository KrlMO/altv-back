-- =============================
-- Initial Schema Migration
-- =============================

-- 1. Country
CREATE TABLE al_country (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(80) NOT NULL,
    iso_code VARCHAR(2) NOT NULL,
    flag_path TEXT
);

-- 2. Global Role
CREATE TABLE al_global_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- 3. Team Role
CREATE TABLE al_team_role (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- 4. User
CREATE TABLE al_user (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    nickname VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    birth_date DATE NOT NULL,
    registered_at TIMESTAMP DEFAULT NOW(),
    country_id INT NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    profile_photo TEXT,
    banner_photo TEXT,
    active BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_user_country FOREIGN KEY (country_id) REFERENCES al_country(id)
);

-- 5. User Global Role (pivot)
CREATE TABLE al_user_global_role (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by INT,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_ugr_user FOREIGN KEY (user_id) REFERENCES al_user(id),
    CONSTRAINT fk_ugr_role FOREIGN KEY (role_id) REFERENCES al_global_role(id),
    CONSTRAINT fk_ugr_created_by FOREIGN KEY (created_by) REFERENCES al_user(id)
);

-- 6. 