package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	"arena/internal/shared"
	"arena/pkg/dto"
	grules "arena/pkg/global_rules"
	"context"
	"errors"
	"strconv"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, t Team) error
	GetTeamByID(ctx context.Context, teamID string) (*TeamReadModel, error)
	GetTeams(ctx context.Context, filter TeamFilter) ([]TeamReadModel, int64, error)
	UpdateTeam(ctx context.Context, t Team) error
	AddMemberToTeam(ctx context.Context, userID, teamID, assignedBy int64, roleCode string) error
	UpdateUserTeamRole(ctx context.Context, userId, teamId, assignedBy int64, roleName string) error
	GetUserTeamRole(ctx context.Context, userId, teamId int64) (string, error)
	DeleteUserFromTeam(ctx context.Context, userId, teamId, assignedBy int64) error
	GetGameBySlug(ctx context.Context, gameSlug string) (int32, error)
	CreateTeamGame(ctx context.Context, teamId, createdBy int64, gameId int32, nameTeamGame, logoPath, bannerPath string) error
	GetMyTeams(ctx context.Context, userId int64, globalRoles []string) ([]MyTeamListItem, error)
}

type Service struct {
	repo       TeamRepository
	sharedRepo *shared.SharedRepository
}

func NewService(repo TeamRepository, sharedRepo *shared.SharedRepository) *Service {
	return &Service{repo: repo, sharedRepo: sharedRepo}
}

func (s *Service) RegisterTeam(ctx context.Context, userID int64, globalRoles []string, dto RegisterTeamRequest) error {
	if identity.IsOrc(globalRoles) {
		return errors.New("User cannot create a team, user is banned from creating teams")
	}

	err := ValidateTeamCreateOrUpdate(dto.Name, dto.Tag, dto.Description, dto.LogoPath, dto.BannerPath)
	if err != nil {
		return err
	}

	exists, err := s.sharedRepo.CountryExists(ctx, dto.CountryId)
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
		Active:      true,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	return s.repo.CreateTeam(ctx, t)
}

func (s *Service) GetTeamByID(ctx context.Context, teamID string) (*TeamResponse, error) {
	t, err := s.repo.GetTeamByID(ctx, teamID)
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
		Country: dto.CountryResponse{
			ID:       t.Country.ID,
			Name:     t.Country.Name,
			FlagPath: flag,
		},
		Active: t.Team.Active,
	}
	return team, nil
}

func (s *Service) GetTeams(ctx context.Context, filter TeamFilter) (dto.PaginatedResponse[TeamResponse], error) {
	teams, total, err := s.repo.GetTeams(ctx, filter)
	if err != nil {
		return dto.PaginatedResponse[TeamResponse]{}, err
	}

	result := make([]TeamResponse, 0, len(teams))
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
			Country: dto.CountryResponse{
				ID:       t.Country.ID,
				Name:     t.Country.Name,
				FlagPath: flag,
			},
			Active: t.Team.Active,
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

func (s *Service) PatchUpdateTeam(ctx context.Context, userId int64, globalRoles []string, teamId string, dto PatchUpdateTeamRequest) error {
	if identity.IsOrc(globalRoles) {
		return errors.New("User cannot update a team, user is banned from updating teams")
	}
	parsedID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}

	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, userId, parsedID)
	t := Team{ID: parsedID}
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}
	if !rules.IsTeamOwnerOrManager(teamRoles) && !identity.IsIluvatarOrAinur(globalRoles) {
		return errors.New("user does not have permission to update this team or is not an administrator")
	}

	if err := ValidateTeamPatch(dto.Name, dto.Tag, dto.Description, dto.LogoPath, dto.BannerPath); err != nil {
		return err
	}

	if dto.Name != nil {
		t.Name = *dto.Name
	}
	if dto.Tag != nil {
		t.Tag = *dto.Tag
	}
	if dto.Description != nil {
		t.Description = *dto.Description
	}
	if dto.LogoPath != nil {
		t.LogoPath = *dto.LogoPath
	}
	if dto.BannerPath != nil {
		t.BannerPath = *dto.BannerPath
	}
	if dto.CountryId != nil {
		exists, err := s.sharedRepo.CountryExists(ctx, *dto.CountryId)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("invalid country")
		}
		t.CountryId = *dto.CountryId
	}

	t.UpdatedBy = userId

	return s.repo.UpdateTeam(ctx, t)
}

