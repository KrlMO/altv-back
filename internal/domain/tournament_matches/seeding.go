package matches

import (
	"fmt"

	"arena/internal/domain/tournament_matches/rules"
)

// SeedingStrategy orders the participants before the generator pairs them.
// New strategies (random, stats-based ranking, etc.) implement this
// interface and register themselves in seedings.
type SeedingStrategy interface {
	// Order returns the participant ids, seed 1 first.
	Order(participants []Participant, manualOrder []int64) ([]int64, error)
}

var seedings = map[string]SeedingStrategy{
	"registration_order": RegistrationOrderSeeding{},
	"manual":             ManualSeeding{},
}

// seedingFor resolves a strategy by name; empty defaults to registration order.
func seedingFor(name string) (string, SeedingStrategy, error) {
	if name == "" {
		name = "registration_order"
	}
	s, ok := seedings[name]
	if !ok {
		return "", nil, fmt.Errorf("%w: %q", rules.ErrUnknownSeeding, name)
	}
	return name, s, nil
}

// RegistrationOrderSeeding is the default deterministic strategy: seed 1 is
// the first accepted registration.
type RegistrationOrderSeeding struct{}

func (RegistrationOrderSeeding) Order(participants []Participant, _ []int64) ([]int64, error) {
	ids := make([]int64, 0, len(participants))
	for _, p := range participants {
		ids = append(ids, p.RegistrationID)
	}
	return ids, nil
}

// ManualSeeding uses the exact order sent by the organizer. It must contain
// every accepted registration exactly once.
type ManualSeeding struct{}

func (ManualSeeding) Order(participants []Participant, manualOrder []int64) ([]int64, error) {
	valid := make(map[int64]bool, len(participants))
	for _, p := range participants {
		valid[p.RegistrationID] = true
	}
	seen := make(map[int64]bool, len(manualOrder))
	for _, id := range manualOrder {
		if !valid[id] {
			return nil, rules.ErrBadManualOrder
		}
		if seen[id] {
			return nil, rules.ErrBadManualOrder
		}
		seen[id] = true
	}
	if len(seen) != len(participants) {
		return nil, rules.ErrBadManualOrder
	}
	return manualOrder, nil
}
