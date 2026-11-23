package games

import "context"

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListGames(ctx context.Context, name string) ([]GameListItemResponse, error) {
	games, err := s.repo.ListGames(ctx, name)
	if err != nil {
		return nil, err
	}

	response := make([]GameListItemResponse, 0, len(games))
	for _, g := range games {
		var colorHex *string
		if g.ColorHex.Valid {
			colorHex = &g.ColorHex.String
		}
		response = append(response, GameListItemResponse{
			ID:       g.ID,
			Name:     g.Name,
			Slug:     g.Slug,
			ColorHex: colorHex,
		})
	}

	return response, nil
}
