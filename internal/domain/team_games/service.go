package teamgames

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/team_games/rules"
	"arena/pkg/dto"
	grules "arena/pkg/global_rules"
	"context"
	"strconv"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
) (rules.TeamGameActionContext, string, int64, error) {
	parsedTeamGameID, err := strconv.ParseInt(teamGameId, 10, 64)
	if err != nil {
		return rules.TeamGameActionContext{}, "", 0, rules.ErrBadIdTeamGame
	}
	if err := identity.CanInteractWithTeamGames(currentUserGlobalRoles); err != nil {
		return rules.TeamGameActionContext{}, "", 0, err
	}
	teamGameName, teamRoles, teamGameRoles, err := s.repo.GetUserTeamAndTeamGameRoles(ctx, currentUserId, parsedTeamGameID)
	if err != nil {
		return rules.TeamGameActionContext{}, "", 0, err
	}
	tgc := rules.TeamGameActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: currentUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		ActorTeamGameRoles: teamGameRoles,
		TeamGameID:         parsedTeamGameID,
		BelongsToTeam:      teamGameName != "" && len(teamRoles) > 0 && len(teamGameRoles) > 0,
	}
	return tgc, teamGameName, parsedTeamGameID, nil
}

func (s *Service) AddMembersToTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto AddTeamGameMembersRequest,
) (string, error) {
	tgc, teamGameName, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "err", err
	}
	if err := rules.CanAddMembersToTeamGame(tgc); err != nil {
		return "", err
	}
	err = s.repo.AddMembersToTeamGame(ctx, currentUserId, parsedTeamGameID, dto.Members, dto.GameId)
	if err != nil {
		return "", err
	}
	return teamGameName, nil
}

func (s *Service) LeaveTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) (string, error) {

	tgc, teamGameName, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "err", err
	}
	if err := rules.CanLeaveTeamGame(tgc); err != nil {
		return "err", err
	}
	err = s.repo.LeaveTeamGame(ctx, currentUserId, parsedTeamGameID)
	if err != nil {
		return "err", err
	}
	return teamGameName, nil
}

func (s *Service) RemoveMembersFromTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto RemoveTeamGameMembersRequest,
) (string, error) {

	tgc, teamGameName, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "err", err
	}

	if err := rules.CanRemoveMembersFromTeamGame(tgc); err != nil {
		return "err", err
	}
	err = s.repo.RemoveMembersFromTeamGame(ctx, currentUserId, parsedTeamGameID, dto.Members)
	if err != nil {
		return "err", err
	}
	return teamGameName, nil
}

func (s *Service) RemoveTeamGameRoleFromMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto RemoveTeamGameRoleFromUserRequest,
) (string, error) {
	tgc, teamGameName, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "err", err
	}

	if err := rules.CanRemoveTeamGameRoleFromMember(tgc); err != nil {
		return "err", err
	}

	err = s.repo.RemoveTeamGameRoleFromMember(ctx, currentUserId, parsedTeamGameID, dto.Users)
	if err != nil {
		return "err", err
	}
	return teamGameName, nil
}

func (s *Service) AddTeamGameRoleToMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto AddTeamGameRoleToMemberRequest,
) (string, error) {
	tgc, teamGameName, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "err", err
	}

	if err := rules.CanAddTeamGameRoleToMember(tgc); err != nil {
		return "err", err
	}

	err = s.repo.AddTeamGameRoleToMember(ctx, currentUserId, parsedTeamGameID, dto.Users)
	if err != nil {
		return "err", err
	}
	return teamGameName, nil
}
