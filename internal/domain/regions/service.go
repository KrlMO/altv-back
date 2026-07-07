package regions

import "context"

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAllRegions(ctx context.Context) ([]RegionResponse, error) {
	regions, err := s.repo.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]RegionResponse, 0, len(regions))
	for _, rg := range regions {
		resp := RegionResponse{
			ID:      rg.ID,
			Name:    rg.Name,
			IsoCode: rg.IsoCode,
		}
		if rg.FlagPath.Valid {
			resp.FlagPath = &rg.FlagPath.String
		}
		response = append(response, resp)
	}

	return response, nil
}
