package teamgames

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTeamGamePublicMembers(
	ctx context.Context,
	teamGameId int64) ([]TeamGameMember, error) {
	members := []TeamGameMember{}
	rows, err := r.db.Query(ctx, `
			SELECT 
				atg.id AS team_game_id,
				au.id AS user_id,
				au.first_name,
				au.last_name,
				au.nickname,
				ac.name AS country_name,
				ac.flag_path,
				ac.iso_code,
				au.profile_photo,
				atg.team_id,
				atg.game_id,
				atgr.name AS team_game_role,
				atmh.joined_at,
				atmh.transfer_type,
				atgpi.player_jersey_image_path
			FROM al_user au
			JOIN al_user_team_game_role autgr
				ON autgr.user_id = au.id
			JOIN al_team_game_role atgr
				ON autgr.role_id = atgr.id
			AND atgr.name IN ('COACH', 'PLAYER', 'TEAM_CAPTAIN')
			JOIN al_team_game atg
				ON autgr.team_game_id = atg.id
			JOIN al_country ac
				ON au.country_id = ac.id
			JOIN al_team_game_player_image atgpi
				ON atgpi.user_id = au.id
			AND atgpi.team_game_id = atg.id
			AND atgpi.is_current = TRUE
			LEFT JOIN al_team_membership_history atmh
				ON atmh.user_id = au.id
				AND atmh.team_game_id = atg.id
				AND atmh.left_at IS NULL
			WHERE atg.id = $1;

		`, teamGameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		member := TeamGameMember{}
		if err := rows.Scan(
			&member.TeamGameID,
			&member.UserID,
			&member.FirstName,
			&member.LastName,
			&member.Nickname,
			&member.Country.Name,
			&member.Country.FlagPath,
			&member.Country.IsoCode,
			&member.ProfilePhoto,
			&member.TeamID,
			&member.GameID,
			&member.TeamGameRole,
			&member.JoinedAt,
			&member.TransferType,
			&member.PlayerImage,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *Repository) GetUserTeamAndTeamGameRoles(
	ctx context.Context,
	userId, teamGameId int64,
) ([]string, []string, error) {
	teamRoles := []string{}
	teamGameRoles := []string{}
	teamId, err := r.GetTeamIdFromTeamGameId(ctx, teamGameId)
	if err != nil {
		return nil, nil, err
	}
	teamRoles, err = r.GetUserTeamRoles(ctx, userId, teamId)
	if err != nil {
		return nil, nil, err
	}
	teamGameRoles, err = r.GetUserTeamGameRoles(ctx, userId, teamGameId)
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
	return roles, nil
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

func (r *Repository) GetUserTeamRoles(ctx context.Context, userID, teamID int64) ([]string, error) {
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
func (r *Repository) AddMembersToTeamGame(
	ctx context.Context,
	currentUserId, teamGameId int64,
	members []NewTeamGameMember,
	gameId int32,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var (
		roleValues []string
		roleArgs   []any
	)

	for i, m := range members {
		base := i * 4
		roleValues = append(roleValues,
			fmt.Sprintf("($%d,$%d,$%d,NOW(),$%d)", base+1, base+2, base+3, base+4),
		)
		roleArgs = append(roleArgs,
			m.UserID,
			teamGameId,
			m.TeamGameRoleID,
			currentUserId,
		)
	}

	queryTeamGame := `
		INSERT INTO al_user_team_game_role
			(user_id, team_game_id, role_id, assigned_at, assigned_by)
		VALUES ` + strings.Join(roleValues, ",")

	if _, err := tx.Exec(ctx, queryTeamGame, roleArgs...); err != nil {
		return err
	}

	r.AddMembershipHistory(ctx, tx, teamGameId, gameId, "ADDED", currentUserId, members)

	return tx.Commit(ctx)
}

func (r *Repository) AddMembershipHistory(
	ctx context.Context,
	tx pgx.Tx,
	teamGameId int64,
	gameId int32,
	transferType string,
	addedBy int64,
	members []NewTeamGameMember,
) error {

	var (
		historyValues []string
		historyArgs   []any
	)

	for i, m := range members {
		base := i * 5
		historyValues = append(historyValues,
			fmt.Sprintf("($%d,$%d,$%d,NOW(),$%d,$%d,NOW(),NOW())", base+1, base+2, base+3, base+4, base+5),
		)
		historyArgs = append(historyArgs,
			m.UserID,
			teamGameId,
			gameId,
			addedBy,
			transferType,
		)
	}

	queryHistory := `
		INSERT INTO al_team_membership_history
			(user_id, team_game_id, game_id, joined_at, added_by, transfer_type, created_at, updated_at)
		VALUES ` + strings.Join(historyValues, ",")

	_, err := tx.Exec(ctx, queryHistory, historyArgs...)
	return err
}
