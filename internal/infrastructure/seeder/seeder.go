package seeder

import (
	"arena/internal/domain/users"
	"arena/pkg/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminConfig struct {
	Email      string
	Password   string
	Nickname   string
	FirstName  string
	LastName   string
	CountryISO string
}

func LoadAdminConfig() *AdminConfig {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	if email == "" || password == "" {
		return nil
	}

	nickname := os.Getenv("ADMIN_NICKNAME")
	if nickname == "" {
		nickname = "admin"
	}

	firstName := os.Getenv("ADMIN_FIRST_NAME")
	if firstName == "" {
		firstName = "Admin"
	}

	lastName := os.Getenv("ADMIN_LAST_NAME")
	if lastName == "" {
		lastName = "System"
	}

	countryISO := os.Getenv("ADMIN_COUNTRY_ISO")
	if countryISO == "" {
		countryISO = "PE"
	}

	return &AdminConfig{
		Email:      email,
		Password:   password,
		Nickname:   nickname,
		FirstName:  firstName,
		LastName:   lastName,
		CountryISO: countryISO,
	}
}

func resolveCountryID(ctx context.Context, q interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, isoCode string) (int, error) {
	var id int

	err := q.QueryRow(ctx, `
		SELECT id FROM al_country WHERE iso_code = $1
	`, isoCode).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("ADMIN_COUNTRY_ISO %q does not match any al_country.iso_code", isoCode)
		}
		return 0, err
	}

	return id, nil
}

func SeedAdmin(db *pgxpool.Pool, cfg *AdminConfig) (int64, error) {
	if cfg == nil {
		log.Println("[seeder] ADMIN_EMAIL/ADMIN_PASSWORD not set, skipping admin seed")
		return 0, nil
	}

	ctx := context.Background()
	repo := users.NewRepository(db)

	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM al_user_auth
			WHERE provider = 'local' AND provider_uid = $1
		)
	`, cfg.Email).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if exists {
		log.Printf("[seeder] admin user %s already exists, skipping\n", cfg.Email)

		superAdminID, err := repo.FindSuperAdminId(ctx, db)
		if err != nil {
			return 0, err
		}

		return superAdminID, nil
	}

	countryID, err := resolveCountryID(ctx, db, cfg.CountryISO)
	if err != nil {
		return 0, err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO al_user (first_name, last_name, nickname, birth_date, country_id, active)
		VALUES ($1, $2, $3, '1990-01-01', $4, true)
		RETURNING id
	`, cfg.FirstName, cfg.LastName, cfg.Nickname, countryID).Scan(&userID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_global_role (user_id, role_id, assigned_at, assigned_by)
		SELECT $1, r.id, NOW(), $1
		FROM al_global_role r
		WHERE r.name = 'SUPER_ADMIN'
	`, userID)
	if err != nil {
		return 0, err
	}

	passwordHash, err := utils.HashPassword(cfg.Password)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_auth (user_id, provider, provider_uid, password_hash, created_by, updated_by, is_verified)
		VALUES ($1, 'local', $2, $3, $1, $1, true)
	`, userID, cfg.Email, passwordHash)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	log.Printf("[seeder] admin user created: %s (id: %d, role: SUPER_ADMIN)\n", cfg.Email, userID)
	return userID, nil
}
