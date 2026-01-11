package rules

import "errors"

var (
	ErrEmptyRoaster  = errors.New("the roaster does not exist")
	ErrBadIdTeamGame = errors.New("Team Game id is incorrect or does not exist")
)
