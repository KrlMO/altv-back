package identity

import "errors"

const (
	RoleSuperAdmin  = "SUPER_ADMIN"
	RoleStaff       = "STAFF"
	RoleUser        = "USER"
	RoleBanned      = "BANNED"
	RoleTeamOwner   = "TEAM_OWNER"
	RoleTeamManager = "TEAM_MANAGER"
	RoleTeamStaff   = "TEAM_STAFF"
	RoleTeamCaptain = "TEAM_CAPTAIN"
	RolePlayer      = "PLAYER"
	RoleCoach       = "COACH"
)

// Organization roles. RoleOrgStaff shares its value with RoleStaff but belongs to a
// different scope: never compare an organization role against a global role slice.
const (
	RoleOrgOwner     = "OWNER"
	RoleOrgAdmin     = "ADMIN"
	RoleOrgModerator = "MODERATOR"
	RoleOrgStaff     = "STAFF"
	RoleOrgAnalyst   = "ANALYST"
	RoleOrgCreator   = "CREATOR"
)

var ErrBannedFromOrganizations = errors.New("user is banned from organization interactions")
