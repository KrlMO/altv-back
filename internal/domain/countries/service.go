package countries

import "context"

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAllCountries(ctx context.Context) ([]CountryResponse, error) {
	countries, err := s.repo.GetAllCountries(ctx)
	if err != nil {
		return nil, err
	}

	var response []CountryResponse

	for _, c := range countries {
		var flag *string
		if c.FlagPath.Valid {
			flag = &c.FlagPath.String
		}

		response = append(response, CountryResponse{
			ID:       c.ID,
			Name:     c.Name,
			FlagPath: flag,
			IsoCode:  c.IsoCode,
		})
	}

	return response, nil
}
