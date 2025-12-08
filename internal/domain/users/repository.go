package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateUser(ctx context.Context, u User) error {
	str := `inserto into al_user (
			first_name,
			last_name,
			nickname,
			password_hash,
			birth_date,
			country_id,
			email,
			profile_photo,
			banner_photo,
			active,
		) values (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`
	_, err := r.db.Exec(ctx, str, u.FirstName, u.LastName, u.NickName, u.PasswordHash, u.BirthDate, u.CountryId, u.Email, u.ProfilePhoto, u.BannerPhoto, u.Active)
	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	str := `select id, first_name, last_name, nickname, email, password_hash from al_user where email = $1`
	err := r.db.QueryRow(ctx, str, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.NickName, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
