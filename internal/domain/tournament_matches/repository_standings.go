package matches

import "context"

// ListStandings returns the accepted registrations with their win/loss
// record across the stage's completed matches.
func (r *Repository) ListStandings(ctx context.Context, stageID, tournamentID int64) ([]StandingsRow, error) {
	rows, err := r.db.Query(ctx, `
		WITH participants AS (
			SELECT r.id, COALESCE(t.name, '') AS team_name
			FROM al_tournament_registration r
			JOIN al_team_game tg ON tg.id = r.team_game_id
			JOIN al_team t ON t.id = tg.team_id
			WHERE r.tournament_id = $2 AND r.status = 'ACCEPTED' AND r.deleted_at IS NULL
		),
		results AS (
			SELECT team_a_registration_id AS reg, team_b_registration_id AS other, winner_registration_id
			FROM al_tournament_match
			WHERE stage_id = $1 AND deleted_at IS NULL AND status IN ('COMPLETED', 'WALKOVER')
			UNION ALL
			SELECT team_b_registration_id, team_a_registration_id, winner_registration_id
			FROM al_tournament_match
			WHERE stage_id = $1 AND deleted_at IS NULL AND status IN ('COMPLETED', 'WALKOVER')
		)
		SELECT
			p.id, p.team_name,
			count(r.other) FILTER (WHERE r.winner_registration_id = p.id) AS wins,
			count(r.other) FILTER (WHERE r.winner_registration_id IS NOT NULL AND r.winner_registration_id <> p.id) AS losses
		FROM participants p
		LEFT JOIN results r ON r.reg = p.id
		GROUP BY p.id, p.team_name
		ORDER BY wins DESC, p.id
	`, stageID, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []StandingsRow{}
	for rows.Next() {
		var s StandingsRow
		var wins, losses int
		if err := rows.Scan(&s.RegistrationID, &s.TeamName, &wins, &losses); err != nil {
			return nil, err
		}
		s.Wins = wins
		s.Losses = losses
		out = append(out, s)
	}
	return out, rows.Err()
}

// PlayedPairs returns the set of pairings that already happened in the stage.
func (r *Repository) PlayedPairs(ctx context.Context, stageID int64) (map[int64]map[int64]bool, error) {
	rows, err := r.db.Query(ctx, `
		SELECT team_a_registration_id, team_b_registration_id
		FROM al_tournament_match
		WHERE stage_id = $1 AND deleted_at IS NULL
		  AND team_a_registration_id IS NOT NULL AND team_b_registration_id IS NOT NULL
	`, stageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pairs := map[int64]map[int64]bool{}
	for rows.Next() {
		var a, b int64
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		if pairs[a] == nil {
			pairs[a] = map[int64]bool{}
		}
		if pairs[b] == nil {
			pairs[b] = map[int64]bool{}
		}
		pairs[a][b] = true
		pairs[b][a] = true
	}
	return pairs, rows.Err()
}
