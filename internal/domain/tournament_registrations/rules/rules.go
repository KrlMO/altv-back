package rules

import (
	"errors"

	"arena/internal/domain/identity"
)

var (
	ErrBadIdRegistration          = errors.New("the registration does not exist or the id is incorrect")
	ErrTournamentNotOpen          = errors.New("the tournament does not accept registrations")
	ErrGameMismatch               = errors.New("the roster's game does not match the tournament's game")
	ErrTeamGameInactive           = errors.New("the roster is inactive or does not exist")
	ErrNoPermissionToRegister     = errors.New("user does not have permission to register this roster")
	ErrAlreadyRegistered          = errors.New("the roster is already registered to this tournament")
	ErrNoPermissionToReview       = errors.New("user does not have permission to review registrations")
	ErrRegistrationNotPending     = errors.New("only pending registrations can be reviewed")
	ErrNoPermissionToCancel       = errors.New("user does not have permission to cancel this registration")
	ErrNoPermissionToManageRoster = errors.New("user does not have permission to manage this roster")
	ErrNotTeamGameMember          = errors.New("user is not a member of the roster")
	ErrUserInAnotherRoster        = errors.New("user is already in a roster of this tournament")
	ErrRosterMemberNotFound       = errors.New("user is not part of this registration's roster")
)

// CanRegisterTeamGame decides whether the actor may register a team_game
// (roster) to a tournament. Permission comes from the team: only the
// team owner/manager (or global staff) can register one of its rosters.
func CanRegisterTeamGame(actorTeamRoles []string, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if identity.HasRole(actorTeamRoles, identity.RoleTeamOwner) ||
		identity.HasRole(actorTeamRoles, identity.RoleTeamManager) {
		return nil
	}
	return ErrNoPermissionToRegister
}

// CanReviewRegistration decides whether the actor may accept/reject
// registrations of a tournament. Allowed: organization OWNER/ADMIN of the
// tournament's organization, tournament TOURNAMENT_OWNER/TOURNAMENT_ADMIN,
// or global staff.
func CanReviewRegistration(actorOrgRoles []string, actorTournamentRoles []string, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if identity.HasRole(actorOrgRoles, identity.RoleOrgOwner) ||
		identity.HasRole(actorOrgRoles, identity.RoleOrgAdmin) {
		return nil
	}
	if identity.HasRole(actorTournamentRoles, "TOURNAMENT_OWNER") ||
		identity.HasRole(actorTournamentRoles, "TOURNAMENT_ADMIN") {
		return nil
	}
	return ErrNoPermissionToReview
}

// CanCancelRegistration decides whether the actor may cancel a registration.
// Allowed: whoever registered it, the team owner/manager of the roster's
// team, or global staff.
func CanCancelRegistration(actorID, registeredBy int64, actorTeamRoles []string, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if actorID == registeredBy {
		return nil
	}
	if identity.HasRole(actorTeamRoles, identity.RoleTeamOwner) ||
		identity.HasRole(actorTeamRoles, identity.RoleTeamManager) {
		return nil
	}
	return ErrNoPermissionToCancel
}

// CanManageRegistrationRoster decides whether the actor may add/remove
// players from a registration's roster. Same permission as registering:
// team owner/manager, plus the roster's TEAM_CAPTAIN.
func CanManageRegistrationRoster(actorTeamRoles []string, actorTeamGameRoles []string, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if identity.HasRole(actorTeamRoles, identity.RoleTeamOwner) ||
		identity.HasRole(actorTeamRoles, identity.RoleTeamManager) {
		return nil
	}
	if identity.HasRole(actorTeamGameRoles, identity.RoleTeamCaptain) {
		return nil
	}
	return ErrNoPermissionToManageRoster
}
