package rules

import "testing"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name       string
		formatCode string
		cfg        map[string]interface{}
		wantErr    bool
	}{
		{"empty config is valid", FormatSwiss, map[string]interface{}{}, false},
		{"nil handled by caller", FormatSwiss, nil, false},

		// SWISS
		{"swiss valid", FormatSwiss, map[string]interface{}{"rounds": 5, "qualification_wins": 3, "elimination_losses": 3, "best_of": 3}, false},
		{"swiss unknown key", FormatSwiss, map[string]interface{}{"groups": 4}, true},
		{"swiss rounds too high", FormatSwiss, map[string]interface{}{"rounds": 50}, true},

		// GROUP_STAGE
		{"groups valid", FormatGroupStage, map[string]interface{}{"groups": 4, "qualify_per_group": 2, "teams_count": 16, "advancing": 8}, false},
		{"groups unknown key", FormatGroupStage, map[string]interface{}{"rounds": 5}, true},
		{"groups below min", FormatGroupStage, map[string]interface{}{"groups": 0}, true},

		// ELIMINATION
		{"single elim valid", FormatSingleElimination, map[string]interface{}{"best_of": 5, "byes": 1}, false},
		{"double elim valid", FormatDoubleElimination, map[string]interface{}{"best_of": 3}, false},
		{"elim unknown key", FormatSingleElimination, map[string]interface{}{"groups": 4}, true},

		// Generic problems
		{"unknown format", "GSL", map[string]interface{}{}, true},
		{"non numeric value", FormatSwiss, map[string]interface{}{"rounds": "five"}, true},
		{"non integer value", FormatSwiss, map[string]interface{}{"rounds": 2.5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.formatCode, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig(%s, %v) error = %v, wantErr %v", tt.formatCode, tt.cfg, err, tt.wantErr)
			}
		})
	}
}

func TestCanManageStages(t *testing.T) {
	tests := []struct {
		name         string
		orgRoles     []string
		tourneyRoles []string
		globalRoles  []string
		wantErr      error
	}{
		{"org owner", []string{"OWNER"}, nil, nil, nil},
		{"org admin", []string{"ADMIN"}, nil, nil, nil},
		{"org moderator", []string{"MODERATOR"}, nil, nil, ErrNoPermission},
		{"tournament owner", nil, []string{"TOURNAMENT_OWNER"}, nil, nil},
		{"tournament admin", nil, []string{"TOURNAMENT_ADMIN"}, nil, nil},
		{"tournament referee", nil, []string{"REFEREE"}, nil, ErrNoPermission},
		{"global staff", nil, nil, []string{"STAFF"}, nil},
		{"nobody", nil, nil, nil, ErrNoPermission},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanManageStages(tt.orgRoles, tt.tourneyRoles, tt.globalRoles); got != tt.wantErr {
				t.Errorf("CanManageStages() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestCanTransitionStageStatus(t *testing.T) {
	valid := [][2]string{
		{"PENDING", "IN_PROGRESS"},
		{"IN_PROGRESS", "COMPLETED"},
	}
	for _, tc := range valid {
		if err := CanTransitionStageStatus(tc[0], tc[1]); err != nil {
			t.Errorf("%s -> %s should be valid, got %v", tc[0], tc[1], err)
		}
	}

	invalid := [][2]string{
		{"PENDING", "COMPLETED"}, // no jumps
		{"COMPLETED", "PENDING"}, // no going back
		{"IN_PROGRESS", "PENDING"},
		{"COMPLETED", "IN_PROGRESS"},
	}
	for _, tc := range invalid {
		if err := CanTransitionStageStatus(tc[0], tc[1]); err != ErrInvalidStatusChange {
			t.Errorf("%s -> %s should be invalid, got %v", tc[0], tc[1], err)
		}
	}
}
