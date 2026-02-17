package auth

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

type AuthRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeToken(ctx context.Context, id int64) error
	RevokeAllByUser(ctx context.Context, userID int64) error
}

func (r *Repository) CreateRefreshToken(
	ctx context.Context,
	token *RefreshToken,
) error {

	_, err := r.db.Exec(ctx, `
		INSERT INTO al_refresh_tokens
			(user_id, token_hash, device_info, ip_address, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6)
	`,
		token.UserID,
		token.TokenHash,
		token.DeviceInfo,
		token.IPAddress,
		token.ExpiresAt,
		token.Revoked,
	)

	return err
}

func (r *Repository) GetByTokenHash(
	ctx context.Context,
	hash string,
) (*RefreshToken, error) {

	t := RefreshToken{}

	err := r.db.QueryRow(ctx, `
		SELECT
			art.id,
			art.user_id,
			art.token_hash,
			art.device_info,
			art.ip_address,
			art.expires_at,
			art.created_at,
			art.revoked
		FROM al_refresh_tokens art
		WHERE art.token_hash = $1
	`, hash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.DeviceInfo,
		&t.IPAddress,
		&t.ExpiresAt,
		&t.CreatedAt,
		&t.Revoked,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

func (r *Repository) RevokeAllByUser(
	ctx context.Context,
	userID int64,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`, userID)

	return err
}

func (r *Repository) RevokeToken(
	ctx context.Context,
	id int64,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_refresh_tokens
		SET revoked = true
		WHERE id = $1 AND revoked = false
	`, id)

	return err
}
