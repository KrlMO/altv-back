package rules

import "errors"

var (
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
	ErrInvalidNickName  = errors.New("invalid nickname")
	ErrInvalidBirthDate = errors.New("invalid birth date")
	ErrInvalidCountry   = errors.New("invalid country")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrEmailNotVerified = errors.New("email not verified")
)
