package shared

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SharedRepository struct {
	db *pgxpool.Pool
}

func NewSharedRepository(db *pgxpool.Pool) *SharedRepository {
	return &SharedRepository{db: db}
}

func (r *SharedRepository) CountryExists(ctx context.Context, countryID int32) (bool, error) {
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

func (r *SharedRepository) GetUserTeamRoles(ctx context.Context, userID, teamID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT tr.name
		FROM al_user_team_role utr
		JOIN al_team_role tr ON utr.role_id = tr.id
		WHERE utr.user_id = $1 AND utr.team_id = $2
	`, userID, teamID)
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
