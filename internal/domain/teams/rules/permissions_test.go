package rules

import (
	"testing"

	grules "arena/pkg/global_rules"
)

func TestIsTeamOwnerOrManager(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{"owner", []string{"TEAM_OWNER"}, true},
		{"manager", []string{"TEAM_MANAGER"}, true},
		{"player", []string{"PLAYER"}, false},
		{"empty", []string{}, false},
		{"owner and player", []string{"TEAM_OWNER", "PLAYER"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTeamOwnerOrManager(tt.roles); got != tt.expected {
				t.Errorf("IsTeamOwnerOrManager(%v) = %v, want %v", tt.roles, got, tt.expected)
			}
		})
	}
}

func TestCanAddMember_IluvatarBypass(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"ILUVATAR"},
		},
		TargetRole: "PLAYER",
	}

	if err := CanAddMember(ctx); err != nil {
		t.Errorf("ILUVATAR should bypass permission check, got error: %v", err)
	}
}

func TestCanAddMember_OwnerCanAddPlayer(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"TEAM_OWNER"},
		},
		TargetRole: "PLAYER",
	}

	if err := CanAddMember(ctx); err != nil {
		t.Errorf("Team owner should add player, got error: %v", err)
	}
}

func TestCanAddMember_OwnerCannotAddOwner(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"TEAM_OWNER"},
		},
		TargetRole: "TEAM_OWNER",
	}

	if err := CanAddMember(ctx); err == nil {
		t.Error("Team owner should not add another owner")
	}
}

func TestCanAddMember_ManagerCannotAddManager(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"TEAM_MANAGER"},
		},
		TargetRole: "TEAM_MANAGER",
	}

	if err := CanAddMember(ctx); err == nil {
		t.Error("Team manager should not add another manager")
	}
}

func TestCanAddMember_PlayerCannotAdd(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"PLAYER"},
		},
		TargetRole: "PLAYER",
	}

	if err := CanAddMember(ctx); err == nil {
		t.Error("Player should not be able to add members")
	}
}

func TestCanDeleteMemberFromTeam_IluvatarBypass(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"ILUVATAR"},
		},
		TargetRole: "TEAM_OWNER",
	}

	if err := CanDeleteMemberFromTeam(ctx); err != nil {
		t.Errorf("ILUVATAR should bypass permission check, got error: %v", err)
	}
}

func TestCanDeleteMemberFromTeam_OwnerCannotDeleteOwner(t *testing.T) {
	ctx := TeamActionContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"TEAM_OWNER"},
		},
		TargetRole: "TEAM_OWNER",
	}

	if err := CanDeleteMemberFromTeam(ctx); err == nil {
		t.Error("Team owner should not delete another owner")
	}
}

func TestCanCreateTeamGame_IluvatarBypass(t *testing.T) {
	ctx := TeamGameContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"ILUVATAR"},
		},
	}

	if err := CanCreateTeamGame(ctx); err != nil {
		t.Errorf("ILUVATAR should bypass permission check, got error: %v", err)
	}
}

func TestCanCreateTeamGame_PlayerCannotCreate(t *testing.T) {
	ctx := TeamGameContext{
		Actor: grules.ActorContext{
			ActorGlobalRoles: []string{"HUMAN"},
			ActorTeamRoles:   []string{"PLAYER"},
		},
	}

	if err := CanCreateTeamGame(ctx); err == nil {
		t.Error("Player should not be able to create team games")
	}
}
