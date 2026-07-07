package rules

import "arena/internal/domain/identity"

func IsTeamOwnerOrManager(roles []string) bool {
	for _, r := range roles {
		if r == identity.RoleTeamManager || r == identity.RoleTeamOwner {
			return true
		}
	}
	return false
}

func IsTeamManager(roles []string) bool {
	for _, r := range roles {
		if r == identity.RoleTeamManager {
			return true
		}
	}
	return false
}

func IsTeamOwner(roles []string) bool {
	for _, r := range roles {
		if r == identity.RoleTeamOwner {
			return true
		}
	}
	return false
}

func IsTeamOwnerOrManagerOrStaff(roles []string) bool {
	for _, r := range roles {
		if r == identity.RoleTeamManager || r == identity.RoleTeamOwner || r == identity.RoleTeamStaff {
			return true
		}
	}
	return false
}
