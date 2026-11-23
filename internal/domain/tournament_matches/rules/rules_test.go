package rules

import "testing"

func TestCanEditTeams(t *testing.T) {
	// teamA=1, teamB=2 current occupants.
	tests := []struct {
		name   string
		status string
		newA   int64
		newB   int64
		want   error
	}{
		{"fill empty slot b in progress", "IN_PROGRESS", 0, 2, nil},
		{"fill empty slot a in progress", "IN_PROGRESS", 1, 0, nil},
		{"replace team in progress", "IN_PROGRESS", 5, 0, ErrMatchNotEditable},
		{"replace team b in progress", "IN_PROGRESS", 0, 7, ErrMatchNotEditable},
		{"completed is frozen", "COMPLETED", 0, 2, ErrMatchNotEditable},
		{"walkover is frozen", "WALKOVER", 5, 0, ErrMatchNotEditable},
		{"pending replace both", "PENDING", 3, 4, nil},
		{"pending replace a only", "PENDING", 3, 0, nil},
		{"same team both slots", "PENDING", 3, 3, ErrTeamDuplicated},
		{"new a equals team b", "PENDING", 2, 0, ErrTeamDuplicated},
		{"new b equals team a", "PENDING", 0, 1, ErrTeamDuplicated},
		{"no slots", "PENDING", 0, 0, ErrNoSlotsToUpdate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanEditTeams(tt.status, 1, 2, tt.newA, tt.newB); got != tt.want {
				t.Errorf("CanEditTeams(%s, 1, 2, %d, %d) = %v, want %v",
					tt.status, tt.newA, tt.newB, got, tt.want)
			}
		})
	}
}

func TestValidateResultErrors(t *testing.T) {
	// best_of 3: winner needs exactly 2 wins.
	if err := ValidateResult("IN_PROGRESS", 1, 2, 1, 2, 0, 3); err != nil {
		t.Errorf("valid 2-0 rejected: %v", err)
	}
	if err := ValidateResult("IN_PROGRESS", 1, 2, 1, 2, 1, 3); err != nil {
		t.Errorf("valid 2-1 rejected: %v", err)
	}
	if err := ValidateResult("IN_PROGRESS", 1, 2, 1, 5, 0, 3); err == nil {
		t.Error("5-0 should be rejected in best_of 3")
	}
	if err := ValidateResult("IN_PROGRESS", 1, 2, 1, 1, 0, 3); err == nil {
		t.Error("1-0 should be rejected in best_of 3")
	}
	if err := ValidateResult("IN_PROGRESS", 1, 2, 3, 2, 0, 3); err == nil {
		t.Error("foreign winner should be rejected")
	}
	if err := ValidateResult("COMPLETED", 1, 2, 1, 2, 0, 3); err == nil {
		t.Error("completed match should not accept results")
	}
	// Walkover skips the score check.
	if err := ValidateResult("WALKOVER", 1, 2, 1, 0, 0, 3); err != nil {
		t.Errorf("walkover with 0-0 rejected: %v", err)
	}
}
