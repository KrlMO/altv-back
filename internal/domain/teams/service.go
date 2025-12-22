package teams

import (
	"arena/pkg/middleware"
	"errors"

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
		CreatedBy:   int32(userID),
		UpdatedBy:   int32(userID),
	}
	return s.repo.CreateTeam(c, t)
}
