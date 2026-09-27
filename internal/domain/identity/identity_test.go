package identity

import (
	"errors"
	"testing"
)

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		role     string
		expected bool
	}{
		{"has role", []string{"USER", "PLAYER"}, "PLAYER", true},
		{"does not have role", []string{"USER"}, "PLAYER", false},
		{"empty roles", []string{}, "PLAYER", false},
		{"empty role to find", []string{"USER"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasRole(tt.roles, tt.role); got != tt.expected {
				t.Errorf("HasRole(%v, %q) = %v, want %v", tt.roles, tt.role, got, tt.expected)
			}
		})
	}
}

func TestIsStaffOrSuperAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{"super admin", []string{"SUPER_ADMIN"}, true},
		{"staff", []string{"STAFF"}, true},
		{"user", []string{"USER"}, false},
		{"super admin with others", []string{"SUPER_ADMIN", "USER"}, true},
		{"empty", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStaffOrSuperAdmin(tt.roles); got != tt.expected {
				t.Errorf("IsStaffOrSuperAdmin(%v) = %v, want %v", tt.roles, got, tt.expected)
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
		{"user can interact", []string{"USER"}, false},
		{"banned cannot interact", []string{"BANNED"}, true},
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

func TestCanInteractWithOrganizations(t *testing.T) {
	tests := []struct {
		name    string
		roles   []string
		wantErr bool
	}{
		{"user can interact", []string{"USER"}, false},
		{"banned cannot interact", []string{"BANNED"}, true},
		{"banned among other roles cannot interact", []string{"USER", "BANNED"}, true},
		{"empty roles can interact", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanInteractWithOrganizations(tt.roles)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanInteractWithOrganizations(%v) error = %v, wantErr %v", tt.roles, err, tt.wantErr)
			}
		})
	}
}

func TestCanInteractWithOrganizations_ReturnsSentinelError(t *testing.T) {
	err := CanInteractWithOrganizations([]string{"BANNED"})

	if !errors.Is(err, ErrBannedFromOrganizations) {
		t.Errorf("CanInteractWithOrganizations([BANNED]) error = %v, want %v", err, ErrBannedFromOrganizations)
	}
}
