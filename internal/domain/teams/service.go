package teams

import (
	"arena/internal/shared"
	"arena/pkg/dto"
	"arena/pkg/storage"
	"context"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, t Team) (int64, error)
	GetTeamByID(ctx context.Context, teamID string) (*TeamReadModel, error)
	GetTeams(ctx context.Context, filter TeamFilter) ([]TeamReadModel, int64, error)
	UpdateTeam(ctx context.Context, t Team) error
	AddMemberToTeam(ctx context.Context, userID, teamID, assignedBy int64, roleCode string) error
	UpdateUserTeamRole(ctx context.Context, userId, teamId, assignedBy int64, roleName string) error
	GetUserTeamRole(ctx context.Context, userId, teamId int64) (string, error)
	DeleteUserFromTeam(ctx context.Context, userId, teamId, assignedBy int64) error
	GetGameBySlug(ctx context.Context, gameSlug string) (int32, error)
	CreateTeamGame(ctx context.Context, teamId, createdBy int64, gameId int32) (int64, error)
	GetMyTeams(ctx context.Context, userId int64, globalRoles []string) ([]MyTeamListItem, error)
	GetUserTeamCount(ctx context.Context, userId int64) (int, error)
}

type Service struct {
	repo       TeamRepository
	sharedRepo *shared.SharedRepository
	storage    storage.Storage
}

func NewService(repo TeamRepository, sharedRepo *shared.SharedRepository, mediaStorage storage.Storage) *Service {
	return &Service{repo: repo, sharedRepo: sharedRepo, storage: mediaStorage}
}

func (s *Service) toTeamResponse(t *TeamReadModel) *TeamResponse {
	flag := ""
	if t.Country.FlagPath.Valid {
		flag = t.Country.FlagPath.String
	}
	logo := ""
	if t.Team.LogoKey != "" {
		logo = s.storage.URL(storage.FolderTeamsLogo, t.Team.LogoKey)
	}
	banner := ""
	if t.Team.BannerKey != "" {
		banner = s.storage.URL(storage.FolderTeamsBanner, t.Team.BannerKey)
	}
	return &TeamResponse{
		ID:          t.Team.ID,
		Name:        t.Team.Name,
		Tag:         t.Team.Tag,
		Description: t.Team.Description,
		Logo:        logo,
		Banner:      banner,
		Country: dto.CountryResponse{
			ID:       t.Country.ID,
			Name:     t.Country.Name,
			FlagPath: flag,
		},
		Active: t.Team.Active,
	}
}
