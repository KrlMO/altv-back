package rules

import (
	"arena/internal/domain/identity"
	trules "arena/internal/domain/teams/rules"
	"errors"
)

func CanAddMembersToTeamGame(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}

	if !trules.IsTeamOwnerOrManagerOrStaff(ctx.Actor.ActorTeamRoles) && !IsTeamCaptain(ctx.ActorTeamGameRoles) {
		return errors.New("user does not have permission to add members to this team game")
	}

	return nil
}
