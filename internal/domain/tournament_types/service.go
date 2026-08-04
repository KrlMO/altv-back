package tournamenttypes

import "context"

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTournamentTypes(ctx context.Context) ([]TournamentTypeListItemResponse, error) {
	types, err := s.repo.ListTournamentTypes(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]TournamentTypeListItemResponse, 0, len(types))
	for _, t := range types {
		response = append(response, TournamentTypeListItemResponse{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	return response, nil
}
