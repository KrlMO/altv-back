package auth

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int64          `json:"id"`
	Nickname     string         `json:"nickname"`
	Auth         []AuthResponse `json:"auth"`
	ProfilePhoto string         `json:"profile_photo"`
}

type AuthResponse struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
	ID          int64  `json:"id"`
	IsVerified  bool   `json:"is_verified"`
}
