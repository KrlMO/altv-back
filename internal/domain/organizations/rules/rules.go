package rules

import "arena/internal/domain/identity"

const unknownRoleRank = 99

var organizationRoleRanks = map[string]int{
	identity.RoleOrgOwner:     0,
	identity.RoleOrgAdmin:     1,
	identity.RoleOrgModerator: 2,
	identity.RoleOrgStaff:     3,
	identity.RoleOrgAnalyst:   4,
	identity.RoleOrgCreator:   5,
}

// RoleRank orders organization roles from most to least privileged.
func RoleRank(role string) int {
	if rank, ok := organizationRoleRanks[role]; ok {
		return rank
	}
	return unknownRoleRank
}
