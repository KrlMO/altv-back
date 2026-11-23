package teamgames

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/team_games/rules"
	grules "arena/pkg/global_rules"
	"arena/pkg/storage"
	"context"
	"strconv"
)

type Service struct {
	repo    *Repository
	storage storage.Storage
}

func NewService(repo *Repository, mediaStorage storage.Storage) *Service {
	return &Service{repo: repo, storage: mediaStorage}
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
