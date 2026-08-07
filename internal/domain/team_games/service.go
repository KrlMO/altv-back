package teamgames

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/team_games/rules"
	trules "arena/internal/domain/teams/rules"
	"arena/pkg/dto"
	grules "arena/pkg/global_rules"
	"arena/pkg/storage"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Service struct {
	repo    *Repository
	storage storage.Storage
}

func NewService(repo *Repository, mediaStorage storage.Storage) *Service {
	return &Service{repo: repo, storage: mediaStorage}
}

func (s *Service) GetTeamGameMembers(
	ctx context.Context,
	teamGameId string,
) (PublicTeamGameMembersResponse, error) {
	parsedTeamID, err := strconv.ParseInt(teamGameId, 10, 64)
	if err != nil {
		return PublicTeamGameMembersResponse{}, rules.ErrBadIdTeamGame
	}
	teamGameMembers, err := s.repo.GetTeamGamePublicMembers(ctx, parsedTeamID)
	if err != nil {
		return PublicTeamGameMembersResponse{}, err
	}
	if len(teamGameMembers) == 0 {
		return PublicTeamGameMembersResponse{}, rules.ErrEmptyRoaster
	}
	teamGameMembersResponse := PublicTeamGameMembersResponse{}

	for i := range teamGameMembers {
		if teamGameMembers[i].TeamGameRole == "COACH" {
			teamGameMembersResponse.Coach = &TeamGameMemberResponse{
				UserID:    teamGameMembers[i].UserID,
				Nickname:  teamGameMembers[i].Nickname,
				FirstName: teamGameMembers[i].FirstName,
				LastName:  teamGameMembers[i].LastName,
				Country: dto.CountryResponse{
					ID:       teamGameMembers[i].Country.ID,
					Name:     teamGameMembers[i].Country.Name,
					FlagPath: teamGameMembers[i].Country.FlagPath.String,
					IsoCode:  teamGameMembers[i].Country.IsoCode,
				},
				Role: teamGameMembers[i].TeamGameRole,
			}
			continue
		}
		teamGameMembersResponse.Roaster = append(teamGameMembersResponse.Roaster, TeamGameMemberResponse{
			UserID:    teamGameMembers[i].UserID,
			Nickname:  teamGameMembers[i].Nickname,
			FirstName: teamGameMembers[i].FirstName,
			LastName:  teamGameMembers[i].LastName,
			Country: dto.CountryResponse{
				ID:       teamGameMembers[i].Country.ID,
				Name:     teamGameMembers[i].Country.Name,
				FlagPath: teamGameMembers[i].Country.FlagPath.String,
				IsoCode:  teamGameMembers[i].Country.IsoCode,
			},
			Role: teamGameMembers[i].TeamGameRole,
		})
	}
	return teamGameMembersResponse, nil
}

func (s *Service) GetTeamGameActionContext(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) (rules.TeamGameActionContext, int64, error) {
	parsedTeamGameID, err := strconv.ParseInt(teamGameId, 10, 64)
	if err != nil {
		return rules.TeamGameActionContext{}, 0, rules.ErrBadIdTeamGame
	}
	if err := identity.CanInteractWithTeamGames(currentUserGlobalRoles); err != nil {
		return rules.TeamGameActionContext{}, 0, err
	}
	teamRoles, teamGameRoles, err := s.repo.GetUserTeamAndTeamGameRoles(ctx, currentUserId, parsedTeamGameID)
	if err != nil {
		return rules.TeamGameActionContext{}, 0, err
	}
	tgc := rules.TeamGameActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: currentUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		ActorTeamGameRoles: teamGameRoles,
		TeamGameID:         parsedTeamGameID,
		BelongsToTeam:      len(teamRoles) > 0 && len(teamGameRoles) > 0,
	}
	return tgc, parsedTeamGameID, nil
}

