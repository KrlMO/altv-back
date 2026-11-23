package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/utils"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
