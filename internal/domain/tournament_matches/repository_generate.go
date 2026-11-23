package matches

import (
	"arena/internal/domain/tournament_matches/rules"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// GetStageInfo returns the stage fields needed to generate matches.
func (r *Repository) GetStageInfo(ctx context.Context, stageID int64) (*StageInfo, error) {
	var s StageInfo
	err := r.db.QueryRow(ctx, `
		SELECT
			ts.id, ts.tournament_id, f.code AS format_code, ts.status,
			(ts.config->>'best_of')::int AS best_of, ts.config
		FROM al_tournament_stage ts
		JOIN al_stage_format f ON f.id = ts.format_id
		WHERE ts.id = $1 AND ts.deleted_at IS NULL
	`, stageID).Scan(
		&s.ID, &s.TournamentID, &s.FormatCode, &s.Status,
		&s.BestOf, &s.Config,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdMatch
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListAcceptedParticipants returns the accepted registrations of the
// tournament, in registration order (seed order by default).
func (r *Repository) ListAcceptedParticipants(ctx context.Context, tournamentID int64) ([]Participant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, COALESCE(t.name, ''), COALESCE(g.name, '')
		FROM al_tournament_registration r
		JOIN al_team_game tg ON tg.id = r.team_game_id
		JOIN al_team t ON t.id = tg.team_id
		JOIN al_game g ON g.id = tg.game_id
		WHERE r.tournament_id = $1 AND r.status = 'ACCEPTED' AND r.deleted_at IS NULL
		ORDER BY r.id
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []Participant{}
	for rows.Next() {
		var p Participant
		if err := rows.Scan(&p.RegistrationID, &p.TeamName, &p.GameName); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, rows.Err()
}

// CountStageMatches returns how many live matches a stage has.
func (r *Repository) CountStageMatches(ctx context.Context, stageID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM al_tournament_match
		WHERE stage_id = $1 AND deleted_at IS NULL
	`, stageID).Scan(&n)
	return n, err
}

// InsertMatches persists the generated matches in one transaction.
func (r *Repository) InsertMatches(ctx context.Context, stageID, tournamentID, createdBy int64, matches []Match) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range matches {
		m := &matches[i]
		if _, err := tx.Exec(ctx, `
			INSERT INTO al_tournament_match
				(stage_id, tournament_id, round_number, match_number,
				 team_a_registration_id, team_b_registration_id, best_of,
				 status, winner_registration_id, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, stageID, tournamentID, m.RoundNumber, m.MatchNumber,
			m.TeamA, m.TeamB, m.BestOf, m.Status, m.Winner, createdBy,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListRoundMatches returns the live matches of one round.
func (r *Repository) ListRoundMatches(ctx context.Context, stageID, roundNumber int64) ([]Match, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, stage_id, tournament_id, round_number, match_number,
			team_a_registration_id, team_b_registration_id, best_of,
			status, score_a, score_b, winner_registration_id,
			scheduled_at, created_at, updated_at
		FROM al_tournament_match
		WHERE stage_id = $1 AND round_number = $2 AND deleted_at IS NULL
		ORDER BY match_number
	`, stageID, roundNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

// MaxRoundNumber returns the highest round of a stage (0 if none).
func (r *Repository) MaxRoundNumber(ctx context.Context, stageID int64) (int32, error) {
	var n int32
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(max(round_number), 0) FROM al_tournament_match
		WHERE stage_id = $1 AND deleted_at IS NULL
	`, stageID).Scan(&n)
	return n, err
}

// StageConfigRounds returns the swiss "rounds" config value (0 if unset).
func (r *Repository) StageConfigRounds(ctx context.Context, stageID int64) (int32, error) {
	var n int32
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE((config->>'rounds')::int, 0)
		FROM al_tournament_stage WHERE id = $1
	`, stageID).Scan(&n)
	return n, err
}

// IsSwissStageActive reports whether the stage still has rounds pending
// according to its config and current round count.
func (r *Repository) IsSwissStageActive(ctx context.Context, stageID int64) (bool, error) {
	var pending int
	err := r.db.QueryRow(ctx, `
		SELECT count(*)
		FROM al_tournament_match
		WHERE stage_id = $1 AND deleted_at IS NULL AND status = 'PENDING'
	`, stageID).Scan(&pending)
	if err != nil {
		return false, err
	}
	return pending == 0, nil
}

// SetMatchStatusesToInProgress marks pending matches of a round as in
// progress when the round starts.
func (r *Repository) SetMatchStatusesToInProgress(ctx context.Context, stageID, roundNumber int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET status = 'IN_PROGRESS', updated_at = NOW()
		WHERE stage_id = $1 AND round_number = $2 AND status = 'PENDING' AND deleted_at IS NULL
	`, stageID, roundNumber)
	return err
}
