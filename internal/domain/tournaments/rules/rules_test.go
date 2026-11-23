package rules

import "testing"

func TestCanCreateTournament(t *testing.T) {
	basicPerms := []string{"CREATE_TOURNAMENT"}
	proPerms := []string{"CREATE_TOURNAMENT", "CREATE_MULTIPLE_TOURNAMENTS", "CREATE_OFFICIAL_TOURNAMENT"}
	enterprisePerms := []string{"CREATE_TOURNAMENT", "CREATE_MULTIPLE_TOURNAMENTS",
		"CREATE_OFFICIAL_TOURNAMENT", "CREATE_MULTIPLE_OFFICIAL_TOURNAMENTS"}

	tests := []struct {
		name           string
		roles          []string
		perms          []string
		globalRoles    []string
		organizerType  string
		active         int
		activeOfficial int
		want           error
	}{
		{"owner with basic creates community", []string{"OWNER"}, basicPerms, nil, "COMMUNITY", 0, 0, nil},
		{"creator with basic creates community", []string{"CREATOR"}, basicPerms, nil, "COMMUNITY", 0, 0, nil},
		{"moderator cannot create", []string{"MODERATOR"}, basicPerms, nil, "COMMUNITY", 0, 0, ErrNoPermission},
		{"no role cannot create", nil, basicPerms, nil, "COMMUNITY", 0, 0, ErrNoPermission},
		{"basic plan without CREATE_TOURNAMENT", []string{"OWNER"}, nil, nil, "COMMUNITY", 0, 0, ErrOrgNoCreatePermission},
		{"basic second active tournament blocked", []string{"OWNER"}, basicPerms, nil, "COMMUNITY", 1, 0, ErrActiveLimitReached},
		{"pro second active tournament allowed", []string{"OWNER"}, proPerms, nil, "COMMUNITY", 3, 0, nil},
		{"basic official blocked", []string{"OWNER"}, basicPerms, nil, "OFFICIAL", 0, 0, ErrOfficialNotAllowed},
		{"pro official allowed", []string{"OWNER"}, proPerms, nil, "OFFICIAL", 0, 0, nil},
		{"pro second official blocked", []string{"OWNER"}, proPerms, nil, "OFFICIAL", 0, 1, ErrOfficialLimitReached},
		{"enterprise second official allowed", []string{"OWNER"}, enterprisePerms, nil, "OFFICIAL", 2, 1, nil},
		{"global staff bypasses everything", nil, nil, []string{"STAFF"}, "OFFICIAL", 5, 5, nil},
		{"super admin bypasses everything", nil, nil, []string{"SUPER_ADMIN"}, "OFFICIAL", 5, 5, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanCreateTournament(tt.roles, tt.perms, tt.globalRoles, tt.organizerType, tt.active, tt.activeOfficial); got != tt.want {
				t.Errorf("CanCreateTournament() = %v, want %v", got, tt.want)
			}
		})
	}
}
