package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	"arena/pkg/middleware"
	"context"
	"errors"
	"regexp"
	"strconv"

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
	if identity.IsOrc(roles) {
		return errors.New("User cannot create a team, user is banned from creating teams")
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
		Country: dto.CountryReponse{
			ID:       t.Country.ID,
			Name:     t.Country.Name,
			FlagPath: flag,
		},
		IsVirtual: t.Team.IsVirtual,
		Active:    t.Team.Active,
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
			Country: dto.CountryReponse{
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

func (s *Service) PatchUpdateTeam(ctx *gin.Context, userId int64, teamId string, dto PatchUpdateTeamRequest) error {
	globalRoles := ctx.MustGet(middleware.ContextRoles).([]string)
	if identity.IsOrc(globalRoles) {
		return errors.New("User cannot update a team, user is banned from updating teams")
	}
	parsedID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, userId, parsedID)
	t := Team{ID: parsedID}
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}
	if !rules.IsTeamOwnerOrManager(teamRoles) && !identity.IsIluvatarOrAinur(globalRoles) {
		return errors.New("user does not have permission to update this team or is not an administrator")
	}

	if dto.Name != nil {
		if len(*dto.Name) < 3 || len(*dto.Name) > 50 {
			return errors.New("Name must be between 3 and 50 characters")
		}
		t.Name = *dto.Name
	}

	if dto.Tag != nil {
		if len(*dto.Tag) < 2 || len(*dto.Tag) > 15 {
			return errors.New("Tag must be between 2 and 15 characters")
		}
		t.Tag = *dto.Tag
	}

	if dto.Description != nil {
		if len(*dto.Description) > 500 {
			return errors.New("Description must be less than 500 characters")
		}
		t.Description = *dto.Description
	}

	if dto.LogoPath != nil {
		_, err := regexp.MatchString(`^https?://`, *dto.LogoPath)
		if err != nil {
			return errors.New("invalid logo_path URL")
		}
		t.LogoPath = *dto.LogoPath
	}
	if dto.BannerPath != nil {
		_, err := regexp.MatchString(`^https?://`, *dto.BannerPath)
		if err != nil {
			return errors.New("invalid banner_path URL")
		}
		t.BannerPath = *dto.BannerPath
	}
	if dto.CountryId != nil {
		exists, err := s.repo.CountryExists(ctx, *dto.CountryId)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("invalid country")
		}
		t.CountryId = *dto.CountryId
	}

	if dto.IsVirtual != nil {
		t.IsVirtual = *dto.IsVirtual
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
	teamRoles, err := s.repo.GetUserTeamRoles(ctx, userID, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: rules.ActorContext{
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

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: rules.ActorContext{
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

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}
	actualRole, err := s.repo.GetUserTeamRole(ctx, parsedUserID, parsedTeamID)
	if errors.Is(err, rules.ErrUserNotMember) {
		return errors.New("user is not a member of this team")
	}

	tac := rules.TeamActionContext{
		Actor: rules.ActorContext{
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
	teamRoles, err := s.repo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	gameId, err := s.repo.GetGameBySlug(ctx, dto.GameSlug)
	if err != nil {
		return errors.New("Error, game does not exist: " + err.Error())
	}

	tgc := rules.TeamGameContext{
		Actor: rules.ActorContext{
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
