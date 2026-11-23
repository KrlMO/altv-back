package users

import (
	"arena/internal/shared"
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u User, provider string) error
	GetUserByProviderUID(ctx context.Context, providerUID string, provider string) (*User, error)
	GetGlobalRolesByUserID(ctx context.Context, userID int64) ([]string, error)
	GetUserProfileByID(ctx context.Context, userID int64) (*UserProfile, error)
	GetUserByID(ctx context.Context, userId int64) (*User, error)
	GetUserProviders(ctx context.Context, userID int64) ([]ProviderAuth, error)
	UpdateUser(ctx context.Context, u User) error
	PatchUser(ctx context.Context, u User) error
	DeactivateUser(ctx context.Context, userID int64) error
	ActivateUser(ctx context.Context, userID int64) error
	SoftDeleteUser(ctx context.Context, userID int64) error
	SearchUsers(ctx context.Context, query string) ([]UserSearchResult, error)
}

type Service struct {
	repo       UserRepository
	sharedRepo *shared.SharedRepository
}

func NewService(r UserRepository, sharedRepo *shared.SharedRepository) *Service {
	return &Service{repo: r, sharedRepo: sharedRepo}
}
