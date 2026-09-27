package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/internal/shared"
	"arena/pkg/utils"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db         *pgxpool.Pool
	sharedRepo *shared.SharedRepository
}

type MembershipHistoryParams struct {
	UserID       int64
	TeamGameID   int64
	GameID       int32
	TransferType string

	AddedBy   *int64
	RemovedBy *int64

	JoinedAt *time.Time
	LeftAt   *time.Time
}

func NewRepository(db *pgxpool.Pool, sharedRepo *shared.SharedRepository) *Repository {
	return &Repository{db: db, sharedRepo: sharedRepo}
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

func (r *Repository) UpsertTeamGamePlayerImage(
	ctx context.Context,
	userID, teamGameID, createdBy int64,
	jerseyImagePath string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE al_team_game_player_image
		SET is_current = FALSE
		WHERE user_id = $1 AND team_game_id = $2
	`, userID, teamGameID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO al_team_game_player_image
			(user_id, team_game_id, player_jersey_image_path, is_current, created_by)
		VALUES ($1, $2, $3, TRUE, $4)
	`, userID, teamGameID, jerseyImagePath, createdBy); err != nil {
		return err
	}

	return tx.Commit(ctx)
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

func (r *Repository) GetGameIdBySlug(ctx context.Context, gameSlug string) (int32, error) {
	var gameId int32
	err := r.db.QueryRow(ctx, `
		SELECT id
		FROM al_game
		WHERE slug = $1
	`, gameSlug).Scan(&gameId)
	if err != nil {
		return 0, err
	}
	return gameId, nil
}

func (r *Repository) GetUserTeamRoles(ctx context.Context, userId, teamId int64) ([]string, error) {
	return r.sharedRepo.GetUserTeamRoles(ctx, userId, teamId)
}

func (r *Repository) CreateTeamGameWithMembers(
	ctx context.Context,
	teamId, createdBy int64,
	gameID int32,
	members []ReqTeamGameMember,
	jerseyPaths map[int64]string,
) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var teamGameId int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO al_team_game(team_id, game_id, created_by, updated_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, teamId, gameID, createdBy, createdBy).Scan(&teamGameId); err != nil {
		return 0, err
	}

	if len(members) > 0 {
		var (
			roleValues []string
			roleArgs   []any
		)
		for i, m := range members {
			base := i * 4
			roleValues = append(roleValues,
				fmt.Sprintf("($%d,$%d,$%d,NOW(),$%d)", base+1, base+2, base+3, base+4),
			)
			roleArgs = append(roleArgs, m.UserID, teamGameId, m.TeamGameRoleID, createdBy)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO al_user_team_game_role
				(user_id, team_game_id, role_id, assigned_at, assigned_by)
			VALUES `+strings.Join(roleValues, ","), roleArgs...); err != nil {
			return 0, err
		}

		if err := r.AddMembershipHistory(ctx, tx, teamGameId, gameID, "SIGNED", createdBy, members); err != nil {
			return 0, err
		}
	}

	for userID, path := range jerseyPaths {
		if _, err := tx.Exec(ctx, `
			UPDATE al_team_game_player_image
			SET is_current = FALSE
			WHERE user_id = $1 AND team_game_id = $2
		`, userID, teamGameId); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO al_team_game_player_image
				(user_id, team_game_id, player_jersey_image_path, is_current, created_by)
			VALUES ($1, $2, $3, TRUE, $4)
		`, userID, teamGameId, path, createdBy); err != nil {
			return 0, err
		}
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamGameId); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return teamGameId, nil
}

