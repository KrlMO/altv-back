package tournaments

import (
	"context"
	"errors"
	"strings"
	"time"

	"arena/internal/domain/tournaments/rules"
)

// CreateTournamentRequest is the payload to create a tournament.
type CreateTournamentRequest struct {
	OrganizationID int64
	GameID         int32
	Name           string
	Description    *string
	StartAt        time.Time
	EndAt          *time.Time
	ColorHex       *string
	CountryID      *int32
	TypeID         *int32
	OrganizerType  string // COMMUNITY | OFFICIAL
}

// CreateTournament validates membership limits and creates a DRAFT tournament.
func (s *Service) CreateTournament(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	req CreateTournamentRequest,
) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 3 || len(req.Name) > 150 {
		return 0, errors.New("name must be between 3 and 150 characters")
	}
	if req.OrganizerType != "COMMUNITY" && req.OrganizerType != "OFFICIAL" {
		return 0, errors.New("invalid organizer type")
	}
	if req.EndAt != nil && req.EndAt.Before(req.StartAt) {
		return 0, errors.New("end date must be after the start date")
	}

	org, err := s.repo.GetOrgContext(ctx, req.OrganizationID, actorID)
	if err != nil {
		return 0, err
	}

	active, activeOfficial, err := s.repo.CountActiveTournaments(ctx, req.OrganizationID)
	if err != nil {
		return 0, err
	}

	if err := rules.CanCreateTournament(
		org.Roles, org.Permissions, actorGlobalRoles,
		req.OrganizerType, active, activeOfficial,
	); err != nil {
		return 0, err
	}

	return s.repo.InsertTournament(ctx, &req, actorID)
}
