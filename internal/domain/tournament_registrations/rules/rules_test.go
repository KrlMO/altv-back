package rules

import "testing"

func TestCanRegisterTeamGame(t *testing.T) {
	tests := []struct {
		name        string
		teamRoles   []string
		globalRoles []string
		wantErr     error
	}{
		{"team owner", []string{"TEAM_OWNER"}, nil, nil},
		{"team manager", []string{"TEAM_MANAGER"}, nil, nil},
		{"global staff", nil, []string{"STAFF"}, nil},
		{"super admin", nil, []string{"SUPER_ADMIN"}, nil},
		{"plain member", []string{"PLAYER"}, nil, ErrNoPermissionToRegister},
		{"team staff", []string{"TEAM_STAFF"}, nil, ErrNoPermissionToRegister},
		{"no roles", nil, nil, ErrNoPermissionToRegister},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanRegisterTeamGame(tt.teamRoles, tt.globalRoles); got != tt.wantErr {
				t.Errorf("CanRegisterTeamGame() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestCanReviewRegistration(t *testing.T) {
	tests := []struct {
		name         string
		orgRoles     []string
		tourneyRoles []string
		globalRoles  []string
		wantErr      error
	}{
		{"org owner", []string{"OWNER"}, nil, nil, nil},
		{"org admin", []string{"ADMIN"}, nil, nil, nil},
		{"org moderator", []string{"MODERATOR"}, nil, nil, ErrNoPermissionToReview},
		{"org staff", []string{"STAFF"}, nil, nil, ErrNoPermissionToReview},
		{"tournament owner", nil, []string{"TOURNAMENT_OWNER"}, nil, nil},
		{"tournament admin", nil, []string{"TOURNAMENT_ADMIN"}, nil, nil},
		{"tournament referee", nil, []string{"REFEREE"}, nil, ErrNoPermissionToReview},
		{"global staff", nil, nil, []string{"STAFF"}, nil},
		{"nobody", nil, nil, nil, ErrNoPermissionToReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanReviewRegistration(tt.orgRoles, tt.tourneyRoles, tt.globalRoles); got != tt.wantErr {
				t.Errorf("CanReviewRegistration() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestCanCancelRegistration(t *testing.T) {
	if got := CanCancelRegistration(7, 7, nil, nil); got != nil {
		t.Errorf("registrant can cancel, got %v", got)
	}
	if got := CanCancelRegistration(7, 1, []string{"TEAM_OWNER"}, nil); got != nil {
		t.Errorf("team owner can cancel, got %v", got)
	}
	if got := CanCancelRegistration(7, 1, nil, []string{"SUPER_ADMIN"}); got != nil {
		t.Errorf("global admin can cancel, got %v", got)
	}
	if got := CanCancelRegistration(7, 1, []string{"PLAYER"}, nil); got != ErrNoPermissionToCancel {
		t.Errorf("plain member cannot cancel, got %v", got)
	}
	if got := CanCancelRegistration(7, 1, nil, nil); got != ErrNoPermissionToCancel {
		t.Errorf("stranger cannot cancel, got %v", got)
	}
}

func TestCanManageRegistrationRoster(t *testing.T) {
	tests := []struct {
		name      string
		teamRoles []string
		tgRoles   []string
		globRoles []string
		wantErr   error
	}{
		{"team owner", []string{"TEAM_OWNER"}, nil, nil, nil},
		{"team manager", []string{"TEAM_MANAGER"}, nil, nil, nil},
		{"roster captain", nil, []string{"TEAM_CAPTAIN"}, nil, nil},
		{"roster player", nil, []string{"PLAYER"}, nil, ErrNoPermissionToManageRoster},
		{"global staff", nil, nil, []string{"STAFF"}, nil},
		{"nobody", nil, nil, nil, ErrNoPermissionToManageRoster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanManageRegistrationRoster(tt.teamRoles, tt.tgRoles, tt.globRoles); got != tt.wantErr {
				t.Errorf("CanManageRegistrationRoster() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
