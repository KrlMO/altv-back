package users

import (
	"context"
	"database/sql"
)

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
