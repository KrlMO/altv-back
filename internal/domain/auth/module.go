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
	mailer Mailer,
	jwtSecret string,
	frontendURL string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *Module {

	service := NewService(
		usersService,
		jwtSecret,
		frontendURL,
		mailer,
		repo,
		accessDuration,
		refreshDuration,
	)

	return &Module{service: service}
}
