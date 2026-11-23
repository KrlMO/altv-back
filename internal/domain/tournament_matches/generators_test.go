package matches

import "testing"

func TestPairFirstRound(t *testing.T) {
	// 4 seeds: 1v4, 2v3
	pairs := pairFirstRound([]int64{10, 20, 30, 40})
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0][0].RegistrationID != 10 || pairs[0][1].RegistrationID != 40 {
		t.Errorf("pair 0 = %v", pairs[0])
	}
	if pairs[1][0].RegistrationID != 20 || pairs[1][1].RegistrationID != 30 {
		t.Errorf("pair 1 = %v", pairs[1])
	}
}

func TestPairFirstRoundByes(t *testing.T) {
	// 3 seeds pad to 4: 1 gets a bye, 2v3
	pairs := pairFirstRound([]int64{10, 20, 30})
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0][0].RegistrationID != 10 || !pairs[0][1].Bye {
		t.Errorf("pair 0 = %v (seed 1 vs bye expected)", pairs[0])
	}
	if pairs[0][1].registration() != nil {
		t.Errorf("bye slot should resolve to nil registration: %v", pairs[0][1])
	}
	if pairs[1][0].RegistrationID != 20 || pairs[1][1].RegistrationID != 30 {
		t.Errorf("pair 1 = %v", pairs[1])
	}
}

func TestSingleEliminationGenerator(t *testing.T) {
	stage := StageInfo{FormatCode: "SINGLE_ELIMINATION", Status: "IN_PROGRESS"}
	bo := int32(3)
	stage.BestOf = &bo

	matches, err := SingleEliminationGenerator{}.Generate(stage, []int64{1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil {
		t.Fatal(err)
	}
	// 8 teams: 4 first-round matches + 2 + 1
	if len(matches) != 7 {
		t.Fatalf("expected 7 matches, got %d", len(matches))
	}
	rounds := map[int32]int{}
	for _, m := range matches {
		rounds[m.RoundNumber]++
		if m.BestOf == nil || *m.BestOf != 3 {
			t.Errorf("match best_of not inherited: %+v", m)
		}
		if m.Status != "PENDING" {
			t.Errorf("match status = %s", m.Status)
		}
	}
	if rounds[1] != 4 || rounds[2] != 2 || rounds[3] != 1 {
		t.Errorf("round distribution = %v", rounds)
	}
	// Later rounds have TBD teams.
	for _, m := range matches {
		if m.RoundNumber > 1 && (m.TeamA != nil || m.TeamB != nil) {
			t.Errorf("round %d should have TBD teams", m.RoundNumber)
		}
	}
}

func TestSingleEliminationGeneratorTooFewTeams(t *testing.T) {
	if _, err := (SingleEliminationGenerator{}).Generate(StageInfo{}, []int64{1}); err == nil {
		t.Fatal("expected error with a single participant")
	}
}

func TestSwissGeneratorRoundOne(t *testing.T) {
	stage := StageInfo{FormatCode: "SWISS", Status: "IN_PROGRESS"}
	matches, err := SwissGenerator{}.Generate(stage, []int64{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if *matches[0].TeamA != 1 || *matches[0].TeamB != 2 {
		t.Errorf("match 0 = %v vs %v", *matches[0].TeamA, *matches[0].TeamB)
	}
	if *matches[1].TeamA != 3 || *matches[1].TeamB != 4 {
		t.Errorf("match 1 = %v vs %v", *matches[1].TeamA, *matches[1].TeamB)
	}
}

func TestSwissGeneratorOddBye(t *testing.T) {
	matches, err := (SwissGenerator{}).Generate(StageInfo{FormatCode: "SWISS"}, []int64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches (1 real + bye), got %d", len(matches))
	}
	bye := matches[1]
	if bye.Status != "WALKOVER" || bye.TeamB != nil || *bye.Winner != 3 {
		t.Errorf("bye match = %+v", bye)
	}
}

func TestPairSwissRoundAvoidsRematches(t *testing.T) {
	standings := []StandingsRow{
		{RegistrationID: 1, Wins: 1},
		{RegistrationID: 2, Wins: 1},
		{RegistrationID: 3, Wins: 0},
		{RegistrationID: 4, Wins: 0},
	}
	played := map[int64]map[int64]bool{1: {2: true}, 2: {1: true}}
	pairs := pairSwissRound(standings, played, []int64{1, 2, 3, 4})
	found := map[[2]int64]bool{}
	for _, p := range pairs {
		found[[2]int64{p[0], p[1]}] = true
	}
	if found[[2]int64{1, 2}] {
		t.Errorf("rematch 1v2 was generated: %v", pairs)
	}
	if !found[[2]int64{1, 3}] && !found[[2]int64{1, 4}] {
		t.Errorf("1 should play 3 or 4: %v", pairs)
	}
}

func TestPairSwissRoundFreeBye(t *testing.T) {
	standings := []StandingsRow{
		{RegistrationID: 1, Wins: 2},
		{RegistrationID: 2, Wins: 1},
		{RegistrationID: 3, Wins: 1},
	}
	pairs := pairSwissRound(standings, map[int64]map[int64]bool{}, []int64{1, 2, 3})
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	// Highest seed pairs first; the leftover gets the free bye.
	if pairs[0][0] != 1 || pairs[0][1] != 2 {
		t.Errorf("pair 0 = %v", pairs[0])
	}
	if pairs[1][0] != 3 || pairs[1][1] != 0 {
		t.Errorf("pair 1 should be a free bye for 3: %v", pairs[1])
	}
}

func TestSeedingStrategies(t *testing.T) {
	participants := []Participant{{RegistrationID: 5}, {RegistrationID: 3}, {RegistrationID: 9}}

	order, err := RegistrationOrderSeeding{}.Order(participants, nil)
	if err != nil || len(order) != 3 || order[0] != 5 {
		t.Errorf("registration order = %v, %v", order, err)
	}

	if _, err := (ManualSeeding{}).Order(participants, []int64{9, 5}); err == nil {
		t.Error("manual seeding must require every participant")
	}
	order, err = (ManualSeeding{}).Order(participants, []int64{9, 5, 3})
	if err != nil || order[0] != 9 || order[2] != 3 {
		t.Errorf("manual order = %v, %v", order, err)
	}

	if _, _, err := seedingFor("bogus"); err == nil {
		t.Error("unknown seeding should fail")
	}
	if _, s, err := seedingFor(""); err != nil || s == nil {
		t.Error("empty seeding should default to registration_order")
	}
}
