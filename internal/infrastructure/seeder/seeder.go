package seeder

import (
	"arena/pkg/utils"
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminConfig struct {
	Email     string
	Password  string
	Nickname  string
	FirstName string
	LastName  string
	CountryID int32
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

	return &AdminConfig{
		Email:     email,
		Password:  password,
		Nickname:  nickname,
		FirstName: firstName,
		LastName:  lastName,
		CountryID: 174,
	}
}

func SeedAdmin(db *pgxpool.Pool, cfg *AdminConfig) error {
	if cfg == nil {
		log.Println("[seeder] ADMIN_EMAIL/ADMIN_PASSWORD not set, skipping admin seed")
		return nil
	}

	ctx := context.Background()

	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM al_user_auth
			WHERE provider = 'local' AND provider_uid = $1
		)
	`, cfg.Email).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("[seeder] admin user %s already exists, skipping\n", cfg.Email)
		return nil
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO al_user (first_name, last_name, nickname, birth_date, country_id, active)
		VALUES ($1, $2, $3, '1990-01-01', $4, true)
		RETURNING id
	`, cfg.FirstName, cfg.LastName, cfg.Nickname, cfg.CountryID).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_global_role (user_id, role_id, assigned_at, assigned_by)
		SELECT $1, r.id, NOW(), $1
		FROM al_global_role r
		WHERE r.name = 'SUPER_ADMIN'
	`, userID)
	if err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(cfg.Password)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_auth (user_id, provider, provider_uid, password_hash, created_by, updated_by, is_verified)
		VALUES ($1, 'local', $2, $3, $1, $1, true)
	`, userID, cfg.Email, passwordHash)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("[seeder] admin user created: %s (id: %d, role: SUPER_ADMIN)\n", cfg.Email, userID)
	return nil
}
