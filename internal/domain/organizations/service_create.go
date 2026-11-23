package organizations

import (
	"context"
	"errors"
	"strings"

	"arena/internal/domain/identity"
)

// CreateOrganizationService handles organization creation.
type CreateOrganizationService struct {
	repo *CreateRepository
}

func NewCreateOrganizationService(repo *CreateRepository) *CreateOrganizationService {
	return &CreateOrganizationService{repo}
}

// CreateOrganization lets any non-banned authenticated user create an
// organization. The creator becomes OWNER and the org starts on the free
// organizer plan.
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
