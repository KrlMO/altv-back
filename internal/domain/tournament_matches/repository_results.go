package matches

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"arena/internal/domain/tournament_matches/rules"
)

// SetMatchResult stores the result of one match.
func (r *Repository) SetMatchResult(ctx context.Context, id, winner int64, scoreA, scoreB int32, updatedBy int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET status = 'COMPLETED', score_a = $2, score_b = $3,
		    winner_registration_id = $4, updated_at = NOW(), updated_by = $5
		WHERE id = $1 AND deleted_at IS NULL AND status <> 'COMPLETED'
	`, id, scoreA, scoreB, winner, updatedBy)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrMatchNotPlayable
	}
	return nil
}

// GetMatchByRoundAndNumber returns a live match by its position in the stage.
func (r *Repository) GetMatchByRoundAndNumber(ctx context.Context, stageID, roundNumber, matchNumber int64) (*Match, error) {
	var m Match
	err := r.db.QueryRow(ctx, `
		SELECT
			id, stage_id, tournament_id, round_number, match_number,
			team_a_registration_id, team_b_registration_id, best_of,
			status, score_a, score_b, winner_registration_id,
			scheduled_at, created_at, updated_at
		FROM al_tournament_match
		WHERE stage_id = $1 AND round_number = $2 AND match_number = $3
		  AND deleted_at IS NULL
	`, stageID, roundNumber, matchNumber).Scan(
		&m.ID, &m.StageID, &m.TournamentID, &m.RoundNumber, &m.MatchNumber,
		&m.TeamA, &m.TeamB, &m.BestOf,
		&m.Status, &m.ScoreA, &m.ScoreB, &m.Winner,
		&m.ScheduledAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdMatch
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SetBracketSlot writes a registration into slot "a" or "b" of a match.
func (r *Repository) SetBracketSlot(ctx context.Context, nextMatchID int64, slot string, registrationID int64) error {
	column := "team_a_registration_id"
	if slot == "b" {
		column = "team_b_registration_id"
	}
	_, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET `+column+` = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, nextMatchID, registrationID)
	return err
}
