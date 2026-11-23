package matches

import (
	"context"

	"arena/internal/domain/tournament_matches/rules"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

// permissions mirrors the stage-management permission check: org OWNER/ADMIN,
// tournament TOURNAMENT_OWNER/TOURNAMENT_ADMIN, or global staff.
func (s *Service) permissions(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	tournamentID int64,
	reporter bool,
) error {
	orgRoles, tournamentRoles, err := s.actorRoles(ctx, actorID, tournamentID)
	if err != nil {
		return err
	}
	if reporter {
		return rules.CanReportResult(orgRoles, tournamentRoles, actorGlobalRoles)
	}
	return rules.CanManageMatches(orgRoles, tournamentRoles, actorGlobalRoles)
}

// GenerateMatches seeds a stage with matches for its format.
func (s *Service) GenerateMatches(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	stageID int64,
	req GenerateMatchesRequest,
) (int, error) {
	stage, err := s.repo.GetStageInfo(ctx, stageID)
	if err != nil {
		return 0, err
	}
	if err := s.permissions(ctx, actorID, actorGlobalRoles, stage.TournamentID, false); err != nil {
		return 0, err
	}
	if stage.Status != "IN_PROGRESS" {
		return 0, rules.ErrStageNotInProgress
	}
	if n, err := s.repo.CountStageMatches(ctx, stageID); err != nil {
		return 0, err
	} else if n > 0 {
		return 0, rules.ErrMatchesExist
	}

	_, strategy, err := seedingFor(req.Seeding)
	if err != nil {
		return 0, err
	}

	participants, err := s.repo.ListAcceptedParticipants(ctx, stage.TournamentID)
	if err != nil {
		return 0, err
	}
	if len(participants) < 2 {
		return 0, rules.ErrNoParticipants
	}

	seeds, err := strategy.Order(participants, req.ManualOrder)
	if err != nil {
		return 0, err
	}

	generator, ok := generatorFor(stage.FormatCode)
	if !ok {
		return 0, rules.ErrUnknownSeeding // format without generator registered
	}

	matches, err := generator.Generate(*stage, seeds)
	if err != nil {
		return 0, err
	}
	if err := s.repo.InsertMatches(ctx, stageID, stage.TournamentID, actorID, matches); err != nil {
		return 0, err
	}
	return len(matches), nil
}

// ListStageMatches returns the matches of a stage (public view).
func (s *Service) ListStageMatches(ctx context.Context, stageID int64) ([]MatchResponse, error) {
	stage, err := s.repo.GetStageInfo(ctx, stageID)
	if err != nil {
		return nil, err
	}
	ms, err := s.repo.ListStageMatches(ctx, stageID)
	if err != nil {
		return nil, err
	}
	return s.toResponses(ctx, ms, stage.TournamentID)
}

// SetMatchResult reports a match result and propagates the winner through
// single elimination brackets.
func (s *Service) SetMatchResult(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	matchID int64,
	req SetMatchResultRequest,
) error {
	m, err := s.repo.GetMatchByID(ctx, matchID)
	if err != nil {
		return err
	}
	if err := s.permissions(ctx, actorID, actorGlobalRoles, m.TournamentID, true); err != nil {
		return err
	}

	var teamA, teamB int64
	if m.TeamA != nil {
		teamA = *m.TeamA
	}
	if m.TeamB != nil {
		teamB = *m.TeamB
	}
	var bestOf int32
	if m.BestOf != nil {
		bestOf = *m.BestOf
	}

	// Winner by score unless the score itself says it (defaults to higher score).
	winner := teamA
	if req.ScoreB > req.ScoreA {
		winner = teamB
	}
	if err := rules.ValidateResult(m.Status, teamA, teamB, winner, req.ScoreA, req.ScoreB, bestOf); err != nil {
		return err
	}

	if err := s.repo.SetMatchResult(ctx, matchID, winner, req.ScoreA, req.ScoreB, actorID); err != nil {
		return err
	}

	// Propagate through the bracket: single elimination moves the winner;
	// double elimination also drops the loser to the lower bracket.
	stage, err := s.repo.GetStageInfo(ctx, m.StageID)
	if err != nil {
		return err
	}
	switch stage.FormatCode {
	case "SINGLE_ELIMINATION":
		return s.propagate(ctx, m, winner, nil)
	case "DOUBLE_ELIMINATION":
		var loser int64
		if winner == teamA {
			loser = teamB
		} else {
			loser = teamA
		}
		return s.propagate(ctx, m, winner, &loser)
	}
	return nil
}

// propagate writes the winner into the next bracket slot; in double
// elimination it also writes the loser into the lower bracket slot.
func (s *Service) propagate(ctx context.Context, m *Match, winner int64, loser *int64) error {
	size, err := s.deSize(ctx, m.StageID)
	if err != nil {
		return err
	}

	round, matchNumber, slot, ok := deNextWinner(size, m.RoundNumber, m.MatchNumber)
	if ok {
		next, err := s.repo.GetMatchByRoundAndNumber(ctx, m.StageID, int64(round), int64(matchNumber))
		if err == nil {
			if err := s.repo.SetBracketSlot(ctx, next.ID, slot, winner); err != nil {
				return err
			}
		}
	}

	if loser != nil {
		lRound, lMatch, lSlot, ok := deNextLoser(size, m.RoundNumber, m.MatchNumber)
		if ok {
			lb, err := s.repo.GetMatchByRoundAndNumber(ctx, m.StageID, int64(lRound), int64(lMatch))
			if err == nil {
				if err := s.repo.SetBracketSlot(ctx, lb.ID, lSlot, *loser); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// deSize derives the padded bracket size from the stage's round-1 match count
// (a DE bracket of size S has S/2 first-round matches).
func (s *Service) deSize(ctx context.Context, stageID int64) (int, error) {
	matches, err := s.repo.ListStageMatches(ctx, stageID)
	if err != nil {
		return 0, err
	}
	firstRound := 0
	for _, m := range matches {
		if m.RoundNumber == 1 {
			firstRound++
		}
	}
	if firstRound == 0 {
		return 0, rules.ErrMatchesExist
	}
	return firstRound * 2, nil
}

// NextSwissRound pairs the active participants by record for the next round.
func (s *Service) NextSwissRound(
	ctx context.Context,
	actorID int64,
	actorGlobalRoles []string,
	stageID int64,
	req NextRoundRequest,
) (int, error) {
	stage, err := s.repo.GetStageInfo(ctx, stageID)
	if err != nil {
		return 0, err
	}
	if err := s.permissions(ctx, actorID, actorGlobalRoles, stage.TournamentID, false); err != nil {
		return 0, err
	}
	if stage.FormatCode != "SWISS" {
		return 0, rules.ErrNoNextRound
	}

	maxRound, err := s.repo.MaxRoundNumber(ctx, stageID)
	if err != nil {
		return 0, err
	}
	rounds := swissRounds(*stage)
	if rounds == 0 {
		participants, err := s.repo.ListAcceptedParticipants(ctx, stage.TournamentID)
		if err != nil {
			return 0, err
		}
		rounds = defaultSwissRounds(len(participants))
	}
	if maxRound >= rounds {
		return 0, rules.ErrNoNextRound
	}

	// The previous round must be fully played.
	prevMatches, err := s.repo.ListRoundMatches(ctx, stageID, int64(maxRound))
	if err != nil {
		return 0, err
	}
	for _, m := range prevMatches {
		if m.Status == string(StatusPending) || m.Status == string(StatusInProgress) {
			return 0, rules.ErrNoNextRound
		}
	}

	standings, err := s.repo.ListStandings(ctx, stageID, stage.TournamentID)
	if err != nil {
		return 0, err
	}
	played, err := s.repo.PlayedPairs(ctx, stageID)
	if err != nil {
		return 0, err
	}

	// Mark active participants per swiss config.
	applySwissCutoff(stage.Config, standings)

	_, strategy, err := seedingFor(req.Seeding)
	if err != nil {
		return 0, err
	}
	participants, err := s.repo.ListAcceptedParticipants(ctx, stage.TournamentID)
	if err != nil {
		return 0, err
	}
	seeds, err := strategy.Order(participants, nil)
	if err != nil {
		return 0, err
	}

	pairs := pairSwissRound(standings, played, seeds)
	if len(pairs) == 0 {
		return 0, rules.ErrNoNextRound
	}

	bestOf := stageBestOf(*stage)
	nextRound := maxRound + 1
	matches := make([]Match, 0, len(pairs))
	for _, p := range pairs {
		a, b := p[0], p[1]
		if b == 0 {
			// Free bye: walkover win.
			matches = append(matches, Match{
				RoundNumber: int32(nextRound),
				MatchNumber: int32(len(matches) + 1),
				TeamA:       &a,
				Status:      string(StatusWalkover),
				Winner:      &a,
				BestOf:      bestOf,
			})
			continue
		}
		teamA, teamB := a, b
		matches = append(matches, Match{
			RoundNumber: int32(nextRound),
			MatchNumber: int32(len(matches) + 1),
			TeamA:       &teamA,
			TeamB:       &teamB,
			BestOf:      bestOf,
			Status:      string(StatusPending),
		})
	}
	if err := s.repo.InsertMatches(ctx, stageID, stage.TournamentID, actorID, matches); err != nil {
		return 0, err
	}
	return len(matches), nil
}

// applySwissCutoff marks participants as inactive once they reached the
// configured qualification wins or elimination losses.
func applySwissCutoff(config []byte, standings []StandingsRow) {
	var cfg map[string]interface{}
	if len(config) > 0 {
		_ = jsonUnmarshal(config, &cfg)
	}
	qualWins, _ := cfg["qualification_wins"].(float64)
	elimLosses, _ := cfg["elimination_losses"].(float64)
	for i := range standings {
		standings[i].Active = true
		if qualWins > 0 && float64(standings[i].Wins) >= qualWins {
			standings[i].Active = false
		}
		if elimLosses > 0 && float64(standings[i].Losses) >= elimLosses {
			standings[i].Active = false
		}
	}
}

// Standings returns the current swiss standings of a stage.
func (s *Service) Standings(ctx context.Context, stageID int64) ([]StandingsRow, error) {
	stage, err := s.repo.GetStageInfo(ctx, stageID)
	if err != nil {
		return nil, err
	}
	standings, err := s.repo.ListStandings(ctx, stageID, stage.TournamentID)
	if err != nil {
		return nil, err
	}
	applySwissCutoff(stage.Config, standings)
	return standings, nil
}
