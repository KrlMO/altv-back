package auth

import (
	"arena/internal/domain/users"
	"arena/pkg/utils"
	"context"
	"errors"
	"time"
)

type Service struct {
	usersService *users.Service
	jwtSecret    string
}

func NewService(usersService *users.Service, jwtSecret string) *Service {
	return &Service{usersService, jwtSecret}
}

func (s *Service) LocalLogin(ctx context.Context, email, password string) (LoginResponse, error) {
	u, roles, err := s.usersService.GetUser(ctx, email, "local")
	if err != nil {
		return LoginResponse{}, errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, u.Auth.PasswordHash) {
		return LoginResponse{}, errors.New("invalid credentials")
	}

	token, err := utils.GenerateJWT(
		u.ID, roles, s.jwtSecret, 30*time.Minute,
	)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: token,
		User: UserResponse{
			ID:           u.ID,
			Nickname:     u.NickName,
			ProviderUID:  u.Auth.ProviderUID,
			ProfilePhoto: u.ProfilePhoto,
		},
	}, nil
}
