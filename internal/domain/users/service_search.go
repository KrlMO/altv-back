package users

import "context"

func (s *Service) SearchUsers(ctx context.Context, query string) ([]UserSearchResult, error) {
	return s.repo.SearchUsers(ctx, query)
}
