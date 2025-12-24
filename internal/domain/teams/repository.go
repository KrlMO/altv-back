package teams

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
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

func (r *Repository) CreateTeam(ctx context.Context, t Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO al_team
		(name, tag, description, logo_path, banner_path, country_id, is_virtual, active, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) returning id
	`, t.Name, t.Tag, t.Description, t.LogoPath, t.BannerPath, t.CountryId, t.IsVirtual, t.Active, t.CreatedBy, t.UpdatedBy).Scan(&t.ID)
	if err != nil {
		return err
	}
	const teamRole = "TEAM_OWNER"
	err = r.InsertUserTeamRole(ctx, tx, t.CreatedBy, t.ID, t.CreatedBy, teamRole)

	return tx.Commit(ctx)
}

func (r *Repository) InsertUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userID, teamID, assignedBy int64,
	roleCode string,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO al_user_team_role
		(user_id, team_id, role_id, assigned_at, assigned_by)
		SELECT
			$1,
			$2,
			r.id,
			NOW(),
			$3
		FROM al_team_role r
		WHERE r.name = $4
	`, userID, teamID, assignedBy, roleCode)

	return err
}

func (r *Repository) GetTeamByID(ctx context.Context, teamID string) (*TeamReadModel, error) {
	var t TeamReadModel

	err := r.db.QueryRow(ctx, `
		SELECT
			at.id, at.name, at.tag, at.description, at.logo_path, at.banner_path,
			at.country_id, at.is_virtual, at.active,
			ac.id, ac.name, ac.iso_code, ac.flag_path
		FROM al_team at
		JOIN al_country ac ON at.country_id = ac.id
		WHERE at.id = $1
	`, teamID).Scan(
		&t.Team.ID,
		&t.Team.Name,
		&t.Team.Tag,
		&t.Team.Description,
		&t.Team.LogoPath,
		&t.Team.BannerPath,
		&t.Team.CountryId,
		&t.Team.IsVirtual,
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
			t.logo_path,
			t.country_id,
			t.is_virtual,
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
			&t.Team.LogoPath,
			&t.Team.CountryId,
			&t.Team.IsVirtual,
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
