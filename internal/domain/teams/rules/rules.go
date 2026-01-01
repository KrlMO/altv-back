package rules

func IsTeamOwnerOrManager(roles []string) bool {
	for _, r := range roles {
		if r == "TEAM_MANAGER" || r == "TEAM_OWNER" {
			return true
		}
	}
	return false
}

func IsTeamManager(roles []string) bool {
	for _, r := range roles {
		if r == "TEAM_MANAGER" {
			return true
		}
	}
	return false
}

func IsTeamOwner(roles []string) bool {
	for _, r := range roles {
		if r == "TEAM_OWNER" {
			return true
		}
	}
	return false
}