func (r *Repository) AddMembershipHistory(
	ctx context.Context,
	tx pgx.Tx,
	teamGameId int64,
	gameId int32,
	transferType string,
	addedBy int64,
	members []ReqTeamGameMember,
) error {

	var (
		historyValues []string
		historyArgs   []any
	)

	for i, m := range members {
		base := i * 5
		historyValues = append(historyValues,
			fmt.Sprintf("($%d,$%d,$%d,NOW(),$%d,$%d,NOW())", base+1, base+2, base+3, base+4, base+5),
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
			(user_id, team_game_id, game_id, joined_at, added_by, transfer_type, created_at)
		VALUES ` + strings.Join(historyValues, ",")

	_, err := tx.Exec(ctx, queryHistory, historyArgs...)
	return err
}

func (r *Repository) LeaveTeamGame(ctx context.Context, userID, teamGameID int64) error {
	transferType, err := r.determineTransferTypeByLeave(ctx, userID, teamGameID)
	if err != nil {
		return err
	}
	gameID := r.GetGameIdFromTeamGameId(ctx, teamGameID)
	if gameID == -1 {
		return rules.ErrCannotFindGame
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = r.deleteUsersFromTeamGame(ctx, tx, []int64{userID}, teamGameID); err != nil {
		return err
	}

	history := MembershipHistoryParams{
		UserID:       userID,
		TeamGameID:   teamGameID,
		GameID:       gameID,
		TransferType: transferType,
		RemovedBy:    &userID,
		LeftAt:       utils.Ptr(time.Now()),
	}

	if err := r.InsertMembershipHistory(ctx, tx, []MembershipHistoryParams{history}); err != nil {
		return err
	}

	if err := r.recomputeTeamGameCountryRegion(ctx, tx, teamGameID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) determineTransferTypeByLeave(
	ctx context.Context,
	userID, teamGameID int64,
) (string, error) {

	var (
		hasFutureTournament bool
		hasActiveTournament bool
	)

	err := r.db.QueryRow(ctx, `
		SELECT
			EXISTS (
				SELECT 1
				FROM al_tournament at
				JOIN al_tournament_roster atr
					ON atr.tournament_id = at.id
				WHERE atr.user_id = $1
				  AND atr.team_game_id = $2
				  AND at.start_at > now()
			) AS has_future,
			EXISTS (
				SELECT 1
				FROM al_tournament at
				JOIN al_tournament_roster atr
					ON atr.tournament_id = at.id
				WHERE atr.user_id = $1
				  AND atr.team_game_id = $2
				  AND now() BETWEEN at.start_at AND at.end_at
			) AS has_active
	`, userID, teamGameID).Scan(&hasFutureTournament, &hasActiveTournament)

	if err != nil {
		return "", err
	}

	if hasActiveTournament {
		return "LEFT_DURING_TOURNAMENT", nil
	}

	if hasFutureTournament {
		return "LEFT_PRE_TOURNAMENT", nil
	}

	return "LEFT", nil
}

func (r *Repository) deleteUsersFromTeamGame(
	ctx context.Context,
	tx pgx.Tx,
	userIDs []int64,
	teamGameID int64,
) error {
	if len(userIDs) == 0 {
		return nil
	}

	_, err := tx.Exec(ctx, `
		DELETE FROM al_user_team_game_role
		WHERE team_game_id = $1
		  AND user_id = ANY($2)
	`, teamGameID, userIDs)

	return err
}

func (r *Repository) InsertMembershipHistory(
	ctx context.Context,
	tx pgx.Tx,
	params []MembershipHistoryParams,
) error {

	if len(params) == 0 {
		return nil
	}

	var (
		values []string
		args   []any
	)

	for i, p := range params {
		base := i * 8

		values = append(values,
			fmt.Sprintf(
				"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,NOW(),NOW())",
				base+1,
				base+2,
				base+3,
				base+4,
				base+5,
				base+6,
				base+7,
				base+8,
			),
		)

		args = append(args,
			p.UserID,
			p.TeamGameID,
			p.GameID,
			p.JoinedAt,
			p.LeftAt,
			p.AddedBy,
			p.RemovedBy,
			p.TransferType,
		)
	}

	query := `
		INSERT INTO al_team_membership_history
			(user_id, team_game_id, game_id, joined_at, left_at,
			 added_by, removed_by, transfer_type, created_at, updated_at)
		VALUES ` + strings.Join(values, ",")

	_, err := tx.Exec(ctx, query, args...)
	return err
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

func (r *Repository) FilterTeamGames(
	ctx context.Context,
	filter TeamGameListFilter,
	page, limit int,
) ([]TeamGameListItem, int64, error) {
	where := []string{"tg.active = true"}
	args := []any{}
	argIdx := 1

	if filter.Name != nil && *filter.Name != "" {
		where = append(where, fmt.Sprintf("t.name ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, *filter.Name)
		argIdx++
	}
	if filter.RegionID != nil {
		where = append(where, fmt.Sprintf("tg.region_id = $%d", argIdx))
		args = append(args, *filter.RegionID)
		argIdx++
	}
	if filter.CountryID != nil {
		where = append(where, fmt.Sprintf("tg.country_id = $%d", argIdx))
		args = append(args, *filter.CountryID)
		argIdx++
	}
	if filter.GameID != nil {
		where = append(where, fmt.Sprintf("tg.game_id = $%d", argIdx))
		args = append(args, *filter.GameID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		WHERE %s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "tg.created_at DESC"
	if filter.SortBy != nil {
		switch *filter.SortBy {
		case "name_asc":
			orderBy = "t.name ASC"
		case "name_desc":
			orderBy = "t.name DESC"
		case "newest":
			orderBy = "tg.created_at DESC"
		case "oldest":
			orderBy = "tg.created_at ASC"
		}
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT
			tg.id,
			tg.game_id,
			t.logo_key,
			t.name AS team_name,
			g.name AS game_name,
			r.name AS region_name,
			r.iso_code AS region_iso,
			c.name AS country_name,
			tg.active,
			tg.created_at
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		JOIN al_game g ON tg.game_id = g.id
		LEFT JOIN al_region r ON tg.region_id = r.id
		LEFT JOIN al_country c ON tg.country_id = c.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []TeamGameListItem
	for rows.Next() {
		var item TeamGameListItem
		if err := rows.Scan(
			&item.ID,
			&item.GameID,
			&item.LogoKey,
			&item.TeamName,
			&item.GameName,
			&item.RegionName,
			&item.RegionISO,
			&item.CountryName,
			&item.Active,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *Repository) DeactivateTeamGame(ctx context.Context, teamGameID, updatedBy int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE al_team_game
		SET active = false, updated_at = NOW(), updated_by = $2
		WHERE id = $1 AND active = true
	`, teamGameID, updatedBy)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return rules.ErrBadIdTeamGame
	}
	return nil
}

