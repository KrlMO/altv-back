package users

import "context"

func (r *Repository) SearchUsers(ctx context.Context, query string) ([]UserSearchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			au.id,
			au.first_name,
			au.last_name,
			au.nickname,
			COALESCE(au.profile_photo, ''),
			TO_CHAR(au.birth_date, 'YYYY-MM-DD'),
			ac.id,
			COALESCE(ac.name, ''),
			COALESCE(ac.flag_path, '')
		FROM al_user au
		JOIN al_country ac ON au.country_id = ac.id
		JOIN al_user_global_role ugr ON ugr.user_id = au.id
		JOIN al_global_role agr ON agr.id = ugr.role_id
		WHERE (au.first_name ILIKE $1 OR au.last_name ILIKE $1 OR au.nickname ILIKE $1) AND agr.name = 'USER'
		LIMIT 5
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []UserSearchResult

	for rows.Next() {
		var usr UserSearchResult
		err := rows.Scan(
			&usr.ID,
			&usr.FirstName,
			&usr.LastName,
			&usr.NickName,
			&usr.ProfilePhoto,
			&usr.BirthDate,
			&usr.Country.ID,
			&usr.Country.Name,
			&usr.Country.FlagPath,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, usr)
	}

	return results, nil
}
