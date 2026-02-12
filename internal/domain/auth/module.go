package auth

import (
	"arena/internal/domain/users"
	"time"
)

type Module struct {
	service *Service
}

func NewModule(
	usersService *users.Service,
	repo *Repository,
	jwtSecret string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *Module {

	service := NewService(
		usersService,
		jwtSecret,
		repo,
		accessDuration,
		refreshDuration,
	)

	return &Module{service: service}
}
