package identity

import "testing"

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		role     string
		expected bool
	}{
		{"has role", []string{"HUMAN", "PLAYER"}, "PLAYER", true},
		{"does not have role", []string{"HUMAN"}, "PLAYER", false},
		{"empty roles", []string{}, "PLAYER", false},
		{"empty role to find", []string{"HUMAN"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasRole(tt.roles, tt.role); got != tt.expected {
				t.Errorf("HasRole(%v, %q) = %v, want %v", tt.roles, tt.role, got, tt.expected)
			}
		})
	}
}

func TestIsIluvatarOrAinur(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{"iluvatar", []string{"ILUVATAR"}, true},
		{"ainur", []string{"AINUR"}, true},
		{"human", []string{"HUMAN"}, false},
		{"iluvatar with others", []string{"ILUVATAR", "HUMAN"}, true},
		{"empty", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIluvatarOrAinur(tt.roles); got != tt.expected {
				t.Errorf("IsIluvatarOrAinur(%v) = %v, want %v", tt.roles, got, tt.expected)
			}
		})
	}
}

func TestCanInteractWithTeams(t *testing.T) {
	tests := []struct {
		name    string
		roles   []string
		wantErr bool
	}{
		{"human can interact", []string{"HUMAN"}, false},
		{"orc cannot interact", []string{"ORC"}, true},
		{"empty roles can interact", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanInteractWithTeams(tt.roles)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanInteractWithTeams(%v) error = %v, wantErr %v", tt.roles, err, tt.wantErr)
			}
		})
	}
}
