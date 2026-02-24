package rules

import "errors"

var (
	ErrUserNotMember     = errors.New("user is not a member of this team")
	ErrCannotDeleteOwner = errors.New("cannot remove team owner")
)
