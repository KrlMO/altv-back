package games

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	ListGames(ctx context.Context, name string) ([]Game, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListGames(ctx context.Context, name string) ([]Game, error) {
	query := `
		SELECT id, name, slug, color_hex
		FROM al_game
		WHERE active = true
	`

	var args []interface{}
	if name != "" {
		query += " AND name ILIKE '%' || $1 || '%'"
		args = append(args, name)
	}

	query += " ORDER BY popularity_rank ASC NULLS LAST LIMIT 5"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug, &g.ColorHex); err != nil {
			return nil, err
		}
		games = append(games, g)
	}

	return games, rows.Err()
}
