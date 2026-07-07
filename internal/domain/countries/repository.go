package countries

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	GetAllCountries(ctx context.Context) ([]Country, error)
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

func (r *Repository) GetAllCountries(ctx context.Context) ([]Country, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, iso_code, flag_path FROM al_country")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var c Country
		if err := rows.Scan(&c.ID, &c.Name, &c.IsoCode, &c.FlagPath); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}

	return countries, nil
}
