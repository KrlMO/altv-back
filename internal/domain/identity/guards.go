package identity

import "errors"

func CanInteractWithTeams(roles []string) error {
	if IsOrc(roles) {
		return errors.New("user is banned from team interactions")
	}
	return nil
}
