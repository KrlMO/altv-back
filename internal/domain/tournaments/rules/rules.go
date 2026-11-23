package rules

import (
	"errors"

	"arena/internal/domain/identity"
)

var (
	ErrNoPermission          = errors.New("user does not have permission to create tournaments for this organization")
	ErrOrgNoCreatePermission = errors.New("the organization's current plan does not allow creating tournaments")
	ErrOfficialNotAllowed    = errors.New("the organization's current plan does not allow official tournaments")
	ErrActiveLimitReached    = errors.New("the organization already has an active tournament; upgrade the plan to create more")
	ErrOfficialLimitReached  = errors.New("the organization already has an active official tournament; upgrade the plan to create more")
	ErrOrgNotFound           = errors.New("the organization does not exist")
)

// Roles allowed to create tournaments on behalf of the organization.
func canCreateByRole(orgRoles []string) bool {
	return identity.HasRole(orgRoles, identity.RoleOrgOwner) ||
		identity.HasRole(orgRoles, identity.RoleOrgAdmin) ||
		identity.HasRole(orgRoles, identity.RoleOrgCreator)
}

// orgHas reports whether the organization's active membership grants a code.
func orgHas(permissions []string, code string) bool {
	return identity.HasRole(permissions, code)
}

// CanCreateTournament validates a tournament creation attempt.
//   - Global staff bypasses every check.
//   - Actor needs OWNER/ADMIN/CREATOR role in the organization.
//   - The organization's membership needs CREATE_TOURNAMENT.
//   - OFFICIAL tournaments need CREATE_OFFICIAL_TOURNAMENT.
//   - Without CREATE_MULTIPLE_* the organization can only have one active
//     tournament (and one active official tournament) at a time.
func CanCreateTournament(
	actorOrgRoles []string,
	orgPermissions []string,
	actorGlobalRoles []string,
	organizerType string,
	activeTournaments int,
	activeOfficialTournaments int,
) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if !canCreateByRole(actorOrgRoles) {
		return ErrNoPermission
	}
	if !orgHas(orgPermissions, "CREATE_TOURNAMENT") {
		return ErrOrgNoCreatePermission
	}
	if organizerType == "OFFICIAL" && !orgHas(orgPermissions, "CREATE_OFFICIAL_TOURNAMENT") {
		return ErrOfficialNotAllowed
	}
	if !orgHas(orgPermissions, "CREATE_MULTIPLE_TOURNAMENTS") && activeTournaments >= 1 {
		return ErrActiveLimitReached
	}
	if organizerType == "OFFICIAL" &&
		!orgHas(orgPermissions, "CREATE_MULTIPLE_OFFICIAL_TOURNAMENTS") &&
		activeOfficialTournaments >= 1 {
		return ErrOfficialLimitReached
	}
	return nil
}
