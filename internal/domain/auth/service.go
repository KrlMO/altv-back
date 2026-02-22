package auth

import (
	"arena/internal/domain/auth/rules"
	"arena/internal/domain/users"
	"arena/pkg/utils"
	"context"
	"time"
)

type Service struct {
	usersService         *users.Service
	jwtSecret            string
	repo                 *Repository
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewService(
	usersService *users.Service,
	jwtSecret string,
	repo *Repository,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *Service {
	return &Service{
		usersService:         usersService,
		jwtSecret:            jwtSecret,
		repo:                 repo,
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

func (s *Service) LocalLogin(
	ctx context.Context,
	email, password, userIP, deviceInfo string,
) (LoginResponse, string, error) {

	u, roles, err := s.usersService.GetUser(ctx, email, "local")
	if err != nil {
		return LoginResponse{}, "", rules.ErrUserNotFound
	}

	if !utils.CheckPasswordHash(password, u.Auth.PasswordHash) {
		return LoginResponse{}, "", rules.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateJWT(
		u.ID,
		roles,
		s.jwtSecret,
		s.accessTokenDuration,
	)
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToGenerateAccessToken
	}

	rawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToGenerateRefreshToken
	}

	hash := hashRefreshToken(rawRefreshToken)

	refresh := &RefreshToken{
		UserID:     u.ID,
		TokenHash:  hash,
		DeviceInfo: &deviceInfo,
		IPAddress:  &userIP,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.refreshTokenDuration),
		Revoked:    false,
	}

	err = s.repo.CreateRefreshToken(ctx, refresh)
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToStoreRefreshToken
	}

	return LoginResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:           u.ID,
			Nickname:     u.NickName,
			ProviderUID:  u.Auth.ProviderUID,
			ProfilePhoto: u.ProfilePhoto,
		},
	}, rawRefreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string, userIP string, deviceInfo string) (LoginResponse, string, error) {

	tokenHash := hashRefreshToken(rawRefreshToken)

	storedToken, err := s.repo.GetByTokenHash(ctx, tokenHash)
	if err != nil || storedToken == nil {
		return LoginResponse{}, "", rules.ErrInvalidRefreshToken
	}

	if storedToken.Revoked {
		_ = s.repo.RevokeAllByUser(ctx, storedToken.UserID)

		return LoginResponse{}, "", rules.ErrRefreshTokenReuseDetected
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return LoginResponse{}, "", rules.ErrRefreshTokenExpired
	}

	err = s.repo.RevokeToken(ctx, storedToken.ID)
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToStoreRefreshToken
	}

	newRawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToGenerateRefreshToken
	}

	newHash := hashRefreshToken(newRawRefreshToken)

	newRefresh := &RefreshToken{
		UserID:     storedToken.UserID,
		TokenHash:  newHash,
		DeviceInfo: utils.StringPtrOrNil(deviceInfo),
		IPAddress:  utils.StringPtrOrNil(userIP),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.refreshTokenDuration),
		Revoked:    false,
	}

	err = s.repo.CreateRefreshToken(ctx, newRefresh)
	if err != nil {
		return LoginResponse{}, "", rules.ErrFailedToStoreRefreshToken
	}

	u, roles, err := s.usersService.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return LoginResponse{}, "", err
	}

	newAccessToken, err := utils.GenerateJWT(
		u.ID,
		roles,
		s.jwtSecret,
		s.accessTokenDuration,
	)

	if err != nil {
		return LoginResponse{}, "", err
	}

	reponse := LoginResponse{
		AccessToken: newAccessToken,
		User: UserResponse{
			ID:           u.ID,
			Nickname:     u.NickName,
			ProviderUID:  u.Auth.ProviderUID,
			ProfilePhoto: u.ProfilePhoto,
		},
	}

	return reponse, newRawRefreshToken, nil
}
