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
	err = r.InsertUserTeamRole(ctx, tx, t.CreatedBy, int32(t.ID), t.CreatedBy, teamRole)

	return tx.Commit(ctx)
}

func (r *Repository) InsertUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userID, teamID, assignedBy int32,
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
