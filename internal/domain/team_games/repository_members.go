package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/utils"
	"context"
	"fmt"
	"strings"
	"time"
)

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

var manageRolePriority = map[string]int{
	"TEAM_CAPTAIN": 1,
	"COACH":        2,
	"ANALYST":      3,
	"STAFF":        4,
	"PLAYER":       5,
}

func primaryManageRole(roles []string) string {
	best := ""
	bestPriority := 1 << 30
	for _, r := range roles {
		if p, ok := manageRolePriority[r]; ok && p < bestPriority {
			best = r
			bestPriority = p
		}
	}
	if best == "" && len(roles) > 0 {
		return roles[0]
	}
	return best
}

type manageMemberAcc struct {
	member TeamGameManageMember
	roles  []string
}

func (r *Repository) GetTeamGameManageMembers(
	ctx context.Context,
	teamGameId int64,
) ([]TeamGameManageMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			au.id AS user_id,
			au.nickname,
			au.first_name,
			au.last_name,
			au.profile_photo,
			au.birth_date,
			ac.id AS country_id,
			ac.name AS country_name,
			ac.flag_path,
			ac.iso_code,
			atgr.name AS team_game_role,
			atgpi.player_jersey_image_path
		FROM al_user au
		JOIN al_user_team_game_role autgr
			ON autgr.user_id = au.id
			AND autgr.team_game_id = $1
		JOIN al_team_game_role atgr
			ON autgr.role_id = atgr.id
		JOIN al_country ac
			ON au.country_id = ac.id
		LEFT JOIN al_team_game_player_image atgpi
			ON atgpi.user_id = au.id
			AND atgpi.team_game_id = $1
			AND atgpi.is_current = TRUE
		ORDER BY au.id
	`, teamGameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accs := map[int64]*manageMemberAcc{}
	order := []int64{}

	for rows.Next() {
		var m TeamGameManageMember
		var role string
		if err := rows.Scan(
			&m.UserID,
			&m.Username,
			&m.FirstName,
			&m.LastName,
			&m.ProfilePhoto,
			&m.BirthDate,
			&m.Country.ID,
			&m.Country.Name,
			&m.Country.FlagPath,
			&m.Country.IsoCode,
			&role,
			&m.JerseyImage,
		); err != nil {
			return nil, err
		}
		acc, ok := accs[m.UserID]
		if !ok {
			acc = &manageMemberAcc{member: m}
			accs[m.UserID] = acc
			order = append(order, m.UserID)
		}
		acc.roles = append(acc.roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]TeamGameManageMember, 0, len(order))
	for _, id := range order {
		acc := accs[id]
		acc.member.TeamRole = primaryManageRole(acc.roles)
		result = append(result, acc.member)
	}
	return result, nil
}

func (r *Repository) AddMembersToTeamGame(
	ctx context.Context,
	currentUserId, teamGameId int64,
	members []ReqTeamGameMember,
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

	if err = r.AddMembershipHistory(ctx, tx, teamGameId, gameId, "SIGNED", currentUserId, members); err != nil {
		return err
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamGameId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) RemoveMembersFromTeamGame(
	ctx context.Context,
	currentUserId, teamGameId int64,
	members []int64,
) error {
	if len(members) == 0 {
		return rules.ErrEmptyMembers
	}

	gameID := r.GetGameIdFromTeamGameId(ctx, teamGameId)
	if gameID == -1 {
		return rules.ErrCannotFindGame
	}

	membershipHistories := make([]MembershipHistoryParams, 0, len(members))
	for _, memberID := range members {
		membershipHistories = append(membershipHistories, MembershipHistoryParams{
			UserID:       memberID,
			TeamGameID:   teamGameId,
			GameID:       gameID,
			TransferType: "REMOVED",
			RemovedBy:    &currentUserId,
			LeftAt:       utils.Ptr(time.Now()),
		})
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	if err = r.deleteUsersFromTeamGame(ctx, tx, members, teamGameId); err != nil {
		return err
	}

	if err := r.InsertMembershipHistory(ctx, tx, membershipHistories); err != nil {
		return err
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamGameId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
