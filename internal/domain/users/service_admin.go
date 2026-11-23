package users

import "context"

func (s *Service) DeactivateUser(ctx context.Context, userID int64) error {
	return s.repo.DeactivateUser(ctx, userID)
}

func (s *Service) ActivateUser(ctx context.Context, userID int64) error {
	return s.repo.ActivateUser(ctx, userID)
}

func (s *Service) SoftDeleteUser(ctx context.Context, userID int64) error {
	return s.repo.SoftDeleteUser(ctx, userID)
}
