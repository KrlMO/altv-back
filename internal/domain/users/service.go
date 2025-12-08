package users

import (
	"arena/pkg/utils"
	"context"
	"errors"
	"time"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(r *Repository, jwtSecret string) *Service {
	return &Service{r, jwtSecret}
}

func (s *Service) Register(u User) error {
	hash, err := utils.HashPassword(u.PasswordHash)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.repo.CreateUser(context.Background(), u)
}

func (s *Service) Login(email, password string) (string, error) {
	u, err := s.repo.GetUserByEmail(context.Background(), email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, u.PasswordHash) {
		return "", errors.New("invalid credentials")
	}
	token, err := utils.GenerateJWT(u.ID, s.jwtSecret, 30*time.Minute)
	return token, err
}
