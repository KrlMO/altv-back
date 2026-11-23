package teams

import (
	"arena/internal/domain/teams/rules"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) InsertUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userID, teamID, assignedBy int64,
	roleCode string,
) error {
	idRole, err := r.GetIdTeamRole(ctx, tx, roleCode)

	if err != nil || idRole == -1 {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_team_role
		(user_id, team_id, role_id, assigned_at, assigned_by)
		values
		($1, $2, $3, NOW(), $4)
	`, userID, teamID, idRole, assignedBy)

	if err != nil {
		return err
	}

	err = r.InsertUserTeamRoleHistory(ctx, tx, userID, teamID, assignedBy, idRole, "ASSIGNED")

	return err
}

func (r *Repository) GetIdTeamRole(ctx context.Context, tx pgx.Tx, roleName string) (int32, error) {
	var idRole int32
	err := tx.QueryRow(ctx, `
		SELECT id FROM al_team_role WHERE name = $1
	`, roleName).Scan(&idRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, errors.New("role does not exist")
		}
		return -1, err
	}
	return idRole, nil
}

func (r *Repository) AddMemberToTeam(
	ctx context.Context,
	userID, teamID, assignedBy int64,
	roleCode string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = r.InsertUserTeamRole(ctx, tx, userID, teamID, assignedBy, roleCode)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) TeamRoleExists(ctx context.Context, tx pgx.Tx, roleName string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_team_role WHERE name = $1
		)
	`, roleName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) UpdateUserTeamRole(
	ctx context.Context,
	userId, teamId, assigned_by int64,
	roleName string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	idTeamRole, err := r.GetIdTeamRole(ctx, tx, roleName)

	if err != nil {
		return err
	}

	if err := r.ClosePreviousPeriodUserTeamRole(ctx, tx, userId, teamId); err != nil {
		return err
	}

	query := `
		update al_user_team_role
		set 
			role_id = $1,
			assigned_at = NOW(),
			assigned_by = $2
		where user_id = $3 and team_id = $4
	`
	_, err = tx.Exec(ctx, query, idTeamRole, assigned_by, userId, teamId)

	if err != nil {
		return err
	}

	if err := r.InsertUserTeamRoleHistory(ctx, tx, userId, teamId, assigned_by, idTeamRole, "UPDATED"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) ClosePreviousPeriodUserTeamRole(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE al_user_team_role_history SET assigned_to = NOW()
		WHERE user_id = $1 AND team_id = $2 AND assigned_to is NULL
	`, userId, teamId)
	return err
}

func (r *Repository) InsertUserTeamRoleHistory(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId, assignedBy int64,
	idRole int32,
	action string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO al_user_team_role_history
		(user_id, team_id, role_id, action, assigned_from, assigned_to, assigned_by)
		values
		($1, $2, $3, $4, NOW(), NULL, $5)
	`, userId, teamId, idRole, action, assignedBy)
	return err
}

func (r *Repository) DeleteUserFromTeam(
	ctx context.Context,
	userId, teamId, assigned_by int64,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := r.CloseUserTeamRoleAsRemoved(ctx, tx, userId, teamId, assigned_by); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM al_user_team_role
		WHERE user_id = $1 AND team_id = $2
	`, userId, teamId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetUserTeamRole(
	ctx context.Context,
	userId, teamId int64,
) (string, error) {
	var actualRole string
	err := r.db.QueryRow(ctx, `
		SELECT atr.name FROM al_user_team_role autr
		JOIN al_team_role atr ON autr.role_id = atr.id
		WHERE autr.user_id = $1 AND autr.team_id = $2
	`, userId, teamId).Scan(&actualRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", rules.ErrUserNotMember
		}
		return "", err
	}

	return actualRole, err
}

func (r *Repository) CloseUserTeamRoleAsRemoved(
	ctx context.Context,
	tx pgx.Tx,
	userId, teamId, assignedBy int64,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE al_user_team_role_history
		SET
			assigned_to = NOW(),
			action = 'REMOVED',
			assigned_by = $3
		WHERE user_id = $1
		  AND team_id = $2
		  AND assigned_to IS NULL
	`, userId, teamId, assignedBy)
	return err
}

func (r *Repository) GetUserTeamCount(ctx context.Context, userId int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM al_team
		WHERE created_by = $1
	`

	var count int
	if err := r.db.QueryRow(ctx, query, userId).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
