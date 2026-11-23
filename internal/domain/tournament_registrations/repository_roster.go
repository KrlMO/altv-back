package registrations

import (
	"arena/internal/domain/tournament_registrations/rules"
	"context"
)

// IsRosterMemberOfTournament reports whether the user is in any live roster
// of the tournament (its own registration included).
func (r *Repository) IsRosterMemberOfTournament(ctx context.Context, userID, tournamentID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_tournament_roster
			WHERE user_id = $1 AND tournament_id = $2
		)
	`, userID, tournamentID).Scan(&exists)
	return exists, err
}

// AddRosterMember adds a player to a registration's roster.
func (r *Repository) AddRosterMember(ctx context.Context, registrationID, tournamentID, userID, addedBy int64, isStarter bool) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO al_tournament_roster
			(registration_id, tournament_id, user_id, team_game_id, is_starter, added_by)
		SELECT id, tournament_id, $3, team_game_id, $4, $5
		FROM al_tournament_registration WHERE id = $1 AND tournament_id = $2
	`, registrationID, tournamentID, userID, isStarter, addedBy)
	if err != nil {
		if isUniqueViolation(err, "tournament_roster_unique_per_tournament") {
			return rules.ErrUserInAnotherRoster
		}
		if isUniqueViolation(err, "tournament_roster_pkey") {
			return rules.ErrUserInAnotherRoster
		}
		return err
	}
	return nil
}

// RemoveRosterMember removes a player from a registration's roster.
func (r *Repository) RemoveRosterMember(ctx context.Context, registrationID, userID int64) error {
	ct, err := r.db.Exec(ctx, `
		DELETE FROM al_tournament_roster WHERE registration_id = $1 AND user_id = $2
	`, registrationID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrRosterMemberNotFound
	}
	return nil
}

// ListRosterMembers returns the players of a registration's roster.
func (r *Repository) ListRosterMembers(ctx context.Context, registrationID int64) ([]RosterMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			tr.registration_id, tr.user_id,
			u.nickname, u.first_name, u.last_name,
			tr.is_starter, pi.player_jersey_image_path, tr.added_at
		FROM al_tournament_roster tr
		JOIN al_user u ON u.id = tr.user_id
		LEFT JOIN al_team_game_player_image pi
			ON pi.user_id = tr.user_id
			AND pi.team_game_id = tr.team_game_id
			AND pi.is_current
		WHERE tr.registration_id = $1
		ORDER BY tr.is_starter DESC, tr.added_at
	`, registrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []RosterMember{}
	for rows.Next() {
		var m RosterMember
		if err := rows.Scan(
			&m.RegistrationID, &m.UserID,
			&m.Nickname, &m.FirstName, &m.LastName,
			&m.IsStarter, &m.JerseyImage, &m.AddedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
