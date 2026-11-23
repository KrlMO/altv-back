package organizations

import (
	"context"
	"sort"

	"arena/internal/domain/identity"
	"arena/internal/domain/organizations/rules"
	"arena/pkg/storage"
)

type OrganizationRepository interface {
	GetUserOrganizations(ctx context.Context, userID int64) ([]MyOrganizationRow, error)
}

type Service struct {
	repo    OrganizationRepository
	storage storage.Storage
}

func NewService(repo OrganizationRepository, mediaStorage storage.Storage) *Service {
	return &Service{repo: repo, storage: mediaStorage}
}

func (s *Service) GetMyOrganizations(
	ctx context.Context,
	userID int64,
	globalRoles []string,
) ([]OrganizationRoleGroupDto, error) {
	if err := identity.CanInteractWithOrganizations(globalRoles); err != nil {
		return nil, err
	}

	rows, err := s.repo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.groupByRole(rows), nil
}

type organizationAccumulator struct {
	id          int64
	name        string
	description *string
	logo        *string
	banner      *string
	active      bool
	roles       map[string]struct{}
	permissions map[string]struct{}
}

func (a *organizationAccumulator) toDto() MyOrganizationDto {
	return MyOrganizationDto{
		ID:          a.id,
		Name:        a.name,
		Description: a.description,
		Logo:        a.logo,
		Banner:      a.banner,
		Active:      a.active,
		Roles:       sortedKeys(a.roles),
		Permissions: sortedKeys(a.permissions),
	}
}

// groupByRole collapses the (organization, role) rows into one item per organization,
// then emits one group per role holding every organization the user holds that role in.
func (s *Service) groupByRole(rows []MyOrganizationRow) []OrganizationRoleGroupDto {
	byOrgID := make(map[int64]*organizationAccumulator, len(rows))
	orgIDsByRole := make(map[string][]int64, len(rows))

	for _, row := range rows {
		acc, seen := byOrgID[row.ID]
		if !seen {
			acc = &organizationAccumulator{
				id:          row.ID,
				name:        row.Name,
				description: row.Description,
				logo:        s.mediaURL(storage.FolderOrgsLogo, row.LogoPath),
				banner:      s.mediaURL(storage.FolderOrgsBanner, row.BannerPath),
				active:      row.Active,
				roles:       make(map[string]struct{}, 1),
				permissions: make(map[string]struct{}, len(row.Permissions)),
			}
			byOrgID[row.ID] = acc
		}

		acc.roles[row.RoleName] = struct{}{}
		for _, permission := range row.Permissions {
			acc.permissions[permission] = struct{}{}
		}

		if !containsID(orgIDsByRole[row.RoleName], row.ID) {
			orgIDsByRole[row.RoleName] = append(orgIDsByRole[row.RoleName], row.ID)
		}
	}

	groups := make([]OrganizationRoleGroupDto, 0, len(orgIDsByRole))
	for _, role := range sortRolesByRank(orgIDsByRole) {
		orgIDs := orgIDsByRole[role]
		items := make([]MyOrganizationDto, 0, len(orgIDs))
		for _, orgID := range orgIDs {
			items = append(items, byOrgID[orgID].toDto())
		}
		groups = append(groups, OrganizationRoleGroupDto{Role: role, Organizations: items})
	}

	return groups
}

func (s *Service) mediaURL(folder string, path *string) *string {
	if path == nil {
		return nil
	}
	url := s.storage.URL(folder, *path)
	return &url
}

func sortRolesByRank(orgIDsByRole map[string][]int64) []string {
	roles := make([]string, 0, len(orgIDsByRole))
	for role := range orgIDsByRole {
		roles = append(roles, role)
	}

	sort.Slice(roles, func(i, j int) bool {
		rankI, rankJ := rules.RoleRank(roles[i]), rules.RoleRank(roles[j])
		if rankI != rankJ {
			return rankI < rankJ
		}
		return roles[i] < roles[j]
	})

	return roles
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func containsID(ids []int64, id int64) bool {
	for _, current := range ids {
		if current == id {
			return true
		}
	}
	return false
}
