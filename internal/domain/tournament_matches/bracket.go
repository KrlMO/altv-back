package matches

// Double-elimination bracket math, shared by the generator and the result
// propagation. All rounds are global numbers:
//
//	1..k              upper bracket (k = log2 of the padded size)
//	k+1..k+(2k-2)     lower bracket (local rounds 1..2k-2)
//	k+(2k-1)          grand final
//
// Conventions (standard DE):
//   - UB round 1 losers pair 2-by-1 into LB round 1 (match i takes the losers
//     of UB matches 2i-1 and 2i).
//   - UB round r>=2 losers drop into LB local round 2(r-1), slot B, facing
//     the LB local round 2(r-2) winners in slot A.
//   - LB major (odd local >1) rounds halve; winners feed slot A/B by parity.
//   - The last LB winner reaches the grand final slot B; UB champion slot A.

// upperBracketRounds returns k for a padded bracket size (power of two).
func upperBracketRounds(size int) int {
	rounds := 0
	for n := size; n > 1; n /= 2 {
		rounds++
	}
	return rounds
}

// lbRounds returns how many lower bracket local rounds exist (0 if size < 4).
func lbRounds(size int) int {
	k := upperBracketRounds(size)
	if k < 2 {
		return 0
	}
	return 2*k - 2
}

// lbMatchesInRound returns the number of LB matches in local round j.
func lbMatchesInRound(size, j int) int {
	switch {
	case j == 1:
		return size / 4
	case j%2 == 0:
		// major round vs UB round j/2+1 losers: same count as that UB round
		return size / (1 << (j/2 + 1))
	default:
		// odd rounds halve the previous even round
		return size / (1 << ((j+1)/2 + 1))
	}
}

// lbGlobalRound maps a local LB round to the global round number.
func lbGlobalRound(size, j int) int32 {
	return int32(upperBracketRounds(size) + j)
}

// grandFinalRound returns the global round number of the grand final.
func grandFinalRound(size int) int32 {
	return int32(upperBracketRounds(size) + lbRounds(size) + 1)
}

// deNextWinner computes where the winner of a match goes.
// Returns the target (round, match, slot) or ok=false for the GF champion.
func deNextWinner(size int, round int32, matchNumber int32) (int32, int32, string, bool) {
	k := int32(upperBracketRounds(size))
	lbLast := int32(lbRounds(size))
	switch {
	case round <= k:
		// Upper bracket: winner advances within the UB.
		if round == k {
			if lbRounds(size) == 0 {
				return 0, 0, "", false // size-2 bracket: the UB final is the final
			}
			return grandFinalRound(size), 1, "a", true // UB champion waits in GF slot A
		}
		return round + 1, (matchNumber + 1) / 2, slotForMatch(matchNumber), true
	case round <= k+lbLast:
		j := int(round - k) // LB local round
		switch {
		case j == int(lbLast):
			return grandFinalRound(size), 1, "b", true // LB champion in GF slot B
		case j%2 == 0:
			// major round winner advances to a halving (odd) round
			return lbGlobalRound(size, j+1), (matchNumber + 1) / 2, slotForMatch(matchNumber), true
		default:
			// minor round winner advances to the next major round, same index
			return lbGlobalRound(size, j+1), matchNumber, "a", true
		}
	default:
		return 0, 0, "", false
	}
}

// deNextLoser computes where the loser of an UB match drops (only UB matches
// produce losers; LB losers are eliminated). ok=false if eliminated or final.
func deNextLoser(size int, round int32, matchNumber int32) (int32, int32, string, bool) {
	k := int32(upperBracketRounds(size))
	if round > k || lbRounds(size) == 0 {
		return 0, 0, "", false // LB losers are out; size-2 brackets have no LB
	}
	switch round {
	case 1:
		// UB R1 losers pair into LB round 1: matches 2i-1 and 2i feed match i.
		return lbGlobalRound(size, 1), (matchNumber + 1) / 2, slotForMatch(matchNumber), true
	default:
		// UB round r>=2 losers drop to LB local round 2(r-1), slot B.
		return lbGlobalRound(size, 2*int(round-1)), matchNumber, "b", true
	}
}

// slotForMatch returns the slot name by match parity inside a round.
func slotForMatch(matchNumber int32) string {
	if matchNumber%2 == 1 {
		return "a"
	}
	return "b"
}

// deSize returns the padded bracket size for the given seed count.
func deSize(seeds int) int {
	size := 1
	for size < seeds {
		size *= 2
	}
	return size
}
