package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"context"
	"errors"
)

func (s *Service) AddTeamGameRoleToMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto AddTeamGameRoleToMemberRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}

	if err := rules.CanAddTeamGameRoleToMember(tgc); err != nil {
		return err
	}

	users, err := s.resolveMemberRoles(ctx, dto.Users)
	if err != nil {
		return err
	}

	return s.repo.AddTeamGameRoleToMember(ctx, currentUserId, parsedTeamGameID, users)
}

func (s *Service) RemoveTeamGameRoleFromMember(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	dto RemoveTeamGameRoleFromUserRequest,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}

	if err := rules.CanRemoveTeamGameRoleFromMember(tgc); err != nil {
		return err
	}

	users, err := s.resolveMemberRoles(ctx, dto.Users)
	if err != nil {
		return err
	}

	return s.repo.RemoveTeamGameRoleFromMember(ctx, currentUserId, parsedTeamGameID, users)
}

func (s *Service) resolveMemberRoles(ctx context.Context, users []TeamGameMemberDto) ([]TeamGameMemberDto, error) {
	resolved := make([]TeamGameMemberDto, 0, len(users))
	for _, u := range users {
		roles := make([]TeamGameRoleDto, 0, len(u.Roles))
		for _, role := range u.Roles {
			id, err := s.repo.GetTeamGameRoleIDByName(ctx, role.Name)
			if err != nil {
				return nil, errors.New("unknown team game role: " + role.Name)
			}
			roles = append(roles, TeamGameRoleDto{ID: id, Name: role.Name})
		}
		resolved = append(resolved, TeamGameMemberDto{UserID: u.UserID, Roles: roles})
	}
	return resolved, nil
}
