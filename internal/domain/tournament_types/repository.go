package tournamenttypes

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	ListTournamentTypes(ctx context.Context) ([]TournamentType, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListTournamentTypes(ctx context.Context) ([]TournamentType, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name
		FROM al_tournament_type
		WHERE active = true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]TournamentType, 0, 8)
	for rows.Next() {
		var t TournamentType
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		types = append(types, t)
	}

	return types, rows.Err()
}
