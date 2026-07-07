package users

import (
	"arena/internal/domain/identity"
	"arena/pkg/querybuilder"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

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

	const humanRole = identity.RoleHuman

	iluvatarId, err := r.FindIluvatarId(ctx, tx)

	if err != nil {
		return err
	}

	err = r.InsertUserGlobalRole(ctx, tx, userID, iluvatarId, humanRole)

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

func (r *Repository) FindIluvatarId(ctx context.Context, tx pgx.Tx) (int64, error) {
	var iluvatarId int64

	err := tx.QueryRow(ctx, `
		SELECT id FROM al_global_role WHERE name = $1
	`, identity.RoleIluvatar).Scan(&iluvatarId)

	if err != nil {
		return 0, err
	}

	return iluvatarId, nil
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

func (r *Repository) GetUserByProviderUID(
	ctx context.Context,
	providerUID string,
	provider string,
) (*User, error) {

	var u User
	var auth ProviderAuth
	var passwordHash sql.NullString

	err := r.db.QueryRow(ctx, `
	SELECT
		au.id,
		au.first_name,
		au.last_name,
		au.nickname,
		au.birth_date,
		au.active,
		aua.id,
		aua.provider,
		aua.provider_uid,
		aua.password_hash,
		aua.is_verified
	FROM al_user_auth aua
	JOIN al_user au ON au.id = aua.user_id
	WHERE aua.provider = $1
	  AND aua.provider_uid = $2
`, provider, providerUID).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.NickName,
		&u.BirthDate,
		&u.Active,
		&auth.ID,
		&auth.ProviderName,
		&auth.ProviderUID,
		&passwordHash,
		&auth.IsVerified,
	)

	if passwordHash.Valid {
		auth.PasswordHash = passwordHash.String
	}

	if err != nil {
		return nil, err
	}

	u.Auth = []ProviderAuth{auth}
	return &u, nil
}

func (r *Repository) GetGlobalRolesByUserID(ctx context.Context, userID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		select agr.name
		from al_global_role agr
		join al_user_global_role ugr on ugr.role_id = agr.id
		where ugr.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, nil
}

func (r *Repository) getBaseProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	up := UserProfile{}

	err := r.db.QueryRow(ctx, `
		SELECT
			au.id,
			au.first_name,
			au.last_name,
			au.nickname,
			COALESCE(au.profile_photo, ''),
			COALESCE(au.banner_photo, ''),
			au.birth_date,
			ac.id,
			ac.name,
			ac.flag_path
		FROM al_user au
		JOIN al_country ac ON au.country_id = ac.id
		WHERE au.id = $1
	`, userID).Scan(
		&up.User.ID,
		&up.User.FirstName,
		&up.User.LastName,
		&up.User.NickName,
		&up.User.ProfilePhoto,
		&up.User.BannerPhoto,
		&up.User.BirthDate,
		&up.Country.ID,
		&up.Country.Name,
		&up.Country.FlagPath,
	)

	if err != nil {
		return nil, err
	}

	return &up, nil
}

func (r *Repository) GetUserProviders(ctx context.Context, userID int64) ([]ProviderAuth, error) {

	rows, err := r.db.Query(ctx, `
		SELECT id, provider, provider_uid, password_hash
		FROM al_user_auth
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auths []ProviderAuth

	for rows.Next() {
		var ua ProviderAuth
		var passwordHash sql.NullString

		err := rows.Scan(
			&ua.ID,
			&ua.ProviderName,
			&ua.ProviderUID,
			&passwordHash,
		)
		if err != nil {
			return nil, err
		}

		if passwordHash.Valid {
			ua.PasswordHash = passwordHash.String
		}

		auths = append(auths, ua)
	}

	return auths, nil
}

func (r *Repository) GetUserProfileByID(ctx context.Context, userID int64) (*UserProfile, error) {

	up, err := r.getBaseProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	auths, err := r.GetUserProviders(ctx, userID)
	if err != nil {
		return nil, err
	}

	up.User.Auth = auths

	return up, nil
}

func (r *Repository) UpdateUser(ctx context.Context, u User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		update al_user set
			first_name = $1,
			last_name = $2,
			nickname = $3,
			profile_photo = $4,
			banner_photo = $5,
			birth_date = $6,
			country_id = $7,
			updated_at = now()
		where id = $8
	`, u.FirstName, u.LastName, u.NickName, u.ProfilePhoto, u.BannerPhoto, u.BirthDate, u.CountryId, u.ID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) PatchUser(ctx context.Context, u User) error {
	b := querybuilder.NewUpdateBuilder()
	b.AddConditionalField("first_name", u.FirstName, u.FirstName != "")
	b.AddConditionalField("last_name", u.LastName, u.LastName != "")
	b.AddConditionalField("nickname", u.NickName, u.NickName != "")
	b.AddConditionalField("profile_photo", u.ProfilePhoto, u.ProfilePhoto != "")
	b.AddConditionalField("banner_photo", u.BannerPhoto, u.BannerPhoto != "")
	b.AddConditionalField("country_id", u.CountryId, u.CountryId > 0)
	b.AddConditionalField("birth_date", u.BirthDate.Time, u.BirthDate.Valid)

	if b.IsEmpty() {
		return errors.New("nothing to update")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query, args := b.Build("al_user", "id", u.ID)

	_, err = tx.Exec(ctx, query, args...)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetUserByID(
	ctx context.Context,
	userId int64,
) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, `
		SELECT
			au.id,
			au.first_name,
			au.last_name,
			au.nickname,
			au.birth_date,
			au.active
		FROM al_user au
		WHERE au.id = $1
	`, userId).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.NickName,
		&u.BirthDate,
		&u.Active)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) DeactivateUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_user
		SET active = false, updated_at = NOW()
		WHERE id = $1
	`, userID)
	return err
}

func (r *Repository) ActivateUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_user
		SET active = true, updated_at = NOW(), deleted_at = NULL
		WHERE id = $1
	`, userID)
	return err
}

func (r *Repository) SoftDeleteUser(ctx context.Context, userID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE al_user
		SET active = false,
			deleted_at = NOW(),
			updated_at = NOW(),
			first_name = 'Deleted',
			last_name = 'User',
			nickname = 'deleted_' || $1::text,
			profile_photo = '',
			banner_photo = ''
		WHERE id = $1
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE al_user_auth
		SET password_hash = '', updated_at = NOW()
		WHERE user_id = $1 AND provider = 'local'
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM al_refresh_tokens WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM al_refresh_tokens WHERE user_id = $1
	`, userID)
	return err
}
