package regions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	GetAllRegions(ctx context.Context) ([]Region, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAllRegions(ctx context.Context) ([]Region, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, iso_code, flag_path
		FROM al_region
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []Region
	for rows.Next() {
		var rg Region
		if err := rows.Scan(&rg.ID, &rg.Name, &rg.IsoCode, &rg.FlagPath); err != nil {
			return nil, err
		}
		regions = append(regions, rg)
	}

	return regions, rows.Err()
}
