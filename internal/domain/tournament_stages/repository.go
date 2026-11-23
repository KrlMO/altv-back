package stages

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"arena/internal/domain/tournament_stages/rules"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

// GetTournamentInfo returns the tournament fields needed by the rules.
func (r *Repository) GetTournamentInfo(ctx context.Context, tournamentID int64) (*TournamentInfo, error) {
	var t TournamentInfo
	err := r.db.QueryRow(ctx, `
		SELECT id, organization_id
		FROM al_tournament
		WHERE id = $1
	`, tournamentID).Scan(&t.ID, &t.OrganizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdStage
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetFormatIDByCode resolves an active stage format by its code.
func (r *Repository) GetFormatIDByCode(ctx context.Context, code string) (int32, error) {
	var id int32
	err := r.db.QueryRow(ctx, `
		SELECT id FROM al_stage_format
		WHERE code = $1 AND active AND deleted_at IS NULL
	`, code).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, rules.ErrUnknownFormat
	}
	return id, err
}

// CreateStage inserts a stage.
func (r *Repository) CreateStage(ctx context.Context, s *Stage) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO al_tournament_stage
			(tournament_id, format_id, name, stage_order, status, config, created_by)
		VALUES ($1, $2, $3, $4, 'PENDING', $5, $6)
		RETURNING id
	`, s.TournamentID, s.FormatID, s.Name, s.StageOrder, s.Config, s.CreatedBy).Scan(&id)
	if err != nil {
		if isUniqueViolation(err, "tournament_stage_order_unique") {
			return 0, errors.New("a stage with this order already exists for the tournament")
		}
		return 0, err
	}
	return id, nil
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

// GetStageByID returns a live stage with its format info.
func (r *Repository) GetStageByID(ctx context.Context, id int64) (*Stage, error) {
	var s Stage
	err := r.db.QueryRow(ctx, `
		SELECT
			ts.id, ts.tournament_id, ts.format_id,
			f.code AS format_code, f.name AS format_name,
			ts.name, ts.stage_order, ts.status, ts.config,
			ts.created_at, ts.updated_at, ts.deleted_at
		FROM al_tournament_stage ts
		JOIN al_stage_format f ON f.id = ts.format_id
		WHERE ts.id = $1 AND ts.deleted_at IS NULL
	`, id).Scan(
		&s.ID, &s.TournamentID, &s.FormatID,
		&s.FormatCode, &s.FormatName,
		&s.Name, &s.StageOrder, &s.Status, &s.Config,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, rules.ErrBadIdStage
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListStagesByTournament returns the live stages of a tournament, ordered.
func (r *Repository) ListStagesByTournament(ctx context.Context, tournamentID int64) ([]Stage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ts.id, ts.tournament_id, ts.format_id,
			f.code AS format_code, f.name AS format_name,
			ts.name, ts.stage_order, ts.status, ts.config,
			ts.created_at, ts.updated_at, ts.deleted_at
		FROM al_tournament_stage ts
		JOIN al_stage_format f ON f.id = ts.format_id
		WHERE ts.tournament_id = $1 AND ts.deleted_at IS NULL
		ORDER BY ts.stage_order
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stages := []Stage{}
	for rows.Next() {
		var s Stage
		if err := rows.Scan(
			&s.ID, &s.TournamentID, &s.FormatID,
			&s.FormatCode, &s.FormatName,
			&s.Name, &s.StageOrder, &s.Status, &s.Config,
			&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
		); err != nil {
			return nil, err
		}
		stages = append(stages, s)
	}
	return stages, rows.Err()
}

// UpdateStage patches a pending stage.
func (r *Repository) UpdateStage(ctx context.Context, id int64, name *string, stageOrder *int32, config []byte, updatedBy int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_stage
		SET
			name = COALESCE($2, name),
			stage_order = COALESCE($3, stage_order),
			config = COALESCE($4, config),
			updated_at = NOW(),
			updated_by = $5
		WHERE id = $1 AND deleted_at IS NULL AND status = 'PENDING'
	`, id, name, stageOrder, config, updatedBy)
	if err != nil {
		if isUniqueViolation(err, "tournament_stage_order_unique") {
			return errors.New("a stage with this order already exists for the tournament")
		}
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrStageNotPending
	}
	return nil
}

// TransitionStageStatus moves a stage forward in its lifecycle.
func (r *Repository) TransitionStageStatus(ctx context.Context, id int64, current, next string, updatedBy int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_stage
		SET status = $3, updated_at = NOW(), updated_by = $4
		WHERE id = $1 AND status = $2 AND deleted_at IS NULL
	`, id, current, next, updatedBy)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrStageNotPending
	}
	return nil
}

// PreviousStageCompleted reports whether the stage immediately before the
// given order is completed (always true for the first stage).
func (r *Repository) PreviousStageCompleted(ctx context.Context, tournamentID, stageOrder int64) (bool, error) {
	var completed bool
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(
			(bool_and(status = 'COMPLETED')) OR count(*) = 0, true
		)
		FROM al_tournament_stage
		WHERE tournament_id = $1 AND stage_order < $2 AND deleted_at IS NULL
	`, tournamentID, stageOrder).Scan(&completed)
	return completed, err
}

// DeleteStage soft-deletes a pending stage.
func (r *Repository) DeleteStage(ctx context.Context, id, deletedBy int64) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE al_tournament_stage
		SET deleted_at = NOW(), deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL AND status = 'PENDING'
	`, id, deletedBy)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return rules.ErrStageNotPending
	}
	return nil
}