func (r *Repository) ActivateTeamGame(ctx context.Context, teamGameID, updatedBy int64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE al_team_game
		SET active = true, updated_at = NOW(), updated_by = $2
		WHERE id = $1 AND active = false
	`, teamGameID, updatedBy)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return rules.ErrBadIdTeamGame
	}
	return nil
}

// recomputeTeamGameCountryRegion determines the team game (roster) country
// and/or region based on the majority of its players (PLAYER + TEAM_CAPTAIN):
//   - if more than half of the players share a country -> country_id is set
//   - otherwise, if more than half share a region -> region_id is set
//   - otherwise both are left NULL
func (r *Repository) recomputeTeamGameCountryRegion(ctx context.Context, tx pgx.Tx, teamGameID int64) error {
	rows, err := tx.Query(ctx, `
		SELECT au.country_id, ac.region_id
		FROM al_user au
		JOIN al_user_team_game_role utgr ON utgr.user_id = au.id
		JOIN al_team_game_role tgr ON utgr.role_id = tgr.id
		LEFT JOIN al_country ac ON ac.id = au.country_id
		WHERE utgr.team_game_id = $1
		  AND tgr.name IN ('PLAYER', 'TEAM_CAPTAIN')
	`, teamGameID)
	if err != nil {
		return err
	}

	countryCounts := map[int64]int{}
	regionCounts := map[int64]int{}
	for rows.Next() {
		var countryID, regionID sql.NullInt64
		if err := rows.Scan(&countryID, &regionID); err != nil {
			rows.Close()
			return err
		}
		if countryID.Valid {
			countryCounts[countryID.Int64]++
		}
		if regionID.Valid {
			regionCounts[regionID.Int64]++
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	total := 0
	for _, count := range countryCounts {
		total += count
	}

	countryID, regionID := resolveMajorityCountryRegion(countryCounts, regionCounts, total)

	_, err = tx.Exec(ctx, `
		UPDATE al_team_game
		SET country_id = $2, region_id = $3, updated_at = NOW()
		WHERE id = $1
	`, teamGameID, countryID, regionID)
	return err
}

// resolveMajorityCountryRegion decides the roster flag:
//   - if a country has more than half of the players -> country wins
//   - otherwise, if a region has more than half -> region wins
//   - otherwise both stay NULL
func resolveMajorityCountryRegion(countryCounts, regionCounts map[int64]int, total int) (sql.NullInt64, sql.NullInt64) {
	var countryID, regionID sql.NullInt64
	if total <= 0 {
		return countryID, regionID
	}
	if id, ok := majorityKey(countryCounts, total); ok {
		countryID = sql.NullInt64{Int64: id, Valid: true}
	} else if id, ok := majorityKey(regionCounts, total); ok {
		regionID = sql.NullInt64{Int64: id, Valid: true}
	}
	return countryID, regionID
}

func majorityKey(counts map[int64]int, total int) (int64, bool) {
	for id, count := range counts {
		if count*2 > total {
			return id, true
		}
	}
	return 0, false
}

func (r *Repository) GetTeamGamesByTeam(ctx context.Context, teamId int64) ([]TeamGameListItem, error) {
	query := `
		SELECT
			tg.id,
			tg.game_id,
			t.logo_key,
			t.name AS team_name,
			g.name AS game_name,
			r.name AS region_name,
			r.iso_code AS region_iso,
			c.name AS country_name,
			tg.active,
			tg.created_at
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		JOIN al_game g ON tg.game_id = g.id
		LEFT JOIN al_region r ON tg.region_id = r.id
		LEFT JOIN al_country c ON tg.country_id = c.id
		WHERE tg.team_id = $1
		ORDER BY tg.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TeamGameListItem
	for rows.Next() {
		var item TeamGameListItem
		if err := rows.Scan(
			&item.ID,
			&item.GameID,
			&item.LogoKey,
			&item.TeamName,
			&item.GameName,
			&item.RegionName,
			&item.RegionISO,
			&item.CountryName,
			&item.Active,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if items == nil {
		items = []TeamGameListItem{}
	}

	return items, rows.Err()
}
