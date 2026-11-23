package matches

import (
	"arena/internal/domain/tournament_matches/rules"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

// ListStageMatches returns the live matches of a stage.
func (r *Repository) ListStageMatches(ctx context.Context, stageID int64) ([]Match, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			m.id, m.stage_id, m.tournament_id, m.round_number, m.match_number,
			m.team_a_registration_id, m.team_b_registration_id, m.best_of,
			m.status, m.score_a, m.score_b, m.winner_registration_id,
			m.scheduled_at, m.created_at, m.updated_at
		FROM al_tournament_match m
		WHERE m.stage_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.round_number, m.match_number
	`, stageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

// GetMatchByID returns a live match.
func (r *Repository) GetMatchByID(ctx context.Context, id int64) (*Match, error) {
	var m Match
	err := r.db.QueryRow(ctx, `
		SELECT
			id, stage_id, tournament_id, round_number, match_number,
			team_a_registration_id, team_b_registration_id, best_of,
			status, score_a, score_b, winner_registration_id,
			scheduled_at, created_at, updated_at
		FROM al_tournament_match
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
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

func scanMatches(rows pgx.Rows) ([]Match, error) {
	matches := []Match{}
	for rows.Next() {
		var m Match
		if err := rows.Scan(
			&m.ID, &m.StageID, &m.TournamentID, &m.RoundNumber, &m.MatchNumber,
			&m.TeamA, &m.TeamB, &m.BestOf,
			&m.Status, &m.ScoreA, &m.ScoreB, &m.Winner,
			&m.ScheduledAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}
