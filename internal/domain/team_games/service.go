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
				Country: dto.CountryReponse{
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
			Country: dto.CountryReponse{
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
	dto AddTeamGameMemberRequest,
) error {
	if err := identity.CanInteractWithTeamGames(currentUserGlobalRoles); err != nil {
		return err
	}
	if len(dto.Members) == 0 {
		return rules.ErrEmptyMembers
	}
	parsedTeamGameID, err := strconv.ParseInt(teamGameId, 10, 64)
	if err != nil {
		return rules.ErrBadIdTeamGame
	}
	teamRoles, teamGameRoles, err := s.repo.GetUserTeamAndTeamGameRoles(ctx, currentUserId, parsedTeamGameID)
	if err != nil {
		return err
	}
	tgc := rules.TeamGameActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: currentUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		ActorTeamGameRoles: teamGameRoles,
		TeamGameID:         parsedTeamGameID,
	}
	if err := rules.CanAddMembersToTeamGame(tgc); err != nil {
		return err
	}
	err = s.repo.AddMembersToTeamGame(ctx, currentUserId, parsedTeamGameID, dto.Members, dto.GameId)
	if err != nil {
		return err
	}
	return nil
}
