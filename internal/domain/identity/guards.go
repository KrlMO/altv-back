package identity

import "errors"

func CanInteractWithTeams(roles []string) error {
	if IsBanned(roles) {
		return errors.New("user is banned from team interactions")
	}
	return nil
}

func CanInteractWithTeamGames(roles []string) error {
	if IsBanned(roles) {
		return errors.New("user is banned from team game interactions")
	}
	return nil
}

func CanInteractWithOrganizations(roles []string) error {
	if IsBanned(roles) {
		return ErrBannedFromOrganizations
	}
	return nil
}
