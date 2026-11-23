package users

import (
	"arena/pkg/querybuilder"
	"context"
	"errors"
)

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
	_, err := r.db.Exec(ctx, `
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

	return err
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

	query, args := b.Build("al_user", "id", u.ID)

	_, err := r.db.Exec(ctx, query, args...)

	return err
}
