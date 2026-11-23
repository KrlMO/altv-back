package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	grules "arena/pkg/global_rules"
	"context"
	"errors"
	"strconv"
)

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
