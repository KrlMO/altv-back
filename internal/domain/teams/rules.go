package teams

func CanCreateTeam(roles []string) bool {
	for _, r := range roles {
		if r == "ORC" {
			return false
		}
	}
	return true
}
