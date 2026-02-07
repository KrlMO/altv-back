package auth

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int64  `json:"id"`
	Nickname     string `json:"nickname"`
	ProviderUID  string `json:"provider_uid"`
	ProfilePhoto string `json:"profile_photo"`
}
