package teams

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/teams/rules"
	grules "arena/pkg/global_rules"
	"context"
	"errors"
	"strconv"
)

func (s *Service) AddMemberToTeam(
	ctx context.Context,
	userID int64,
	teamID string,
	globalRole []string,
	dto AddMemberRequest) error {
	if err := rules.ValidateTeamRole(dto.Role); err != nil {
		return err
	}
	if err := identity.CanInteractWithTeams(globalRole); err != nil {
		return err
	}
	parsedTeamID, err := strconv.ParseInt(teamID, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, userID, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          userID,
			ActorGlobalRoles: globalRole,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: dto.UserID,
		TargetRole:   dto.Role,
	}

	if err := rules.CanAddMember(tac); err != nil {
		return err
	}

	return s.repo.AddMemberToTeam(ctx, dto.UserID, parsedTeamID, userID, dto.Role)
}

func (s *Service) UpdateUserTeamRole(
	ctx context.Context,
	currentUserId int64,
	curretUserGlobalRole []string,
	teamId, userId string,
	dto UpdateTeamMemberRoleRequest) error {
	if err := rules.ValidateTeamRole(*dto.Role); err != nil {
		return err
	}
	if err := identity.CanInteractWithTeams(curretUserGlobalRole); err != nil {
		return err
	}

	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	parsedUserID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return errors.New("invalid user ID")
	}

	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: curretUserGlobalRole,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: parsedUserID,
		TargetRole:   *dto.Role,
	}

	if err := rules.CanUpdateMemberRole(tac); err != nil {
		return err
	}

	return s.repo.UpdateUserTeamRole(ctx, parsedUserID, parsedTeamID, currentUserId, *dto.Role)
}

func (s *Service) DeleteUserFromTeam(
	ctx context.Context,
	currentUserId int64,
	curretUserGlobalRoles []string,
	teamId, userId string,
) error {
	if err := identity.CanInteractWithTeams(curretUserGlobalRoles); err != nil {
		return err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return errors.New("invalid team ID")
	}
	parsedUserID, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return errors.New("invalid user ID")
	}

	teamRoles, err := s.sharedRepo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return errors.New("Error fetching user team roles: " + err.Error())
	}
	actualRole, err := s.repo.GetUserTeamRole(ctx, parsedUserID, parsedTeamID)
	if errors.Is(err, rules.ErrUserNotMember) {
		return errors.New("user is not a member of this team")
	}

	tac := rules.TeamActionContext{
		Actor: grules.ActorContext{
			ActorID:          currentUserId,
			ActorGlobalRoles: curretUserGlobalRoles,
			ActorTeamRoles:   teamRoles,
		},
		TargetUserID: parsedUserID,
		TargetRole:   actualRole,
	}

	if err := rules.CanDeleteMemberFromTeam(tac); err != nil {
		return errors.New("Error deleting user from team: " + err.Error())
	}
	return s.repo.DeleteUserFromTeam(ctx, tac.TargetUserID, parsedTeamID, tac.Actor.ActorID)
}
