package users

import "arena/pkg/model"

type UserProfile struct {
	User    User
	Country model.Country
}
