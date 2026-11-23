package matches

import "testing"

func TestDoubleEliminationStructure(t *testing.T) {
	stage := StageInfo{FormatCode: "DOUBLE_ELIMINATION", Status: "IN_PROGRESS"}
	matches, err := (DoubleEliminationGenerator{}).Generate(stage, []int64{1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil {
		t.Fatal(err)
	}

	// 8 teams: UB 4+2+1 = 7, LB 2+2+1+1 = 6, GF 1 → 14 matches.
	if len(matches) != 14 {
		t.Fatalf("expected 14 matches, got %d", len(matches))
	}

	rounds := map[int32]int{}
	for _, m := range matches {
		rounds[m.RoundNumber]++
	}
	// UB rounds 1..3, LB global rounds 4..7, GF round 8.
	want := map[int32]int{1: 4, 2: 2, 3: 1, 4: 2, 5: 2, 6: 1, 7: 1, 8: 1}
	for r, n := range want {
		if rounds[r] != n {
			t.Errorf("round %d: expected %d matches, got %d", r, n, rounds[r])
		}
	}
}

func TestDoubleEliminationDrops(t *testing.T) {
	size := 8 // k=3, LB rounds 1..6 global 4..9? (2k-2=4 LB rounds, GF round 8)

	// UB R1 match 1 loser → LB R1 (global 4) match 1 slot a.
	r, m, slot, ok := deNextLoser(size, 1, 1)
	if !ok || r != 4 || m != 1 || slot != "a" {
		t.Errorf("UB R1 M1 loser drop = %d,%d,%s,%v", r, m, slot, ok)
	}
	r, m, slot, _ = deNextLoser(size, 1, 2)
	if r != 4 || m != 1 || slot != "b" {
		t.Errorf("UB R1 M2 loser drop = %d,%d,%s", r, m, slot)
	}
	r, m, slot, _ = deNextLoser(size, 1, 4)
	if r != 4 || m != 2 || slot != "b" {
		t.Errorf("UB R1 M4 loser drop = %d,%d,%s", r, m, slot)
	}

	// UB R2 match 1 loser → LB local 2 (global 5) match 1 slot b.
	r, m, slot, _ = deNextLoser(size, 2, 1)
	if r != 5 || m != 1 || slot != "b" {
		t.Errorf("UB R2 M1 loser drop = %d,%d,%s", r, m, slot)
	}

	// UB R3 (final) loser → LB local 4 (global 7) match 1 slot b.
	r, m, slot, _ = deNextLoser(size, 3, 1)
	if r != 7 || m != 1 || slot != "b" {
		t.Errorf("UB R3 M1 loser drop = %d,%d,%s", r, m, slot)
	}

	// LB losers are eliminated.
	if _, _, _, ok := deNextLoser(size, 4, 1); ok {
		t.Error("LB losers must be eliminated")
	}
}

func TestDoubleEliminationWinners(t *testing.T) {
	size := 8

	// UB winners advance within the UB.
	r, m, slot, _ := deNextWinner(size, 1, 1)
	if r != 2 || m != 1 || slot != "a" {
		t.Errorf("UB R1 M1 winner = %d,%d,%s", r, m, slot)
	}
	r, m, slot, _ = deNextWinner(size, 1, 2)
	if r != 2 || m != 1 || slot != "b" {
		t.Errorf("UB R1 M2 winner = %d,%d,%s", r, m, slot)
	}

	// UB champion waits in GF slot A (GF round 8).
	r, m, slot, _ = deNextWinner(size, 3, 1)
	if r != 8 || m != 1 || slot != "a" {
		t.Errorf("UB champion = %d,%d,%s", r, m, slot)
	}

	// LB R1 winner → LB R2 (global 5) same match index, slot a.
	r, m, slot, _ = deNextWinner(size, 4, 1)
	if r != 5 || m != 1 || slot != "a" {
		t.Errorf("LB R1 winner = %d,%d,%s", r, m, slot)
	}

	// LB R2 (major) winner halves into LB R3.
	r, m, slot, _ = deNextWinner(size, 5, 2)
	if r != 6 || m != 1 || slot != "b" {
		t.Errorf("LB R2 winner = %d,%d,%s", r, m, slot)
	}

	// Last LB winner reaches GF slot B.
	r, m, slot, _ = deNextWinner(size, 7, 1)
	if r != 8 || m != 1 || slot != "b" {
		t.Errorf("LB champion = %d,%d,%s", r, m, slot)
	}

	// GF winner: nothing to propagate.
	if _, _, _, ok := deNextWinner(size, 8, 1); ok {
		t.Error("GF winner must not propagate")
	}
}
