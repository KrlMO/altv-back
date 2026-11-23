package registrations

import (
	"arena/internal/domain/tournament_registrations/rules"
	"arena/internal/shared"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db         *pgxpool.Pool
	sharedRepo *shared.SharedRepository
}

func NewRepository(db *pgxpool.Pool, sharedRepo *shared.SharedRepository) *Repository {
	return &Repository{db, sharedRepo}
}

// GetTournamentInfo returns the tournament fields needed by the rules.
func (r *Repository) GetTournamentInfo(ctx context.Context, tournamentID int64) (*TournamentInfo, error) {
	var t TournamentInfo
	err := r.db.QueryRow(ctx, `
		SELECT id, game_id, status, organization_id, start_at, end_at
		FROM al_tournament
		WHERE id = $1
	`, tournamentID).Scan(
		&t.ID, &t.GameID, &t.Status, &t.OrganizationID, &t.StartAt, &t.EndAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdRegistration
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTeamGameInfo returns the roster's team and game.
func (r *Repository) GetTeamGameInfo(ctx context.Context, teamGameID int64) (teamID int64, gameID int32, active bool, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT team_id, game_id, COALESCE(active, true)
		FROM al_team_game
		WHERE id = $1
	`, teamGameID).Scan(&teamID, &gameID, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, false, rules.ErrTeamGameInactive
	}
	return
}

// GetUserTeamRoles returns the actor's roles within a team.
func (r *Repository) GetUserTeamRoles(ctx context.Context, userID, teamID int64) ([]string, error) {
	return r.sharedRepo.GetUserTeamRoles(ctx, userID, teamID)
}

// GetUserTeamGameRoles returns the actor's roles within a roster.
func (r *Repository) GetUserTeamGameRoles(ctx context.Context, userID, teamGameID int64) ([]string, error) {
	roles := []string{}
	rows, err := r.db.Query(ctx, `
		SELECT tr.name
		FROM al_user_team_game_role utr
		JOIN al_team_game_role tr ON utr.role_id = tr.id
		WHERE utr.user_id = $1 AND utr.team_game_id = $2
	`, userID, teamGameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, rows.Err()
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
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
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
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, rows.Err()
}

// IsTeamGameMember reports whether the user belongs to the roster.
func (r *Repository) IsTeamGameMember(ctx context.Context, userID, teamGameID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_user_team_game_role
			WHERE user_id = $1 AND team_game_id = $2
		)
	`, userID, teamGameID).Scan(&exists)
	return exists, err
}

// CreateRegistration registers a roster to a tournament as PENDING.
func (r *Repository) CreateRegistration(ctx context.Context, tournamentID, teamGameID, registeredBy int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO al_tournament_registration
			(tournament_id, team_game_id, status, registered_by, created_by)
		VALUES ($1, $2, 'PENDING', $3, $3)
		RETURNING id
	`, tournamentID, teamGameID, registeredBy).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "tournament_registration_unique") {
			return 0, rules.ErrAlreadyRegistered
		}
		return 0, err
	}
	return id, nil
}

// AcceptRegistration marks a pending registration as accepted.
func (r *Repository) AcceptRegistration(ctx context.Context, id, reviewerID int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_registration
		SET status = 'ACCEPTED', reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW(), updated_by = $2
		WHERE id = $1 AND status = 'PENDING' AND deleted_at IS NULL
	`, id, reviewerID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrRegistrationNotPending
	}
	return nil
}

// GetRegistrationByID returns a live registration.
func (r *Repository) GetRegistrationByID(ctx context.Context, id int64) (*Registration, error) {
	var reg Registration
	err := r.db.QueryRow(ctx, `
		SELECT id, tournament_id, team_game_id, status, registered_by,
		       reviewed_by, reviewed_at, created_at, updated_at
		FROM al_tournament_registration
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&reg.ID, &reg.TournamentID, &reg.TeamGameID, &reg.Status,
		&reg.RegisteredBy, &reg.ReviewedBy, &reg.ReviewedAt,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdRegistration
	}
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// DeleteRegistration removes the registration; the roster rows cascade.
func (r *Repository) DeleteRegistration(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `
		DELETE FROM al_tournament_registration WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrBadIdRegistration
	}
	return nil
}

// ListRegistrationsByTournament lists the live registrations of a tournament.
// If acceptedOnly, only ACCEPTED ones are returned (public view).
func (r *Repository) ListRegistrationsByTournament(ctx context.Context, tournamentID int64, acceptedOnly bool) ([]RegistrationResponse, error) {
	where := "r.tournament_id = $1 AND r.deleted_at IS NULL"
	if acceptedOnly {
		where += " AND r.status = 'ACCEPTED'"
	}
	query := `
		SELECT
			r.id, r.tournament_id, r.team_game_id, r.status,
			r.registered_by, r.reviewed_by, r.reviewed_at, r.created_at,
			COALESCE(t.name, '') AS team_name,
			COALESCE(g.name, '') AS game_name
		FROM al_tournament_registration r
		JOIN al_team_game tg ON tg.id = r.team_game_id
		JOIN al_team t ON t.id = tg.team_id
		JOIN al_game g ON g.id = tg.game_id
		WHERE ` + where + `
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RegistrationResponse{}
	for rows.Next() {
		var it RegistrationResponse
		if err := rows.Scan(
			&it.ID, &it.TournamentID, &it.TeamGameID, &it.Status,
			&it.RegisteredBy, &it.ReviewedBy, &it.ReviewedAt, &it.CreatedAt,
			&it.TeamName, &it.GameName,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
