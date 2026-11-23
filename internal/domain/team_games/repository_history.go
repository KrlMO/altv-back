package teamgames

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
