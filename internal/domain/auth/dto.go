package auth

type LoginRequest struct {
	EmailOrNickname string `json:"email_or_nickname" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int32  `json:"id"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	ProfilePhoto string `json:"profile_photo"`
}
