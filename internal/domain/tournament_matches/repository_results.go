package matches

import (
	"arena/internal/domain/tournament_matches/rules"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

// GetNextBracketSlot resolves where the winner of a match goes.
func (r *Repository) GetNextBracketSlot(ctx context.Context, m *Match) (*Match, string, error) {
	var next Match
	err := r.db.QueryRow(ctx, `
		SELECT
			id, stage_id, tournament_id, round_number, match_number,
			team_a_registration_id, team_b_registration_id, best_of,
			status, score_a, score_b, winner_registration_id,
			scheduled_at, created_at, updated_at
		FROM al_tournament_match
		WHERE stage_id = $1 AND round_number = $2 AND match_number = $3
		  AND deleted_at IS NULL
	`, m.StageID, m.RoundNumber+1, (m.MatchNumber+1)/2).Scan(
		&next.ID, &next.StageID, &next.TournamentID, &next.RoundNumber, &next.MatchNumber,
		&next.TeamA, &next.TeamB, &next.BestOf,
		&next.Status, &next.ScoreA, &next.ScoreB, &next.Winner,
		&next.ScheduledAt, &next.CreatedAt, &next.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", pgx.ErrNoRows
	}
	if err != nil {
		return nil, "", err
	}
	if m.MatchNumber%2 == 1 {
		return &next, "team_a_registration_id", nil
	}
	return &next, "team_b_registration_id", nil
}

// SetBracketSlot writes the winner into a slot of the next match.
func (r *Repository) SetBracketSlot(ctx context.Context, nextMatchID int64, slot string, registrationID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET `+slot+` = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, nextMatchID, registrationID)
	return err
}
