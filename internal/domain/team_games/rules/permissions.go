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

func CanLeaveTeamGame(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}
	if !IsTeamGameMember(ctx.BelongsToTeam) {
		return ErrUserDoesNotBelongToTeamGame
	}
	if IsTeamCaptain(ctx.ActorTeamGameRoles) {
		return ErrCannotLeaveTeamGame
	}
	return nil
}

func CanRemoveMembersFromTeamGame(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}

	if !trules.IsTeamOwnerOrManagerOrStaff(ctx.Actor.ActorTeamRoles) && !IsTeamCaptain(ctx.ActorTeamGameRoles) {
		return errors.New("user does not have permission to remove members from this team game")
	}

	return nil
}

func CanUpdateTeamGameMemberRole(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}

	if !trules.IsTeamOwnerOrManagerOrStaff(ctx.Actor.ActorTeamRoles) {
		return errors.New("user does not have permission to update member roles in this team game")
	}

	return nil
}

func CanRemoveTeamGameRoleFromMember(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}

	if !IsTeamGameMember(ctx.BelongsToTeam) {
		return ErrUserDoesNotBelongToTeamGame
	}

	if !trules.IsTeamOwnerOrManagerOrStaff(ctx.Actor.ActorTeamRoles) {
		return errors.New("user does not have permission to remove roles from members in this team game")
	}

	return nil
}

func CanAddTeamGameRoleToMember(ctx TeamGameActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.Actor.ActorGlobalRoles) {
		return nil
	}

	if !IsTeamGameMember(ctx.BelongsToTeam) {
		return ErrUserDoesNotBelongToTeamGame
	}

	if !trules.IsTeamOwnerOrManagerOrStaff(ctx.Actor.ActorTeamRoles) {
		return errors.New("user does not have permission to add roles to members in this team game")
	}

	return nil
}
