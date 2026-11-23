package teamgames

import (
	"arena/internal/shared"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db         *pgxpool.Pool
	sharedRepo *shared.SharedRepository
}

func NewRepository(db *pgxpool.Pool, sharedRepo *shared.SharedRepository) *Repository {
	return &Repository{db: db, sharedRepo: sharedRepo}
}
