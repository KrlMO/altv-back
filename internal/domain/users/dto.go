package users

import "time"

type RegisterUserRequest struct {
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	NickName  string    `json:"nickname" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
	CountryId int32     `json:"country_id" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	EmailOrNickname string `json:"email_or_nickname" binding:"required,email"`
	Password        string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID           int32  `json:"id"`
	Nickname     string `json:"nickname"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	CountryID    int32  `json:"country_id"`
	ProfilePhoto string `json:"profile_photo"`
}
