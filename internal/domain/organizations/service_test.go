package organizations

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"arena/internal/domain/identity"
)

type mockOrganizationRepository struct {
	rows    []MyOrganizationRow
	err     error
	called  bool
	lastUID int64
}

func (m *mockOrganizationRepository) GetUserOrganizations(ctx context.Context, userID int64) ([]MyOrganizationRow, error) {
	m.called = true
	m.lastUID = userID
	return m.rows, m.err
}

type fakeStorage struct{}

func (f *fakeStorage) Save(ctx context.Context, folder, filename string, src io.Reader, size int64, contentType string) error {
	return nil
}

func (f *fakeStorage) URL(folder, filename string) string {
	return "https://media.test/" + folder + "/" + filename
}

func (f *fakeStorage) Delete(ctx context.Context, folder, filename string) error {
	return nil
}

func ptr(s string) *string {
	return &s
}

func TestGetMyOrganizations_BannedUserIsRejected(t *testing.T) {
	repo := &mockOrganizationRepository{}
	svc := NewService(repo, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleBanned})

	if !errors.Is(err, identity.ErrBannedFromOrganizations) {
		t.Fatalf("expected ErrBannedFromOrganizations, got %v", err)
	}
	if groups != nil {
		t.Errorf("expected nil groups, got %v", groups)
	}
	if repo.called {
		t.Error("repository should not be queried for a banned user")
	}
}

func TestGetMyOrganizations_GroupsByRole(t *testing.T) {
	repo := &mockOrganizationRepository{
		rows: []MyOrganizationRow{
			{
				ID: 1, Name: "Acme", Description: ptr("first"),
				LogoPath: ptr("acme.png"), BannerPath: nil, Active: true,
				RoleName:    identity.RoleOrgOwner,
				Permissions: []string{"VIEW_ADVANCED_ANALYTICS", "CREATE_TOURNAMENT"},
			},
			{
				ID: 1, Name: "Acme", Description: ptr("first"),
				LogoPath: ptr("acme.png"), BannerPath: nil, Active: true,
				RoleName:    identity.RoleOrgAnalyst,
				Permissions: []string{"VIEW_ADVANCED_ANALYTICS", "CREATE_TOURNAMENT"},
			},
			{
				ID: 2, Name: "Globex", Description: nil,
				LogoPath: nil, BannerPath: ptr("globex.jpg"), Active: true,
				RoleName:    identity.RoleOrgAnalyst,
				Permissions: []string{"CREATE_TOURNAMENT"},
			},
		},
	}
	svc := NewService(repo, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleUser})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastUID != 7 {
		t.Errorf("expected userID 7, got %d", repo.lastUID)
	}

	acme := MyOrganizationDto{
		ID: 1, Name: "Acme", Description: ptr("first"),
		Logo: ptr("https://media.test/organizations/logo/acme.png"), Banner: nil, Active: true,
		Roles:       []string{"ANALYST", "OWNER"},
		Permissions: []string{"CREATE_TOURNAMENT", "VIEW_ADVANCED_ANALYTICS"},
	}
	globex := MyOrganizationDto{
		ID: 2, Name: "Globex", Description: nil,
		Logo: nil, Banner: ptr("https://media.test/organizations/banner/globex.jpg"), Active: true,
		Roles:       []string{"ANALYST"},
		Permissions: []string{"CREATE_TOURNAMENT"},
	}

	expected := []OrganizationRoleGroupDto{
		{Role: identity.RoleOrgOwner, Organizations: []MyOrganizationDto{acme}},
		{Role: identity.RoleOrgAnalyst, Organizations: []MyOrganizationDto{acme, globex}},
	}

	if !reflect.DeepEqual(groups, expected) {
		t.Errorf("got %+v, want %+v", groups, expected)
	}
}

func TestGetMyOrganizations_OrdersGroupsByRolePrivilege(t *testing.T) {
	repo := &mockOrganizationRepository{
		rows: []MyOrganizationRow{
			{ID: 1, Name: "Acme", Active: true, RoleName: identity.RoleOrgCreator},
			{ID: 2, Name: "Globex", Active: true, RoleName: identity.RoleOrgOwner},
			{ID: 3, Name: "Initech", Active: true, RoleName: identity.RoleOrgAdmin},
		},
	}
	svc := NewService(repo, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleUser})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotRoles := make([]string, 0, len(groups))
	for _, group := range groups {
		gotRoles = append(gotRoles, group.Role)
	}

	expectedRoles := []string{identity.RoleOrgOwner, identity.RoleOrgAdmin, identity.RoleOrgCreator}
	if !reflect.DeepEqual(gotRoles, expectedRoles) {
		t.Errorf("got roles %v, want %v", gotRoles, expectedRoles)
	}
}

func TestGetMyOrganizations_UnknownRoleRanksLast(t *testing.T) {
	repo := &mockOrganizationRepository{
		rows: []MyOrganizationRow{
			{ID: 1, Name: "Acme", Active: true, RoleName: "SOMETHING_ELSE"},
			{ID: 2, Name: "Globex", Active: true, RoleName: identity.RoleOrgOwner},
		},
	}
	svc := NewService(repo, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleUser})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if groups[0].Role != identity.RoleOrgOwner {
		t.Errorf("expected OWNER first, got %s", groups[0].Role)
	}
	if groups[1].Role != "SOMETHING_ELSE" {
		t.Errorf("expected unknown role last, got %s", groups[1].Role)
	}
}

func TestGetMyOrganizations_NoOrganizationsReturnsEmptyGroups(t *testing.T) {
	svc := NewService(&mockOrganizationRepository{}, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleUser})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if groups == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestGetMyOrganizations_PropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("boom")
	svc := NewService(&mockOrganizationRepository{err: wantErr}, &fakeStorage{})

	groups, err := svc.GetMyOrganizations(context.Background(), 7, []string{identity.RoleUser})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if groups != nil {
		t.Errorf("expected nil groups, got %v", groups)
	}
}
