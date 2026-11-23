package teams

import (
	"arena/internal/domain/identity"
	"arena/pkg/querybuilder"
	"context"
	"errors"
)

func (r *Repository) CreateTeam(ctx context.Context, t Team) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO al_team
		(name, tag, description, logo_key, banner_key, country_id, active, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) returning id
	`, t.Name, t.Tag, t.Description, t.LogoKey, t.BannerKey, t.CountryId, t.Active, t.CreatedBy, t.UpdatedBy).Scan(&t.ID)
	if err != nil {
		return 0, err
	}
	const teamRole = identity.RoleTeamOwner
	if err := r.InsertUserTeamRole(ctx, tx, t.CreatedBy, t.ID, t.CreatedBy, teamRole); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (r *Repository) GetTeamByID(ctx context.Context, teamID string) (*TeamReadModel, error) {
	var t TeamReadModel

	err := r.db.QueryRow(ctx, `
		SELECT
			at.id, at.name, at.tag, at.description, at.logo_key, at.banner_key,
			at.country_id, at.active,
			ac.id, ac.name, ac.iso_code, ac.flag_path
		FROM al_team at
		JOIN al_country ac ON at.country_id = ac.id
		WHERE at.id = $1
	`, teamID).Scan(
		&t.Team.ID,
		&t.Team.Name,
		&t.Team.Tag,
		&t.Team.Description,
		&t.Team.LogoKey,
		&t.Team.BannerKey,
		&t.Team.CountryId,
		&t.Team.Active,
		&t.Country.ID,
		&t.Country.Name,
		&t.Country.IsoCode,
		&t.Country.FlagPath,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetTeams(ctx context.Context, filter TeamFilter) ([]TeamReadModel, int64, error) {
	query := `
		SELECT
			t.id,
			t.name,
			t.tag,
			t.description,
			t.logo_key,
			t.banner_key,
			t.country_id,
			t.active,

			ac.id        AS country_id,
			ac.name      AS country_name,
			ac.iso_code,
			ac.flag_path

		FROM al_team t
		JOIN al_country ac 
			ON ac.id = t.country_id

		JOIN al_user_team_role utr
			ON utr.team_id = t.id

		JOIN al_team_role tr
			ON tr.id = utr.role_id
			AND tr.name = 'TEAM_OWNER'

		WHERE
			($1::INT IS NULL OR t.country_id = $1::INT)
		AND ($2::BIGINT IS NULL OR utr.user_id = $2::BIGINT)
		AND ($3::BOOLEAN IS NULL OR t.active = $3::BOOLEAN)
		AND ($4::TEXT IS NULL OR t.name ILIKE '%' || $4 || '%')
		AND ($5::TEXT IS NULL OR t.tag  ILIKE '%' || $5 || '%')

		ORDER BY t.id
		LIMIT $6 OFFSET $7;
	`

	rows, err := r.db.Query(ctx, query,
		filter.CountryID,
		filter.OwnerID,
		filter.Active,
		filter.NameLike,
		filter.TagLike,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var teams []TeamReadModel
	for rows.Next() {
		var t TeamReadModel
		err := rows.Scan(
			&t.Team.ID,
			&t.Team.Name,
			&t.Team.Tag,
			&t.Team.Description,
			&t.Team.LogoKey,
			&t.Team.BannerKey,
			&t.Team.CountryId,
			&t.Team.Active,

			&t.Country.ID,
			&t.Country.Name,
			&t.Country.IsoCode,
			&t.Country.FlagPath,
		)
		if err != nil {
			return nil, 0, err
		}
		teams = append(teams, t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	total := int64(len(teams))

	return teams, total, nil
}

func (r *Repository) UpdateTeam(ctx context.Context, t Team) error {
	b := querybuilder.NewUpdateBuilder()
	b.AddConditionalField("name", t.Name, t.Name != "")
	b.AddConditionalField("tag", t.Tag, t.Tag != "")
	b.AddConditionalField("description", t.Description, t.Description != "")
	b.AddConditionalField("logo_key", t.LogoKey, t.LogoKey != "")
	b.AddConditionalField("banner_key", t.BannerKey, t.BannerKey != "")
	b.AddConditionalField("country_id", t.CountryId, t.CountryId != 0)
	b.AddField("updated_by", t.UpdatedBy)

	if b.IsEmpty() {
		return errors.New("nothing to update")
	}

	query, args := b.Build("al_team", "id", t.ID)

	_, err := r.db.Exec(ctx, query, args...)

	return err
}

func (r *Repository) GetMyTeams(ctx context.Context, userId int64, globalRoles []string) ([]MyTeamListItem, error) {
	query := `
		SELECT DISTINCT
			t.id,
			t.name,
			t.tag,
			t.logo_key,
			tr.name AS member_role
		FROM al_team t
		JOIN al_user_team_role utr ON utr.team_id = t.id
		JOIN al_team_role tr ON tr.id = utr.role_id
		WHERE utr.user_id = $1
		ORDER BY t.id
	`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows == nil {
		return nil, nil
	}

	var teams []MyTeamListItem
	for rows.Next() {
		var t MyTeamListItem
		if err := rows.Scan(&t.ID, &t.Name, &t.Tag, &t.LogoKey, &t.MemberRole); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}
