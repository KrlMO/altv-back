package auth

import "arena/internal/domain/users"

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int64                         `json:"id"`
	Nickname     string                        `json:"nickname"`
	Auth         []users.ProvidersAuthResponse `json:"auth"`
	ProfilePhoto string                        `json:"profile_photo"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailVerificationResponse struct {
	Status         string `json:"status"`
	ResendCooldown int64  `json:"resend_cooldown,omitempty"`
}
