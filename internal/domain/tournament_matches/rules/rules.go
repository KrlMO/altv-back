package rules

import (
	"errors"
	"fmt"

	"arena/internal/domain/identity"
)

var (
	ErrBadIdMatch         = errors.New("the match does not exist or the id is incorrect")
	ErrStageNotInProgress = errors.New("matches can only be generated while the stage is in progress")
	ErrMatchesExist       = errors.New("this stage already has matches")
	ErrNoParticipants     = errors.New("the stage has no accepted registrations to pair")
	ErrNoPermission       = errors.New("user does not have permission to manage this tournament's matches")
	ErrUnknownSeeding     = errors.New("unknown seeding strategy")
	ErrBadManualOrder     = errors.New("manual seeding requires every accepted registration exactly once")
	ErrMatchNotPlayable   = errors.New("the match is not playable in its current state")
	ErrMatchHasNoTeams    = errors.New("both teams must be set to report a result")
	ErrBadWinner          = errors.New("the winner must be one of the two teams")
	ErrBadScore           = errors.New("the score is inconsistent with best_of")
	ErrNoNextRound        = errors.New("there are no pending matches of the current round or the stage is finished")
	ErrMatchNotEditable   = errors.New("the match teams can no longer be edited")
	ErrTeamDuplicated     = errors.New("a team cannot appear twice in the match")
	ErrTeamBusyInRound    = errors.New("the team already plays in this round")
	ErrNoSlotsToUpdate    = errors.New("no team slots provided")
	ErrBadSlot            = errors.New("invalid slot; must be a or b")
)

// CanManageMatches decides whether the actor may generate rounds and set
// results: organization OWNER/ADMIN, tournament
// TOURNAMENT_OWNER/TOURNAMENT_ADMIN, or global staff.
func CanManageMatches(actorOrgRoles, actorTournamentRoles, actorGlobalRoles []string) error {
	if identity.IsStaffOrSuperAdmin(actorGlobalRoles) {
		return nil
	}
	if identity.HasRole(actorOrgRoles, identity.RoleOrgOwner) ||
		identity.HasRole(actorOrgRoles, identity.RoleOrgAdmin) {
		return nil
	}
	if identity.HasRole(actorTournamentRoles, "TOURNAMENT_OWNER") ||
		identity.HasRole(actorTournamentRoles, "TOURNAMENT_ADMIN") {
		return nil
	}
	return ErrNoPermission
}

// CanReportResult extends match management to tournament MATCH_ADMIN and
// REFEREE roles.
func CanReportResult(actorOrgRoles, actorTournamentRoles, actorGlobalRoles []string) error {
	if err := CanManageMatches(actorOrgRoles, actorTournamentRoles, actorGlobalRoles); err == nil {
		return nil
	}
	if identity.HasRole(actorTournamentRoles, "MATCH_ADMIN") ||
		identity.HasRole(actorTournamentRoles, "REFEREE") {
		return nil
	}
	return ErrNoPermission
}

// CanEditTeams validates slot updates for a match.
//   - COMPLETED/WALKOVER matches are frozen forever.
//   - PENDING matches: any slot can be set.
//   - IN_PROGRESS matches: only empty slots can be filled (e.g. completing
//     a bracket TBD while the round plays).
func CanEditTeams(status string, teamA, teamB int64, newA, newB int64) error {
	if status == "COMPLETED" || status == "WALKOVER" {
		return ErrMatchNotEditable
	}
	if newA == 0 && newB == 0 {
		return ErrNoSlotsToUpdate
	}
	if newA != 0 && newA == newB {
		return ErrTeamDuplicated
	}
	if newA != 0 && teamB != 0 && newA == teamB {
		return ErrTeamDuplicated
	}
	if newB != 0 && teamA != 0 && newB == teamA {
		return ErrTeamDuplicated
	}
	if status == "IN_PROGRESS" {
		// Only filling empty slots is allowed while the match is live.
		if newA != 0 && teamA != 0 && newA != teamA {
			return ErrMatchNotEditable
		}
		if newB != 0 && teamB != 0 && newB != teamB {
			return ErrMatchNotEditable
		}
	}
	return nil
}

// ValidateResult checks a reported result against the match state.
func ValidateResult(status string, teamA, teamB, winner int64, scoreA, scoreB, bestOf int32) error {
	if status != "IN_PROGRESS" && status != "PENDING" && status != "WALKOVER" {
		return ErrMatchNotPlayable
	}
	if teamA == 0 || teamB == 0 {
		return ErrMatchHasNoTeams
	}
	if winner != teamA && winner != teamB {
		return ErrBadWinner
	}
	// A walkover skips the best_of score check (defaults 0-0).
	if status == "WALKOVER" || bestOf <= 0 {
		return nil
	}
	half := bestOf/2 + 1 // wins needed in a best_of series
	winnerScore := scoreB
	loserScore := scoreA
	if winner == teamA {
		winnerScore, loserScore = scoreA, scoreB
	}
	if winnerScore < half || winnerScore > bestOf || loserScore >= winnerScore || loserScore < 0 {
		return fmt.Errorf("%w: best_of=%d requires a winner score between %d and %d", ErrBadScore, bestOf, half, bestOf)
	}
	return nil
}
