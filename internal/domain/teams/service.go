package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	"arena/internal/shared"
	"arena/pkg/dto"
	grules "arena/pkg/global_rules"
	"arena/pkg/storage"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
)

var allowedImageTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

const maxImageSize = 5 * 1024 * 1024

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

func validateTeamImage(fh *multipart.FileHeader, field string) (string, error) {
	if fh == nil {
		return "", nil
	}
	if fh.Size > maxImageSize {
		return "", errors.New("The " + field + " image cannot exceed 5MB")
	}
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", errors.New("The " + field + " image must be PNG, JPEG or WEBP")
	}
	return ext, nil
}

func (s *Service) saveTeamImage(ctx context.Context, fh *multipart.FileHeader, folder, field string) (string, error) {
	ext, err := validateTeamImage(fh, field)
	if err != nil {
		return "", err
	}
	filename := storage.RandomFilename(ext)

	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := s.storage.Save(ctx, folder, filename, f, fh.Size, fh.Header.Get("Content-Type")); err != nil {
		return "", err
	}
	return filename, nil
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

func (s *Service) RegisterTeam(ctx context.Context, userID int64, globalRoles []string, dto *CreateTeamRequestDto) (*TeamResponse, error) {
	if identity.IsBanned(globalRoles) {
		return nil, errors.New("user cannot create a team, user is banned from creating teams")
	}

	if !identity.IsStaffOrSuperAdmin(globalRoles) {
		teamCount, err := s.repo.GetUserTeamCount(ctx, userID)
		if err != nil {
			return nil, err
		}
		if teamCount >= 1 {
			return nil, errors.New("user already has a team, cannot create another one")
		}
	}

	description := ""
	if dto.Description != nil {
		description = *dto.Description
	}

	if err := ValidateTeamCreateOrUpdate(dto.Name, dto.Tag, description); err != nil {
		return nil, err
	}

	exists, err := s.sharedRepo.CountryExists(ctx, dto.CountryId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("country does not exist")
	}

	t := Team{
		Name:        dto.Name,
		Tag:         dto.Tag,
		Description: description,
		CountryId:   dto.CountryId,
		Active:      true,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	if dto.Logo != nil {
		filename, err := s.saveTeamImage(ctx, dto.Logo, storage.FolderTeamsLogo, "logo")
		if err != nil {
			return nil, err
		}
		t.LogoKey = filename
	}
	if dto.Banner != nil {
		filename, err := s.saveTeamImage(ctx, dto.Banner, storage.FolderTeamsBanner, "banner")
		if err != nil {
			return nil, err
		}
		t.BannerKey = filename
	}

	teamID, err := s.repo.CreateTeam(ctx, t)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.GetTeamByID(ctx, strconv.FormatInt(teamID, 10))
	if err != nil {
		return nil, err
	}
	return s.toTeamResponse(created), nil
}

func (s *Service) GetTeamByID(ctx context.Context, teamID string) (*TeamResponse, error) {
	t, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return s.toTeamResponse(t), nil
}

func (s *Service) GetTeams(ctx context.Context, filter TeamFilter) (dto.PaginatedResponse[TeamResponse], error) {
	teams, total, err := s.repo.GetTeams(ctx, filter)
	if err != nil {
		return dto.PaginatedResponse[TeamResponse]{}, err
	}

	result := make([]TeamResponse, 0, len(teams))
	for _, t := range teams {
		result = append(result, *s.toTeamResponse(&t))
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
	if identity.IsBanned(globalRoles) {
		return errors.New("user cannot update a team, user is banned from updating teams")
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
	if !rules.IsTeamOwnerOrManager(teamRoles) && !identity.IsStaffOrSuperAdmin(globalRoles) {
		return errors.New("user does not have permission to update this team or is not an administrator")
	}

	current, err := s.repo.GetTeamByID(ctx, teamId)
	if err != nil {
		return errors.New("Error fetching team: " + err.Error())
	}

	if err := ValidateTeamPatch(dto.Name, dto.Tag, dto.Description); err != nil {
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
	if dto.Logo != nil {
		filename, err := s.saveTeamImage(ctx, dto.Logo, storage.FolderTeamsLogo, "logo")
		if err != nil {
			return err
		}
		t.LogoKey = filename
		if current.Team.LogoKey != "" {
			_ = s.storage.Delete(ctx, storage.FolderTeamsLogo, current.Team.LogoKey)
		}
	}
	if dto.Banner != nil {
		filename, err := s.saveTeamImage(ctx, dto.Banner, storage.FolderTeamsBanner, "banner")
		if err != nil {
			return err
		}
		t.BannerKey = filename
		if current.Team.BannerKey != "" {
			_ = s.storage.Delete(ctx, storage.FolderTeamsBanner, current.Team.BannerKey)
		}
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
) (int64, error) {
	if err := identity.CanInteractWithTeams(currentUserGlobalRoles); err != nil {
		return 0, err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return 0, errors.New("invalid team ID")
	}
	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return 0, errors.New("Error fetching user team roles: " + err.Error())
	}

	gameId, err := s.repo.GetGameBySlug(ctx, dto.GameSlug)
	if err != nil {
		return 0, errors.New("Error, game does not exist: " + err.Error())
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
		return 0, errors.New("Error creating Team Game: " + err.Error())
	}
	return s.repo.CreateTeamGame(ctx, parsedTeamID, tgc.Actor.ActorID, tgc.GameId)
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
		var logo *string
		if t.LogoKey != nil {
			u := s.storage.URL(storage.FolderTeamsLogo, *t.LogoKey)
			logo = &u
		}
		result = append(result, MyTeamListItemDto{
			ID:         t.ID,
			Name:       t.Name,
			Tag:        t.Tag,
			Logo:       logo,
			MemberRole: t.MemberRole,
		})
	}
	return result, nil
}
