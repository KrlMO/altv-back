package rules

import (
	"arena/internal/domain/identity"
	"errors"
)

func ValidateTeamRole(role string) error {
	switch role {
	case identity.RoleTeamOwner, identity.RoleTeamManager, identity.RoleTeamStaff:
		return nil
	default:
		return errors.New("invalid team role specified")
	}
}
