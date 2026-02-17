package rules

import "errors"

var (
	ErrInvalidCredentials           = errors.New("invalid credentials")
	ErrUserNotFound                 = errors.New("user not found")
	ErrInvalidRefreshToken          = errors.New("invalid refresh token")
	ErrRefreshTokenExpired          = errors.New("refresh token expired")
	ErrRefreshTokenNotFound         = errors.New("refresh token not found")
	ErrFailedToGenerateAccessToken  = errors.New("failed to generate access token")
	ErrFailedToGenerateRefreshToken = errors.New("failed to generate refresh token")
	ErrFailedToStoreRefreshToken    = errors.New("failed to store refresh token")
)
