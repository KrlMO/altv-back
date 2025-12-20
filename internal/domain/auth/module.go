package auth

import "arena/internal/domain/users"

type Module struct {
	service *Service
}

func NewModule(usersService *users.Service, jwtSecret string) *Module {
	service := NewService(usersService, jwtSecret)
	return &Module{service}
}
