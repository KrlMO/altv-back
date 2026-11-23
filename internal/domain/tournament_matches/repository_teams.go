package matches

import (
	"context"

	"arena/internal/domain/tournament_matches/rules"
)

// SetMatchTeams writes the provided slots of a match.
func (r *Repository) SetMatchTeams(ctx context.Context, id int64, teamA, teamB int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET
			team_a_registration_id = COALESCE(NULLIF($2, 0), team_a_registration_id),
			team_b_registration_id = COALESCE(NULLIF($3, 0), team_b_registration_id),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL AND status = 'PENDING'
	`, id, teamA, teamB)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrMatchNotEditable
	}
	return nil
}

// SetMatchTeam writes a single slot of a match (IN_PROGRESS friendly).
func (r *Repository) SetMatchTeam(ctx context.Context, id int64, slot string, registrationID int64) error {
	column := "team_a_registration_id"
	if slot == "b" {
		column = "team_b_registration_id"
	}
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_match
		SET `+column+` = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, registrationID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrMatchNotEditable
	}
	return nil
}

// RegistrationBusyInRound reports whether the registration already plays in
// another match of the same round (excluding the given match).
func (r *Repository) RegistrationBusyInRound(ctx context.Context, stageID, roundNumber, registrationID, excludeMatchID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_tournament_match
			WHERE stage_id = $1 AND round_number = $2
			  AND id <> $3 AND deleted_at IS NULL
			  AND (team_a_registration_id = $4 OR team_b_registration_id = $4)
		)
	`, stageID, roundNumber, excludeMatchID, registrationID).Scan(&exists)
	return exists, err
}
