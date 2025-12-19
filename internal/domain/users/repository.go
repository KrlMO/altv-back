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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID int

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

	const humanRoleID = 5
	const iluvatarRoleID = 1

	_, err = tx.Exec(ctx, `
		insert into al_user_global_role (
			user_id,
			role_id,
			assigned_at,
			assigned_by
		) values ($1, $2, now(), $3)
	`, userID, humanRoleID, iluvatarRoleID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
