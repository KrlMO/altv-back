package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	"arena/pkg/dto"
	"arena/pkg/storage"
	"context"
	"errors"
	"strconv"
)

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
