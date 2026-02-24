package rules

func IsTeamCaptain(roles []string) bool {
	for _, r := range roles {
		if r == "TEAM_CAPTAIN" {
			return true
		}
	}
	return false
}

func IsStaff(roles []string) bool {
	for _, r := range roles {
		if r == "STAFF" {
			return true
		}
	}
	return false
}

func IsCoach(roles []string) bool {
	for _, r := range roles {
		if r == "COACH" {
			return true
		}
	}
	return false
}

func IsAnalyst(roles []string) bool {
	for _, r := range roles {
		if r == "ANALYST" {
			return true
		}
	}
	return false
}

func IsPlayer(roles []string) bool {
	for _, r := range roles {
		if r == "PLAYER" {
			return true
		}
	}
	return false
}

func IsTeamGameMember(belongsToTeam bool) bool {
	return belongsToTeam
}
