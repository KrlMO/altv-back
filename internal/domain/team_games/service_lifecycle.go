package teamgames

import (
	"arena/internal/domain/identity"
	"arena/internal/domain/team_games/rules"
	trules "arena/internal/domain/teams/rules"
	"arena/pkg/storage"
	"context"
	"errors"
	"log"
	"mime/multipart"
	"strconv"
)

func (s *Service) CreateTeamGameWithMembers(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamId, gameSlug string,
	members []AddTeamGameMemberRequestItem,
	jerseyFiles map[int64]*multipart.FileHeader,
) (int64, error) {
	if err := identity.CanInteractWithTeamGames(currentUserGlobalRoles); err != nil {
		return 0, err
	}
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return 0, errors.New("invalid team ID")
	}

	gameID, err := s.repo.GetGameIdBySlug(ctx, gameSlug)
	if err != nil {
		return 0, errors.New("game does not exist: " + err.Error())
	}

	teamRoles, err := s.repo.GetUserTeamRoles(ctx, currentUserId, parsedTeamID)
	if err != nil {
		return 0, errors.New("error fetching user team roles: " + err.Error())
	}
	if !identity.IsStaffOrSuperAdmin(currentUserGlobalRoles) && !trules.IsTeamOwnerOrManager(teamRoles) {
		return 0, errors.New("user does not have permission to create a team game")
	}

	resolved := make([]ReqTeamGameMember, 0, len(members))
	jerseyPaths := make(map[int64]string, len(jerseyFiles))
	savedFilenames := make([]string, 0, len(jerseyFiles))
	for _, m := range members {
		roleID, err := s.repo.GetTeamGameRoleIDByName(ctx, m.Role)
		if err != nil {
			return 0, errors.New("unknown team game role: " + m.Role)
		}
		resolved = append(resolved, ReqTeamGameMember{UserID: m.UserID, TeamGameRoleID: roleID})

		file, ok := jerseyFiles[m.UserID]
		if !ok {
			continue
		}
		ext, err := validateJerseyImage(file)
		if err != nil {
			return 0, err
		}
		filename := storage.RandomFilename(ext)
		f, err := file.Open()
		if err != nil {
			return 0, err
		}
		if err := s.storage.Save(ctx, storage.FolderTeamGameJersey, filename, f, file.Size, file.Header.Get("Content-Type")); err != nil {
			f.Close()
			return 0, err
		}
		f.Close()
		savedFilenames = append(savedFilenames, filename)
		jerseyPaths[m.UserID] = filename
	}

	teamGameID, err := s.repo.CreateTeamGameWithMembers(ctx, parsedTeamID, currentUserId, gameID, resolved, jerseyPaths)
	if err != nil {
		for _, name := range savedFilenames {
			if err := s.storage.Delete(ctx, storage.FolderTeamGameJersey, name); err != nil {
				log.Printf("failed to clean up jersey %s: %v", name, err)
			}
		}
		return 0, err
	}
	return teamGameID, nil
}

func (s *Service) LeaveTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanLeaveTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.LeaveTeamGame(ctx, currentUserId, parsedTeamGameID)
}

func (s *Service) DeactivateTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanDeactivateTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.DeactivateTeamGame(ctx, parsedTeamGameID, currentUserId)
}

func (s *Service) ActivateTeamGame(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
) error {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return err
	}
	if err := rules.CanActivateTeamGame(tgc); err != nil {
		return err
	}
	return s.repo.ActivateTeamGame(ctx, parsedTeamGameID, currentUserId)
}
