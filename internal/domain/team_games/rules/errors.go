package rules

import "errors"

var (
	ErrEmptyRoaster                = errors.New("the roaster does not exist or does not have members")
	ErrBadIdTeamGame               = errors.New("team game id is incorrect or does not exist")
	ErrEmptyMembers                = errors.New("members field is required and cannot be empty")
	ErrUserDoesNotBelongToTeamGame = errors.New("user does not belong to the team game")
	ErrCannotLeaveTeamGame         = errors.New("user cannot leave the team game")
	ErrCannotFindGame              = errors.New("cannot find the specified game or doesnt exist")
	ErrUserCannotUpdateMemberRole  = errors.New("user does not have permission to update member roles in this team game")
)
