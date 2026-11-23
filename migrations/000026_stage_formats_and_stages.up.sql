-- ============================================
-- STAGE FORMAT CATALOG
-- Formatos de stage extensibles: agregar un formato nuevo en el futuro
-- es un INSERT aquí (más su generador de matches), no una migración destructiva.
-- ============================================

CREATE TABLE al_stage_format (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

-- Seeded with the first existing user as creator (admin is seeded at startup).
INSERT INTO al_stage_format (code, name, description, created_by)
SELECT v.code, v.name, v.description, (SELECT MIN(id) FROM al_user)
FROM (VALUES
    ('SINGLE_ELIMINATION', 'Eliminación simple', 'Los equipos se enfrentan a una sola ronda; el perdedor queda eliminado.'),
    ('DOUBLE_ELIMINATION', 'Doble eliminación', 'Bracket superior e inferior; un equipo queda eliminado tras dos derrotas.'),
    ('SWISS', 'Sistema suizo', 'Rondas emparejadas por récord; avanzan los que alcancen las victorias requeridas (estilo Major de CS2).'),
    ('GROUP_STAGE', 'Fase de grupos', 'Los equipos se dividen en grupos; avanzan los mejores de cada grupo.')
) AS v(code, name, description)
WHERE EXISTS (SELECT 1 FROM al_user);

-- ============================================
-- TOURNAMENT STAGE
-- Etapas de un torneo (estilo HLTV: Play-In → Group Stage → Playoffs).
-- El formato vive en el catálogo; la configuración específica del formato
-- vive en config JSONB y se valida en la aplicación según el código.
-- ============================================

CREATE TABLE al_tournament_stage (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tournament_id BIGINT NOT NULL,
    format_id INT NOT NULL,
    name VARCHAR(150) NOT NULL,
    stage_order INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    config JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,

    CONSTRAINT tournament_stage_status_check
        CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED')),
    CONSTRAINT tournament_stage_order_unique UNIQUE (tournament_id, stage_order),

    FOREIGN KEY (tournament_id) REFERENCES al_tournament(id),
    FOREIGN KEY (format_id) REFERENCES al_stage_format(id),
    FOREIGN KEY (created_by) REFERENCES al_user(id),
    FOREIGN KEY (updated_by) REFERENCES al_user(id),
    FOREIGN KEY (deleted_by) REFERENCES al_user(id)
);

CREATE INDEX tournament_stage_tournament_idx ON al_tournament_stage (tournament_id);
