package matches

import (
	"encoding/json"
	"sort"

	"arena/internal/domain/tournament_matches/rules"
)

// SwissGenerator creates the first round of a swiss stage. Subsequent rounds
// are generated on demand (NextRound) pairing by record.
type SwissGenerator struct{}

func (SwissGenerator) Generate(stage StageInfo, seeds []int64) ([]Match, error) {
	if len(seeds) < 2 {
		return nil, rules.ErrNoParticipants
	}

	bestOf := stageBestOf(stage)
	matches := make([]Match, 0, len(seeds)/2)
	for i := 0; i+1 < len(seeds); i += 2 {
		a, b := seeds[i], seeds[i+1]
		matches = append(matches, Match{
			RoundNumber: 1,
			MatchNumber: int32(len(matches) + 1),
			TeamA:       &a,
			TeamB:       &b,
			BestOf:      bestOf,
			Status:      string(StatusPending),
		})
	}
	// Odd participant: bye match (walkover) so the standing is fair.
	if len(seeds)%2 == 1 {
		odd := seeds[len(seeds)-1]
		matches = append(matches, Match{
			RoundNumber: 1,
			MatchNumber: int32(len(matches) + 1),
			TeamA:       &odd,
			Status:      string(StatusWalkover),
			BestOf:      bestOf,
			Winner:      &odd,
		})
	}
	return matches, nil
}

// swissRounds reads the configured round count from the stage config.
func swissRounds(stage StageInfo) int32 {
	var cfg map[string]interface{}
	if len(stage.Config) > 0 {
		if err := json.Unmarshal(stage.Config, &cfg); err != nil {
			return 0
		}
	}
	if v, ok := cfg["rounds"].(float64); ok {
		return int32(v)
	}
	return 0
}

// defaultSwissRounds is the classic log2 sizing for a swiss stage.
func defaultSwissRounds(participants int) int32 {
	rounds := int32(1)
	for 1<<rounds < participants {
		rounds++
	}
	return rounds
}

// pairSwissRound pairs the participants by record (wins descending),
// avoiding rematches when possible. seeds breaks ties deterministically.
// Callers pass only the participants still active in the stage.
func pairSwissRound(standings []StandingsRow, played map[int64]map[int64]bool, seeds []int64) [][2]int64 {
	seedRank := make(map[int64]int, len(seeds))
	for i, id := range seeds {
		seedRank[id] = i
	}

	active := make([]StandingsRow, len(standings))
	copy(active, standings)
	sort.SliceStable(active, func(i, j int) bool {
		if active[i].Wins != active[j].Wins {
			return active[i].Wins > active[j].Wins
		}
		if active[i].Losses != active[j].Losses {
			return active[i].Losses < active[j].Losses
		}
		return seedRank[active[i].RegistrationID] < seedRank[active[j].RegistrationID]
	})

	pairs := make([][2]int64, 0, len(active)/2)
	used := make(map[int64]bool, len(active))
	for i := 0; i < len(active); i++ {
		a := active[i]
		if used[a.RegistrationID] {
			continue
		}
		used[a.RegistrationID] = true
		opponent := int64(0)
		for j := i + 1; j < len(active); j++ {
			b := active[j]
			if used[b.RegistrationID] {
				continue
			}
			if !played[a.RegistrationID][b.RegistrationID] {
				opponent = b.RegistrationID
				break
			}
			// Rematch fallback: remember the closest opponent but keep looking.
			if opponent == 0 {
				opponent = b.RegistrationID
			}
		}
		if opponent != 0 {
			used[opponent] = true
			pairs = append(pairs, [2]int64{a.RegistrationID, opponent})
		} else {
			// Nobody left: free bye (walkover).
			pairs = append(pairs, [2]int64{a.RegistrationID, 0})
		}
	}
	return pairs
}
