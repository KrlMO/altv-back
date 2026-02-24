package rules

import "errors"

func ValidateTeamRole(role string) error {
	switch role {
	case "TEAM_OWNER", "TEAM_MANAGER", "TEAM_STAFF":
		return nil
	default:
		return errors.New("invalid team role specified")
	}
}
