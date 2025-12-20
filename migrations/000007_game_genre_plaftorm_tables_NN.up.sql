-- ============================================
-- GAME - GENRE
-- ============================================
CREATE TABLE al_game_genre (
    game_id INT NOT NULL,
    genre_id INT NOT NULL,

    PRIMARY KEY (game_id, genre_id),

    FOREIGN KEY (game_id) REFERENCES al_game(id),
    FOREIGN KEY (genre_id) REFERENCES al_genre_e_sports(id)
);

-- ============================================
-- GAME - PLATFORM
-- ============================================

CREATE TABLE al_game_platform (
    game_id INT NOT NULL,
    platform_id INT NOT NULL,

    PRIMARY KEY (game_id, platform_id),

    FOREIGN KEY (game_id) REFERENCES al_game(id),
    FOREIGN KEY (platform_id) REFERENCES al_platform(id)
);