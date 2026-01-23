package rules

import "errors"

var (
	ErrEmptyRoaster  = errors.New("the roaster does not exist or does not have members")
	ErrBadIdTeamGame = errors.New("Team Game id is incorrect or does not exist")
	ErrEmptyMembers  = errors.New("members field is required and cannot be empty")
)
