package stages

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"arena/internal/domain/identity"
	"arena/internal/domain/tournament_stages/rules"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

func toResponse(s *Stage) StageResponse {
	cfg := map[string]interface{}{}
	if len(s.Config) > 0 {
		_ = json.Unmarshal(s.Config, &cfg)
	}
	return StageResponse{
		ID:           s.ID,
		TournamentID: s.TournamentID,
		FormatCode:   s.FormatCode,
		FormatName:   s.FormatName,
		Name:         s.Name,
		StageOrder:   s.StageOrder,
		Status:       s.Status,
		Config:       cfg,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// buildPermissions resolves the actor's org and tournament roles and checks
// stage management permission.
func (s *Service) buildPermissions(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	tournamentID int64,
) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	tournament, err := s.repo.GetTournamentInfo(ctx, tournamentID)
	if err != nil {
		return err
	}
	orgRoles, err := s.repo.GetUserOrganizationRoles(ctx, actorID, tournament.OrganizationID)
	if err != nil {
		return err
	}
	tournamentRoles, err := s.repo.GetUserTournamentRoles(ctx, actorID, tournament.ID)
	if err != nil {
		return err
	}
	return rules.CanManageStages(orgRoles, tournamentRoles, actorGlobalRoles)
}

// encodeConfig marshals and validates the stage config against its format.
func encodeConfig(formatCode string, cfg map[string]interface{}) ([]byte, error) {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	if err := rules.ValidateConfig(formatCode, cfg); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.New("invalid config")
	}
	return raw, nil
}

// CreateStage creates a stage for a tournament.
func (s *Service) CreateStage(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	tournamentID int64,
	req CreateStageRequest,
) (int64, error) {
	if err := s.buildPermissions(ctx, actorID, actorGlobalRoles, tournamentID); err != nil {
		return 0, err
	}

	formatID, err := s.repo.GetFormatIDByCode(ctx, req.FormatCode)
	if err != nil {
		return 0, err
	}

	config, err := encodeConfig(req.FormatCode, req.Config)
	if err != nil {
		return 0, err
	}

	stage := &Stage{
		TournamentID: tournamentID,
		FormatID:     formatID,
		Name:         req.Name,
		StageOrder:   req.StageOrder,
		Config:       config,
		CreatedBy:    actorID,
	}
	return s.repo.CreateStage(ctx, stage)
}

// ListTournamentStages returns the stages of a tournament (public view).
func (s *Service) ListTournamentStages(ctx context.Context, tournamentID int64) ([]StageResponse, error) {
	if _, err := s.repo.GetTournamentInfo(ctx, tournamentID); err != nil {
		return nil, err
	}
	stages, err := s.repo.ListStagesByTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	out := make([]StageResponse, 0, len(stages))
	for i := range stages {
		out = append(out, toResponse(&stages[i]))
	}
	return out, nil
}

// UpdateStage patches a pending stage.
func (s *Service) UpdateStage(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	stageID int64,
	req UpdateStageRequest,
) error {
	stage, err := s.repo.GetStageByID(ctx, stageID)
	if err != nil {
		return err
	}
	if stage.Status != "PENDING" {
		return rules.ErrStageNotPending
	}
	if err := s.buildPermissions(ctx, actorID, actorGlobalRoles, stage.TournamentID); err != nil {
		return err
	}

	var config []byte
	if req.Config != nil {
		formatCode := stage.FormatCode
		// Config must stay valid for the stage's format.
		if err := rules.ValidateConfig(formatCode, req.Config); err != nil {
			return err
		}
		config, err = json.Marshal(req.Config)
		if err != nil {
			return errors.New("invalid config")
		}
	}

	return s.repo.UpdateStage(ctx, stageID, req.Name, req.StageOrder, config, actorID)
}

// TransitionStageStatus moves a stage forward (PENDING -> IN_PROGRESS ->
// COMPLETED). A stage cannot start while earlier stages are not completed.
func (s *Service) TransitionStageStatus(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	stageID int64,
	next string,
) error {
	stage, err := s.repo.GetStageByID(ctx, stageID)
	if err != nil {
		return err
	}
	if err := s.buildPermissions(ctx, actorID, actorGlobalRoles, stage.TournamentID); err != nil {
		return err
	}
	if err := rules.CanTransitionStageStatus(stage.Status, next); err != nil {
		return err
	}
	completed, err := s.repo.PreviousStageCompleted(ctx, stage.TournamentID, int64(stage.StageOrder))
	if err != nil {
		return err
	}
	if !completed {
		return rules.ErrPreviousStageOpen
	}
	return s.repo.TransitionStageStatus(ctx, stageID, stage.Status, next, actorID)
}

// DeleteStage soft-deletes a pending stage.
func (s *Service) DeleteStage(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	stageID int64,
) error {
	stage, err := s.repo.GetStageByID(ctx, stageID)
	if err != nil {
		return err
	}
	if stage.Status != "PENDING" {
		return rules.ErrStageNotPending
	}
	if err := s.buildPermissions(ctx, actorID, actorGlobalRoles, stage.TournamentID); err != nil {
		return err
	}
	return s.repo.DeleteStage(ctx, stageID, actorID)
}

func ParseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

// ListStageFormats returns the active stage formats of the catalog.
func (s *Service) ListStageFormats(ctx context.Context) ([]StageFormatDTO, error) {
	return s.repo.ListStageFormats(ctx)
}
