package matches

import "context"

// GetTournamentInfo returns the tournament fields needed for permissions.
func (r *Repository) GetTournamentInfo(ctx context.Context, tournamentID int64) (id, organizationID int64, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT id, organization_id FROM al_tournament WHERE id = $1
	`, tournamentID).Scan(&id, &organizationID)
	return id, organizationID, err
}

// GetUserOrganizationRoles returns the actor's roles within an organization.
func (r *Repository) GetUserOrganizationRoles(ctx context.Context, userID, organizationID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT r.name
		FROM al_user_organization_role uor
		JOIN al_organization_role r ON r.id = uor.role_id
		WHERE uor.user_id = $1 AND uor.organization_id = $2
		  AND COALESCE(uor.active, true)
	`, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// GetUserTournamentRoles returns the actor's roles assigned for a tournament.
func (r *Repository) GetUserTournamentRoles(ctx context.Context, userID, tournamentID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT tr.name
		FROM al_user_tournament_role utr
		JOIN al_tournament_role tr ON tr.id = utr.role_id
		WHERE utr.user_id = $1 AND utr.tournament_id = $2
		  AND COALESCE(utr.active, true)
	`, userID, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// ListTeamLabels resolves registration ids to team names for a tournament.
func (r *Repository) ListTeamLabels(ctx context.Context, tournamentID int64) ([]TeamLabel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, COALESCE(t.name, '')
		FROM al_tournament_registration r
		JOIN al_team_game tg ON tg.id = r.team_game_id
		JOIN al_team t ON t.id = tg.team_id
		WHERE r.tournament_id = $1 AND r.deleted_at IS NULL
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := []TeamLabel{}
	for rows.Next() {
		var l TeamLabel
		if err := rows.Scan(&l.RegistrationID, &l.TeamName); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}
