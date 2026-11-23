package matches

import (
	"arena/internal/domain/tournament_matches/rules"
)

// DoubleEliminationGenerator builds the full DE structure: seeded upper
// bracket, lower bracket with TBD teams, and the grand final. Winners advance
// automatically via propagation; UB losers drop to the LB automatically too.
type DoubleEliminationGenerator struct{}

func (DoubleEliminationGenerator) Generate(stage StageInfo, seeds []int64) ([]Match, error) {
	if len(seeds) < 2 {
		return nil, rules.ErrNoParticipants
	}

	size := deSize(len(seeds))
	if size < 4 {
		// 2-team DE is a plain final; fall back to single elim structure.
		return SingleEliminationGenerator{}.Generate(stage, seeds)
	}

	bestOf := stageBestOf(stage)
	matches := make([]Match, 0)

	// Upper bracket: seeded round 1 + TBD placeholder rounds.
	ubPairs := pairFirstRound(seeds)
	ubRounds := int32(upperBracketRounds(size))
	teamsInRound := len(ubPairs)
	for r := int32(1); r <= ubRounds; r++ {
		for i := 0; i < teamsInRound; i++ {
			m := Match{
				RoundNumber: r,
				MatchNumber: int32(i + 1),
				BestOf:      bestOf,
				Status:      string(StatusPending),
			}
			if r == 1 {
				m.TeamA = ubPairs[i][0].registration()
				m.TeamB = ubPairs[i][1].registration()
			}
			matches = append(matches, m)
		}
		teamsInRound /= 2
	}

	// Lower bracket: all TBD.
	lbRoundsCount := lbRounds(size)
	for j := 1; j <= lbRoundsCount; j++ {
		for i := 0; i < lbMatchesInRound(size, j); i++ {
			matches = append(matches, Match{
				RoundNumber: lbGlobalRound(size, j),
				MatchNumber: int32(i + 1),
				BestOf:      bestOf,
				Status:      string(StatusPending),
			})
		}
	}

	// Grand final.
	matches = append(matches, Match{
		RoundNumber: grandFinalRound(size),
		MatchNumber: 1,
		BestOf:      bestOf,
		Status:      string(StatusPending),
	})

	return matches, nil
}
