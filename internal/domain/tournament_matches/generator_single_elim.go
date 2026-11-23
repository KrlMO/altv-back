package matches

import (
	"encoding/json"

	"arena/internal/domain/tournament_matches/rules"
)

// seedSlot is a participant placed in the bracket; Bye marks an empty slot.
type seedSlot struct {
	RegistrationID int64
	Bye            bool
}

func (s seedSlot) registration() *int64 {
	if s.Bye {
		return nil
	}
	id := s.RegistrationID
	return &id
}

// pairFirstRound turns a seed order into first-round pairings using the
// standard bracket interleaving (1 vs n, 2 vs n-1, ...) so top seeds meet as
// late as possible. Odd counts are padded with byes.
func pairFirstRound(seeds []int64) [][2]seedSlot {
	size := 1
	for size < len(seeds) {
		size *= 2
	}
	slots := make([]seedSlot, size)
	for i := range slots {
		slots[i] = seedSlot{Bye: true}
	}
	for i, id := range seeds {
		slots[i] = seedSlot{RegistrationID: id}
	}

	pairs := make([][2]seedSlot, 0, size/2)
	for i := 0; i < size/2; i++ {
		pairs = append(pairs, [2]seedSlot{slots[i], slots[size-1-i]})
	}
	return pairs
}

// totalRoundsFor returns how many rounds a bracket with the given first-round
// pairings has.
func totalRoundsFor(firstRoundPairings int) int32 {
	rounds := int32(1)
	for n := firstRoundPairings; n > 1; n /= 2 {
		rounds++
	}
	return rounds
}

// SingleEliminationGenerator builds the full bracket of a stage. Later rounds
// are created with TBD teams; reported results propagate winners forward.
type SingleEliminationGenerator struct{}

func (SingleEliminationGenerator) Generate(stage StageInfo, seeds []int64) ([]Match, error) {
	if len(seeds) < 2 {
		return nil, rules.ErrNoParticipants
	}

	pairs := pairFirstRound(seeds)
	bestOf := stageBestOf(stage)
	totalRounds := totalRoundsFor(len(pairs))

	matches := make([]Match, 0, len(pairs)*int(totalRounds))
	for i, p := range pairs {
		matches = append(matches, Match{
			RoundNumber: 1,
			MatchNumber: int32(i + 1),
			TeamA:       p[0].registration(),
			TeamB:       p[1].registration(),
			BestOf:      bestOf,
			Status:      string(StatusPending),
		})
	}

	// Subsequent rounds: placeholders to be filled as results come in.
	teamsInRound := len(pairs)
	for r := int32(2); r <= totalRounds; r++ {
		teamsInRound /= 2
		for i := 0; i < teamsInRound; i++ {
			matches = append(matches, Match{
				RoundNumber: r,
				MatchNumber: int32(i + 1),
				BestOf:      bestOf,
				Status:      string(StatusPending),
			})
		}
	}
	return matches, nil
}

// stageBestOf resolves the series length: stage.BestOf first, then the
// best_of key of the stage config.
func stageBestOf(stage StageInfo) *int32 {
	if stage.BestOf != nil {
		return stage.BestOf
	}
	var cfg map[string]interface{}
	if len(stage.Config) > 0 {
		if err := json.Unmarshal(stage.Config, &cfg); err != nil {
			return nil
		}
	}
	if v, ok := cfg["best_of"].(float64); ok {
		bo := int32(v)
		return &bo
	}
	return nil
}