func (s *Service) AddMemberToTeam(
	ctx context.Context,
	userID int64,
	teamID string,
	globalRole []string,
	dto AddMemberRequest) error {
	if err := rules.ValidateTeamRole(dto.Role); err != nil {
		return err
	}
	if err := identity.CanInteractWithTeams(globalRole); err != nil {
		return err
	}
	parsedTeamID, err := strconv.ParseInt(teamID, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, userID, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          userID,
			ActorGlobalRoles: globalRole,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: dto.UserID,
		TargetRole:   dto.Role,
	}

	if err := rules.CanAddMember(tac); err != nil {
		return err
	}

	return s.repo.AddMemberToTeam(ctx, dto.UserID, parsedTeamID, userID, dto.Role)
}

func (s *Service) UpdateUserTeamRole(
	ctx context.Context,
	currentUserId int64,
	curretUserGlobalRole []string,
	teamId, userId string,
	dto UpdateTeamMemberRoleRequest) error {
	if err := rules.ValidateTeamRole(*dto.Role); err != nil {
		return err
	}
	if err := identity.CanInteractWithTeams(curretUserGlobalRole); err != nil {
		return err
	}

	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	parsedUserID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return errors.New("invalid user ID")
	}

	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: curretUserGlobalRole,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: parsedUserID,
		TargetRole:   *dto.Role,
	}

	if err := rules.CanUpdateMemberRole(tac); err != nil {
		return err
	}

	return s.repo.UpdateUserTeamRole(ctx, parsedUserID, parsedTeamID, currentUserId, *dto.Role)
}

func (s *Service) DeleteUserFromTeam(
	ctx context.Context,
	currentUserId int64,
	curretUserGlobalRoles []string,
	teamId, userId string,
) error {
	if err := identity.CanInteractWithTeams(curretUserGlobalRoles); err != nil {
		return err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	parsedUserID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return errors.New("invalid user ID")
	}

	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}
	actualRole, err := s.repo.GetUserTeamRole(ctx, parsedUserID, parsedTeamID)
	if errors.Is(err, rules.ErrUserNotMember) {
		return errors.New("user is not a member of this team")
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: curretUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: parsedUserID,
		TargetRole:   actualRole,
	}

	if err := rules.CanDeleteMemberFromTeam(tac); err != nil {
		return errors.New("Error deleting user from team: " + err.Error())
	}
	return s.repo.DeleteUserFromTeam(ctx, tac.TargetUserID, parsedTeamID, tac.Actor.ActorID)
}

func (s *Service) AddTeamGame(
	ctx context.Context,
	teamId string,
	currentUserId int64,
	currentUserGlobalRoles []string,
	dto CreateTeamGameRequest,
) error {
	if err := identity.CanInteractWithTeams(currentUserGlobalRoles); err != nil {
		return err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	gameId, err := s.repo.GetGameBySlug(ctx, dto.GameSlug)
	if err != nil {
		return errors.New("Error, game does not exist: " + err.Error())
	}

	tgc := rules.TeamGameContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: currentUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		GameId: gameId,
	}

	if err := rules.CanCreateTeamGame(tgc); err != nil {
		return errors.New("Error creating Team Game: " + err.Error())
	}
	return s.repo.CreateTeamGame(ctx, parsedTeamID, tgc.Actor.ActorID, tgc.GameId, dto.Name, *dto.LogoPath, *dto.BannerPath)
}

func (s *Service) GetMyTeams(ctx context.Context, userId int64, globalRoles []string) ([]MyTeamListItemDto, error) {
	if err := identity.CanInteractWithTeams(globalRoles); err != nil {
		return nil, err
	}
	teams, err := s.repo.GetMyTeams(ctx, userId, globalRoles)
	if err != nil {
		return nil, err
	}

	if teams == nil {
		return []MyTeamListItemDto{}, nil
	}

	var result []MyTeamListItemDto
	for _, t := range teams {
		result = append(result, MyTeamListItemDto{
			ID:   t.ID,
			Name: t.Name,
			Tag:  t.Tag,
			Logo: t.LogoPath,
		})
	}
	return result, nil
}
