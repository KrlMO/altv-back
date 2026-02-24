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

	auth := u.Auth[0]

	if auth.PasswordHash == "" {
		return LoginResponse{}, "", rules.ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, auth.PasswordHash) {
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
			Auth:         mapAuth(u.Auth),
			ProfilePhoto: u.ProfilePhoto,
		},
	}, rawRefreshToken, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	rawRefreshToken string,
	userIP string,
	deviceInfo string,
) (LoginResponse, string, error) {

	tokenHash := hashRefreshToken(rawRefreshToken)

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return LoginResponse{}, "", err
	}
	defer tx.Rollback(ctx)

	storedToken, err := s.repo.GetByHashForUpdate(ctx, tx, tokenHash)
	if err != nil {
		return LoginResponse{}, "", err
	}
	if storedToken == nil {
		return LoginResponse{}, "", rules.ErrInvalidRefreshToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return LoginResponse{}, "", rules.ErrRefreshTokenExpired
	}

	if storedToken.Revoked {
		_ = s.repo.RevokeAllByUserTx(ctx, tx, storedToken.UserID)
		return LoginResponse{}, "", rules.ErrRefreshTokenReuseDetected
	}

	err = s.repo.RevokeTokenTx(ctx, tx, storedToken.ID)
	if err != nil {
		return LoginResponse{}, "", err
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

	err = s.repo.CreateRefreshTokenTx(ctx, tx, newRefresh)
	if err != nil {
		return LoginResponse{}, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return LoginResponse{}, "", err
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

	response := LoginResponse{
		AccessToken: newAccessToken,
		User: UserResponse{
			ID:           u.ID,
			Nickname:     u.NickName,
			ProfilePhoto: u.ProfilePhoto,
			Auth:         mapAuth(u.Auth),
		},
	}

	return response, newRawRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	tokenHash := hashRefreshToken(token)

	return s.repo.RevokeByHash(ctx, tokenHash)
}
