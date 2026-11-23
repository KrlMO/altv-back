package rules

import "testing"

func TestRoleRank(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected int
	}{
		{"owner", "OWNER", 0},
		{"admin", "ADMIN", 1},
		{"moderator", "MODERATOR", 2},
		{"staff", "STAFF", 3},
		{"analyst", "ANALYST", 4},
		{"creator", "CREATOR", 5},
		{"unknown role ranks last", "SOMETHING_ELSE", unknownRoleRank},
		{"empty role ranks last", "", unknownRoleRank},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoleRank(tt.role); got != tt.expected {
				t.Errorf("RoleRank(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}
}

func TestRoleRank_OrdersByPrivilege(t *testing.T) {
	hierarchy := []string{"OWNER", "ADMIN", "MODERATOR", "STAFF", "ANALYST", "CREATOR"}

	for i := 1; i < len(hierarchy); i++ {
		previous, current := RoleRank(hierarchy[i-1]), RoleRank(hierarchy[i])
		if previous >= current {
			t.Errorf("RoleRank(%q) = %d should be lower than RoleRank(%q) = %d",
				hierarchy[i-1], previous, hierarchy[i], current)
		}
	}
}
