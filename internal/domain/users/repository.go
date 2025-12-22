package users

import (
	"context"
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

func (r *Repository) CreateUser(ctx context.Context, u User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID int32

	err = tx.QueryRow(ctx, `
		insert into al_user (
			first_name,
			last_name,
			nickname,
			password_hash,
			birth_date,
			country_id,
			email,
			profile_photo,
			banner_photo,
			active
		) values (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		)
		returning id
	`, u.FirstName, u.LastName, u.NickName, u.PasswordHash, u.BirthDate,
		u.CountryId, u.Email, u.ProfilePhoto, u.BannerPhoto, u.Active).
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

	return tx.Commit(ctx)
}

func (r *Repository) FindIluvatarId(ctx context.Context, tx pgx.Tx) (int32, error) {
	var iluvatarId int32

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
	userID, assignedBy int32,
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

func (r *Repository) GetUserByEmailOrNickname(ctx context.Context, EmailOrNickname string) (*User, error) {
	u := User{}
	str := `select id, first_name, last_name, nickname, email, password_hash, profile_photo from al_user where email = $1 or nickname = $1`
	err := r.db.QueryRow(ctx, str, EmailOrNickname).Scan(&u.ID, &u.FirstName, &u.LastName, &u.NickName, &u.Email, &u.PasswordHash, &u.ProfilePhoto)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetGlobalRolesByUserID(ctx context.Context, userID int32) ([]string, error) {
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

func (r *Repository) GetUserProfileByID(ctx context.Context, userID int32) (*UserProfile, error) {
	up := UserProfile{}
	str := `
		select
			au.id,
			au.first_name,
			au.last_name,
			au.nickname,
			au.email,
			au.profile_photo,
			au.banner_photo,
			au.birth_date,
			ac.id,
			ac."name",
			ac.flag_path
		from al_user au
		join al_country ac on au.country_id = ac.id
		where au.id = $1
	`
	err := r.db.QueryRow(ctx, str, userID).Scan(&up.User.ID, &up.User.FirstName, &up.User.LastName, &up.User.NickName, &up.User.Email,
		&up.User.ProfilePhoto, &up.User.BannerPhoto, &up.User.BirthDate,
		&up.Country.ID, &up.Country.Name, &up.Country.FlagPath)
	if err != nil {
		return nil, err
	}
	return &up, nil
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
