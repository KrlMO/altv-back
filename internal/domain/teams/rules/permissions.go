package rules

import (
	"arena/internal/domain/identity"
	"errors"
)

func CanAddMember(ctx TeamActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.ActorGlobalRoles) {
		return nil
	}

	if !IsTeamOwnerOrManager(ctx.ActorTeamRoles) {
		return errors.New("user does not have permission to add members to this team")
	}

	if IsTeamOwner(ctx.ActorTeamRoles) && ctx.TargetRole == "TEAM_OWNER" {
		return errors.New("team owner cannot add another owner")
	}

	if IsTeamManager(ctx.ActorTeamRoles) &&
		(ctx.TargetRole == "TEAM_OWNER" || ctx.TargetRole == "TEAM_MANAGER") {
		return errors.New("team manager cannot assign owner or manager roles")
	}

	return nil
}

func CanUpdateMemberRole(ctx TeamActionContext) error {
	if identity.IsIluvatarOrAinur(ctx.ActorGlobalRoles) {
		return nil
	}

	if !IsTeamOwnerOrManager(ctx.ActorTeamRoles) {
		return errors.New("user does not have permission to update roles")
	}

	if IsTeamOwner(ctx.ActorTeamRoles) && ctx.TargetRole == "TEAM_OWNER" {
		return errors.New("team owner cannot update another owner")
	}

	if IsTeamManager(ctx.ActorTeamRoles) &&
		(ctx.TargetRole == "TEAM_OWNER" || ctx.TargetRole == "TEAM_MANAGER") {
		return errors.New("team manager cannot assign owner or manager roles")
	}

	return nil
}
