package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"context"
	"fmt"
	"strings"
)

func (r *Repository) GetTeamGameRoleIDByName(ctx context.Context, roleName string) (int32, error) {
	var roleID int32
	err := r.db.QueryRow(ctx, `
		SELECT id
		FROM al_team_game_role
		WHERE name = $1
	`, roleName).Scan(&roleID)
	if err != nil {
		return 0, err
	}
	return roleID, nil
}

func (r *Repository) GetUserTeamAndTeamGameRoles(
	ctx context.Context,
	userId, teamGameId int64,
) ([]string, []string, error) {
	teamId, err := r.GetTeamIdFromTeamGameId(ctx, teamGameId)
	if err != nil {
		return nil, nil, err
	}
	teamRoles, err := r.sharedRepo.GetUserTeamRoles(ctx, userId, teamId)
	if err != nil {
		return nil, nil, err
	}
	teamGameRoles, err := r.GetUserTeamGameRoles(ctx, userId, teamGameId)
	if err != nil {
		return nil, nil, err
	}
	return teamRoles, teamGameRoles, nil
}

func (r *Repository) GetUserTeamGameRoles(ctx context.Context, userID, teamGameID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT tr.name
		FROM al_user_team_game_role utr
		JOIN al_team_game_role tr ON utr.role_id = tr.id
		WHERE utr.user_id = $1 AND utr.team_game_id = $2
	`, userID, teamGameID)
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
	return roles, rows.Err()
}

func (r *Repository) GetTeamIdFromTeamGameId(ctx context.Context, teamGameId int64) (int64, error) {
	var teamId int64
	err := r.db.QueryRow(ctx, `
		SELECT team_id
		FROM al_team_game
		WHERE id = $1
	`, teamGameId).Scan(&teamId)
	if err != nil {
		return 0, err
	}
	return teamId, nil
}

func (r *Repository) GetUserTeamRoles(ctx context.Context, userId, teamId int64) ([]string, error) {
	return r.sharedRepo.GetUserTeamRoles(ctx, userId, teamId)
}

func (r *Repository) GetGameIdFromTeamGameId(ctx context.Context, teamGameID int64) int32 {
	var gameID int32
	err := r.db.QueryRow(ctx, `
		SELECT game_id
		FROM al_team_game
		WHERE id = $1
	`, teamGameID).Scan(&gameID)
	if err != nil {
		return -1
	}
	return gameID
}

func (r *Repository) RemoveTeamGameRoleFromMember(
	ctx context.Context,
	userId, teamId int64,
	members []TeamGameMemberDto,
) error {
	if len(members) == 0 {
		return rules.ErrEmptyMembers
	}

	var (
		roleValues []string
		roleArgs   []any
		argIndex   = 1
	)

	for _, m := range members {
		for _, role := range m.Roles {
			roleValues = append(roleValues,
				fmt.Sprintf("($%d,$%d,$%d)", argIndex, argIndex+1, argIndex+2),
			)

			roleArgs = append(roleArgs,
				m.UserID,
				teamId,
				role.ID,
			)

			argIndex += 3
		}
	}

	query := `
		DELETE FROM al_user_team_game_role
		WHERE (user_id, team_game_id, role_id) IN (` + strings.Join(roleValues, ",") + `)
	`
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, roleArgs...)

	if err != nil {
		return err
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) AddTeamGameRoleToMember(
	ctx context.Context,
	userId, teamId int64,
	members []TeamGameMemberDto,
) error {
	if len(members) == 0 {
		return rules.ErrEmptyMembers
	}

	var (
		roleValues []string
		roleArgs   []any
		argIndex   = 1
	)

	for _, m := range members {
		for _, role := range m.Roles {
			roleValues = append(roleValues,
				fmt.Sprintf("($%d,$%d,$%d,NOW(),$%d)", argIndex, argIndex+1, argIndex+2, argIndex+3),
			)

			roleArgs = append(roleArgs,
				m.UserID,
				teamId,
				role.ID,
				userId,
			)

			argIndex += 4
		}
	}

	query := `
		INSERT INTO al_user_team_game_role
			(user_id, team_game_id, role_id, assigned_at, assigned_by)
		VALUES ` + strings.Join(roleValues, ",")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, roleArgs...)

	if err != nil {
		return err
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