func (s *Service) AddMembersToTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto AddTeamGameMembersRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanAddMembersToTeamGame(tgc); err != nil {
		return err
	}
	members := make([]ReqTeamGameMember, 0, len(dto.Members))
	for _, m := range dto.Members {
		roleID, err := s.repo.GetTeamGameRoleIDByName(ctx, m.Role)
		if err != nil {
			return errors.New("unknown team game role: " + m.Role)
		}
		members = append(members, ReqTeamGameMember{UserID: m.UserID, TeamGameRoleID: roleID})
	}
	err = s.repo.AddMembersToTeamGame(ctx, currentUserId, parsedTeamGameID, members, dto.GameId)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateTeamGameWithMembers(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamId, gameSlug string,
	members []AddTeamGameMemberRequestItem,
	jerseyFiles map[int64]*multipart.FileHeader,
) (int64, error) {
	if err := identity.CanInteractWithTeamGames(currentUserGlobalRoles); err != nil {
		return 0, err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return 0, errors.New("invalid team ID")
	}

	gameID, err := s.repo.GetGameIdBySlug(ctx, gameSlug)
	if err != nil {
		return 0, errors.New("game does not exist: " + err.Error())
	}

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return 0, errors.New("error fetching user team roles: " + err.Error())
	}
	if !identity.IsStaffOrSuperAdmin(currentUserGlobalRoles) && !trules.IsTeamOwnerOrManager(teamRoles) {
		return 0, errors.New("user does not have permission to create a team game")
	}

	resolved := make([]ReqTeamGameMember, 0, len(members))
	jerseyPaths := make(map[int64]string, len(jerseyFiles))
	savedFilenames := make([]string, 0, len(jerseyFiles))
	for _, m := range members {
		roleID, err := s.repo.GetTeamGameRoleIDByName(ctx, m.Role)
		if err != nil {
			return 0, errors.New("unknown team game role: " + m.Role)
		}
		resolved = append(resolved, ReqTeamGameMember{UserID: m.UserID, TeamGameRoleID: roleID})

		file, ok := jerseyFiles[m.UserID]
		if !ok {
			continue
		}
		ext, err := validateJerseyImage(file)
		if err != nil {
			return 0, err
		}
		filename := storage.RandomFilename(ext)
		f, err := file.Open()
		if err != nil {
			return 0, err
		}
		if err := s.storage.Save(ctx, storage.FolderTeamGameJersey, filename, f, file.Size, file.Header.Get("Content-Type")); err != nil {
			f.Close()
			return 0, err
		}
		f.Close()
		savedFilenames = append(savedFilenames, filename)
		jerseyPaths[m.UserID] = filename
	}

	teamGameID, err := s.repo.CreateTeamGameWithMembers(ctx, parsedTeamID, currentUserId, gameID, resolved, jerseyPaths)
	if err != nil {
		for _, name := range savedFilenames {
			s.storage.Delete(ctx, storage.FolderTeamGameJersey, name)
		}
		return 0, err
	}
	return teamGameID, nil
}

func (s *Service) GetTeamGameManageMembers(
	ctx context.Context,
	teamGameId string,
) ([]TeamGameManageMemberResponse, error) {
	parsedTeamGameID, err := strconv.ParseInt(teamGameId, 10, 64)
	if err != nil {
		return nil, rules.ErrBadIdTeamGame
	}
	members, err := s.repo.GetTeamGameManageMembers(ctx, parsedTeamGameID)
	if err != nil {
		return nil, err
	}
	result := make([]TeamGameManageMemberResponse, 0, len(members))
	for _, m := range members {
		resp := TeamGameManageMemberResponse{
			UserID:      m.UserID,
			Username:    m.Username,
			FirstName:   m.FirstName,
			LastName:    m.LastName,
			CountryCode: m.Country.IsoCode,
			TeamRole:    m.TeamRole,
			Country: dto.CountryResponse{
				ID:       m.Country.ID,
				Name:     m.Country.Name,
				FlagPath: m.Country.FlagPath.String,
				IsoCode:  m.Country.IsoCode,
			},
		}
		if m.ProfilePhoto.Valid {
			resp.Avatar = m.ProfilePhoto.String
		}
		if m.BirthDate.Valid {
			resp.BirthDate = &m.BirthDate.Time
		}
		if m.JerseyImage.Valid {
			resp.JerseyPhoto = s.storage.URL(storage.FolderTeamGameJersey, m.JerseyImage.String)
		}
		result = append(result, resp)
	}
	return result, nil
}

var allowedImageTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

const maxImageSize = 5 * 1024 * 1024

func validateJerseyImage(fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "", errors.New("file is required")
	}
	if fh.Size > maxImageSize {
		return "", errors.New("the jersey image cannot exceed 5MB")
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
		return "", errors.New("the jersey image must be PNG, JPEG or WEBP")
	}
	return ext, nil
}

func (s *Service) UploadMemberJersey(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	memberUserId int64,
	file *multipart.FileHeader,
) (string, error) {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "", err
	}
	if err := rules.CanAddMembersToTeamGame(tgc); err != nil {
		return "", err
	}

	ext, err := validateJerseyImage(file)
	if err != nil {
		return "", err
	}
	filename := storage.RandomFilename(ext)

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := s.storage.Save(ctx, storage.FolderTeamGameJersey, filename, f, file.Size, file.Header.Get("Content-Type")); err != nil {
		return "", err
	}

	if err := s.repo.UpsertTeamGamePlayerImage(ctx, memberUserId, parsedTeamGameID, currentUserId, filename); err != nil {
		return "", err
	}

	return s.storage.URL(storage.FolderTeamGameJersey, filename), nil
}

