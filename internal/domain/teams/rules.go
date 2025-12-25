package teams

func IsOrcMember(roles []string) bool {
	for _, r := range roles {
		if r == "ORC" {
			return true
		}
	}

	return false
}

func IsTeamOwnerOrManager(roles []string) bool {
	for _, r := range roles {
		if r == "TEAM_MANAGER" || r == "TEAM_OWNER" {
			return true
		}
	}
	return false
}

func IsIluvatarOrAinur(roles []string) bool {
	for _, r := range roles {
		if r == "ILUVATAR" || r == "AINUR" {
			return true
		}
	}
	return false
}
