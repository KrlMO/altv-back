package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

	const humanRole = "HUMAN"

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
	`, "ILUVATAR").Scan(&iluvatarId)

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
	var auth UserAuth
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

	u.Auth = []UserAuth{auth}
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
			au.profile_photo,
			au.banner_photo,
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

func (r *Repository) getUserProviders(ctx context.Context, userID int64) ([]UserAuth, error) {

	rows, err := r.db.Query(ctx, `
		SELECT id, provider, provider_uid, password_hash
		FROM al_user_auth
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auths []UserAuth

	for rows.Next() {
		var ua UserAuth
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

	auths, err := r.getUserProviders(ctx, userID)
	if err != nil {
		return nil, err
	}

	up.User.Auth = auths

	return up, nil
}

func (r *Repository) CountryExists(ctx context.Context, countryID int32) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		select exists(
			select 1 from al_country where id = $1
		)
	`, countryID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
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

	set := []string{}
	args := []interface{}{}
	i := 1

	if u.FirstName != "" {
		set = append(set, fmt.Sprintf("first_name = $%d", i))
		args = append(args, u.FirstName)
		i++
	}

	if u.LastName != "" {
		set = append(set, fmt.Sprintf("last_name = $%d", i))
		args = append(args, u.LastName)
		i++
	}

	if u.NickName != "" {
		set = append(set, fmt.Sprintf("nickname = $%d", i))
		args = append(args, u.NickName)
		i++
	}

	if u.ProfilePhoto != "" {
		set = append(set, fmt.Sprintf("profile_photo = $%d", i))
		args = append(args, u.ProfilePhoto)
		i++
	}

	if u.BannerPhoto != "" {
		set = append(set, fmt.Sprintf("banner_photo = $%d", i))
		args = append(args, u.BannerPhoto)
		i++
	}

	if u.CountryId > 0 {
		set = append(set, fmt.Sprintf("country_id = $%d", i))
		args = append(args, u.CountryId)
		i++
	}

	if u.BirthDate.Valid {
		set = append(set, fmt.Sprintf("birth_date = $%d", i))
		args = append(args, u.BirthDate.Time)
		i++
	}

	if len(set) == 0 {
		return errors.New("nothing to update")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`
		UPDATE al_user
		SET %s, updated_at = NOW()
		WHERE id = $%d
	`, strings.Join(set, ", "), i)

	args = append(args, u.ID)

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

func (r *Repository) GetUserAuthByUserID(
	ctx context.Context,
	userId int64,
) ([]UserAuth, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			provider,
			provider_uid,
			password_hash
		FROM al_user_auth
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auths []UserAuth

	for rows.Next() {
		var ua UserAuth

		err := rows.Scan(
			&ua.ID,
			&ua.ProviderName,
			&ua.ProviderUID,
			&ua.PasswordHash,
		)
		if err != nil {
			return nil, err
		}

		auths = append(auths, ua)
	}

	return auths, nil
}