func (s *Service) LeaveTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanLeaveTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.LeaveTeamGame(ctx, currentUserId, parsedTeamGameID)
}

func (s *Service) RemoveMembersFromTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto RemoveTeamGameMembersRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}

	if err := rules.CanRemoveMembersFromTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.RemoveMembersFromTeamGame(ctx, currentUserId, parsedTeamGameID, dto.Members)
}

func (s *Service) toTeamGameListItemResponse(item TeamGameListItem) TeamGameListItemResponse {
	resp := TeamGameListItemResponse{
		ID:       item.ID,
		GameID:   item.GameID,
		TeamName: item.TeamName,
		GameName: item.GameName,
		Active:   item.Active,
	}
	if item.LogoKey.Valid {
		u := s.storage.URL(storage.FolderTeamsLogo, item.LogoKey.String)
		resp.Logo = &u
	}
	if item.RegionName.Valid {
		resp.RegionName = &item.RegionName.String
	}
	if item.RegionISO.Valid {
		resp.RegionISO = &item.RegionISO.String
	}
	if item.CountryName.Valid {
		resp.CountryName = &item.CountryName.String
	}
	if item.CreatedAt.Valid {
		resp.CreatedAt = &item.CreatedAt.Time
	}
	return resp
}

func (s *Service) ListTeamGames(
	ctx context.Context,
	filter TeamGameListFilter,
	page int,
) (dto.PaginatedResponse[TeamGameListItemResponse], error) {
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	items, total, err := s.repo.FilterTeamGames(ctx, filter, page, limit)
	if err != nil {
		return dto.PaginatedResponse[TeamGameListItemResponse]{}, err
	}

	response := make([]TeamGameListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, s.toTeamGameListItemResponse(item))
	}

	return dto.PaginatedResponse[TeamGameListItemResponse]{
		Data: response,
		Meta: dto.PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func (s *Service) ListTeamGamesByTeam(ctx context.Context, teamId string) ([]TeamGameListItemResponse, error) {
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return nil, rules.ErrBadIdTeamGame
	}

	items, err := s.repo.GetTeamGamesByTeam(ctx, parsedTeamID)
	if err != nil {
		return nil, err
	}

	response := make([]TeamGameListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, s.toTeamGameListItemResponse(item))
	}

	return response, nil
}

func (s *Service) DeactivateTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanDeactivateTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.DeactivateTeamGame(ctx, parsedTeamGameID, currentUserId)
}

func (s *Service) ActivateTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanActivateTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.ActivateTeamGame(ctx, parsedTeamGameID, currentUserId)
}

func (s *Service) AddTeamGameRoleToMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto AddTeamGameRoleToMemberRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}

	if err := rules.CanAddTeamGameRoleToMember(tgc); err != nil {
		return err
	}

	users, err := s.resolveMemberRoles(ctx, dto.Users)
	if err != nil {
		return err
	}

	return s.repo.AddTeamGameRoleToMember(ctx, currentUserId, parsedTeamGameID, users)
}

func (s *Service) RemoveTeamGameRoleFromMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto RemoveTeamGameRoleFromUserRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}

	if err := rules.CanRemoveTeamGameRoleFromMember(tgc); err != nil {
		return err
	}

	users, err := s.resolveMemberRoles(ctx, dto.Users)
	if err != nil {
		return err
	}

	return s.repo.RemoveTeamGameRoleFromMember(ctx, currentUserId, parsedTeamGameID, users)
}

func (s *Service) resolveMemberRoles(ctx context.Context, users []TeamGameMemberDto) ([]TeamGameMemberDto, error) {
	resolved := make([]TeamGameMemberDto, 0, len(users))
	for _, u := range users {
		roles := make([]TeamGameRoleDto, 0, len(u.Roles))
		for _, role := range u.Roles {
			id, err := s.repo.GetTeamGameRoleIDByName(ctx, role.Name)
			if err != nil {
				return nil, errors.New("unknown team game role: " + role.Name)
			}
			roles = append(roles, TeamGameRoleDto{ID: id, Name: role.Name})
		}
		resolved = append(resolved, TeamGameMemberDto{UserID: u.UserID, Roles: roles})
	}
	return resolved, nil
}
