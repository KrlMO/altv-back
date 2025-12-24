package teams

import (
	"arena/internal/domain/users"
	"arena/pkg/middleware"
	"errors"

	"arena/pkg/dto"

	"github.com/gin-gonic/gin"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterTeam(c *gin.Context, userID int64, dto RegisterTeamRequest) error {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !CanCreateTeam(roles) {
		return errors.New("User cannot create a team")
	}

	err := ValidateTeamCreateOrUpdate(dto.Name, dto.Tag, dto.Description, dto.LogoPath, dto.BannerPath)
	if err != nil {
		return err
	}

	exists, err := s.repo.CountryExists(c, dto.CountryId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Country does not exist")
	}
	t := Team{
		Name:        dto.Name,
		Tag:         dto.Tag,
		Description: dto.Description,
		LogoPath:    dto.LogoPath,
		BannerPath:  dto.BannerPath,
		CountryId:   dto.CountryId,
		IsVirtual:   dto.IsVirtual,
		Active:      true,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	return s.repo.CreateTeam(c, t)
}

func (s *Service) GetTeamByID(c *gin.Context, teamID string) (*TeamResponse, error) {
	t, err := s.repo.GetTeamByID(c, teamID)
	if err != nil {
		return nil, err
	}
	flag := ""
	if t.Country.FlagPath.Valid {
		flag = t.Country.FlagPath.String
	}
	team := &TeamResponse{
		ID:          t.Team.ID,
		Name:        t.Team.Name,
		Tag:         t.Team.Tag,
		Description: t.Team.Description,
		LogoPath:    t.Team.LogoPath,
		BannerPath:  t.Team.BannerPath,
		Country: users.CountryReponse{
			ID:       t.Country.ID,
			Name:     t.Country.Name,
			FlagPath: flag,
		},
		IsVirtual: t.Team.IsVirtual,
		Active:    t.Team.Active,
	}
	return team, nil
}

func (s *Service) GetTeams(c *gin.Context, filter TeamFilter) (dto.PaginatedResponse[TeamResponse], error) {
	teams, total, err := s.repo.GetTeams(c, filter)
	if err != nil {
		return dto.PaginatedResponse[TeamResponse]{}, err
	}

	var result []TeamResponse
	for _, t := range teams {
		flag := ""
		if t.Country.FlagPath.Valid {
			flag = t.Country.FlagPath.String
		}
		team := TeamResponse{
			ID:          t.Team.ID,
			Name:        t.Team.Name,
			Tag:         t.Team.Tag,
			Description: t.Team.Description,
			LogoPath:    t.Team.LogoPath,
			BannerPath:  t.Team.BannerPath,
			Country: users.CountryReponse{
				ID:       t.Country.ID,
				Name:     t.Country.Name,
				FlagPath: flag,
			},
			IsVirtual: t.Team.IsVirtual,
			Active:    t.Team.Active,
		}
		result = append(result, team)
	}
	response := dto.PaginatedResponse[TeamResponse]{
		Data: result,
		Meta: dto.PaginationMeta{
			Total:  total,
			Limit:  filter.Limit,
			Offset: filter.Offset,
		},
	}

	return response, nil
}
