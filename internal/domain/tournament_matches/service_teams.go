package matches

import (
	"context"

	"arena/internal/domain/tournament_matches/rules"
)

// SetMatchTeams assigns teams to a match's slots manually. Slots omitted
// (zero) are left untouched. In IN_PROGRESS matches only empty slots can be
// filled; completed matches are frozen.
func (s *Service) SetMatchTeams(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	matchID int64,
	req SetMatchTeamsRequest,
) error {
	m, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return err
	}
	if err := s.permissions(ctx, actorID, actorGlobalRoles, m.TournamentID, false); err != nil {
		return err
	}

	var teamA, teamB int64
	if m.TeamA != nil {
		teamA = *m.TeamA
	}
	if m.TeamB != nil {
		teamB = *m.TeamB
	}
	if err := rules.CanEditTeams(m.Status, teamA, teamB, req.TeamARegistrationID, req.TeamBRegistrationID); err != nil {
		return err
	}

	if req.TeamARegistrationID != 0 {
		if err := s.checkSlotFree(ctx, m, "a", req.TeamARegistrationID); err != nil {
			return err
		}
	}
	if req.TeamBRegistrationID != 0 {
		if err := s.checkSlotFree(ctx, m, "b", req.TeamBRegistrationID); err != nil {
			return err
		}
	}

	if m.Status == string(StatusInProgress) {
		// Live match: only fill empty slots.
		fillA := req.TeamARegistrationID != 0 && m.TeamA == nil
		fillB := req.TeamBRegistrationID != 0 && m.TeamB == nil
		if fillA {
			if err := s.repo.SetMatchTeam(ctx, matchID, "a", req.TeamARegistrationID); err != nil {
				return err
			}
		}
		if fillB {
			return s.repo.SetMatchTeam(ctx, matchID, "b", req.TeamBRegistrationID)
		}
		if fillA || fillB {
			return nil
		}
		// Reaching here means an invalid IN_PROGRESS change slipped through.
		return rules.ErrMatchNotEditable
	}

	return s.repo.SetMatchTeams(ctx, matchID, req.TeamARegistrationID, req.TeamBRegistrationID)
}

// checkSlotFree rejects teams already playing in another match of the round.
func (s *Service) checkSlotFree(ctx context.Context, m *Match, slot string, registrationID int64) error {
	busy, err := s.repo.RegistrationBusyInRound(ctx, m.StageID, int64(m.RoundNumber), registrationID, m.ID)
	if err != nil {
		return err
	}
	if busy {
		return rules.ErrTeamBusyInRound
	}
	return nil
}
