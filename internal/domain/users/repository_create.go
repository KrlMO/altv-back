package users

import (
	"arena/internal/domain/identity"
	"context"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateUser(ctx context.Context, u User, provider string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID int64

	err = tx.QueryRow(ctx, `
		insert into al_user (
			first_name,
			last_name,
			nickname,
			birth_date,
			country_id,
			profile_photo,
			banner_photo,
			active
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8
		)
		returning id
	`, u.FirstName, u.LastName, u.NickName, u.BirthDate,
		u.CountryId, u.ProfilePhoto, u.BannerPhoto, u.Active).
		Scan(&userID)

	if err != nil {
		return err
	}

	const userRole = identity.RoleUser

	superAdminId, err := r.FindSuperAdminId(ctx, tx)

	if err != nil {
		return err
	}

	err = r.InsertUserGlobalRole(ctx, tx, userID, superAdminId, userRole)

	if err != nil {
		return err
	}

	err = r.InsertUserAuth(ctx, tx, userID, userID, userID, provider, *u.Auth[0].ProviderUID, u.Auth[0].PasswordHash)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) InsertUserAuth(
	ctx context.Context,
	tx pgx.Tx,
	userID, createdBy, updatedBy int64,
	providerName, providerUID, passwordHash string) error {
	_, err := tx.Exec(ctx, `
		insert into al_user_auth (
			user_id,
			provider,
			provider_uid,
			password_hash,
			created_by,
			updated_by
		) values ($1,$2,$3,$4,$5,$6)
	`, userID, providerName, providerUID, passwordHash, createdBy, updatedBy)
	return err
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *Repository) FindSuperAdminId(ctx context.Context, q rowQuerier) (int64, error) {
	var superAdminId int64

	err := q.QueryRow(ctx, `
		SELECT u.id
		FROM al_user u
		JOIN al_user_global_role ugr ON ugr.user_id = u.id
		JOIN al_global_role gr ON gr.id = ugr.role_id
		WHERE gr.name = $1
		ORDER BY ugr.assigned_at
		LIMIT 1
	`, identity.RoleSuperAdmin).Scan(&superAdminId)

	if err != nil {
		return 0, err
	}

	return superAdminId, nil
}

func (r *Repository) InsertUserGlobalRole(
	ctx context.Context,
	tx pgx.Tx,
	userID, assignedBy int64,
	globalRole string,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO al_user_global_role
		(user_id, role_id, assigned_at, assigned_by)
		SELECT
			$1,
			r.id,
			NOW(),
			$2
		FROM al_global_role r
		WHERE r.name = $3
	`, userID, assignedBy, globalRole)
	return err
}
