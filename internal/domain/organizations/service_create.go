package organizations

import (
	"context"
	"errors"
	"strings"

	"arena/internal/domain/identity"
)

// ErrOrganizationLimitReached is returned when a free user tries to create a
// second organization. Paid plans will lift this limit later.
var ErrOrganizationLimitReached = errors.New(
	"ya tienes una organización; crear más organizaciones estará disponible próximamente")

// CreateOrganizationService handles organization creation.
type CreateOrganizationService struct {
	repo *CreateRepository
}

func NewCreateOrganizationService(repo *CreateRepository) *CreateOrganizationService {
	return &CreateOrganizationService{repo}
}

// CreateOrganization lets any non-banned authenticated user create ONE
// organization (free plan). The creator becomes OWNER and the org starts on
// the free organizer plan.
func (s *CreateOrganizationService) CreateOrganization(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	name string,
	description *string,
) (int64, error) {
	if identity.IsBanned(actorGlobalRoles) {
		return 0, errors.New("user is banned")
	}

	// Free plan: one organization per user. Staff bypasses.
	if !identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		owned, err := s.repo.CountOwnedOrganizations(ctx, actorID)
		if err != nil {
			return 0, err
		}
		if owned >= 1 {
			return 0, ErrOrganizationLimitReached
		}
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return 0, errors.New("name must be between 2 and 100 characters")
	}
	if description != nil {
		trimmed := strings.TrimSpace(*description)
		description = &trimmed
		if len(trimmed) > 1000 {
			return 0, errors.New("description too long")
		}
	}
	return s.repo.CreateOrganization(ctx, name, description, actorID)
}
