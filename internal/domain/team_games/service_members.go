package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/dto"
	"arena/pkg/storage"
	"context"
	"errors"
	"strconv"
)

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
