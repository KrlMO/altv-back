package organizations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

// GetUserOrganizations returns one row per (organization, role) pair the user holds,
// along with the permission codes granted by the organization's currently valid plans.
func (r *Repository) GetUserOrganizations(ctx context.Context, userID int64) ([]MyOrganizationRow, error) {
	const query = `
		SELECT
			o.id,
			o.name,
			o.description,
			o.logo_path,
			o.banner_path,
			COALESCE(o.active, true) AS active,
			r.name AS role_name,
			COALESCE(
				ARRAY(
					SELECT DISTINCT p.code
					FROM al_organization_membership om
					JOIN al_membership m
					  ON m.id = om.membership_id
					 AND m.deleted_at IS NULL
					 AND m.active
					JOIN al_membership_permission mp
					  ON mp.membership_id = m.id
					 AND mp.deleted_at IS NULL
					JOIN al_permission p
					  ON p.id = mp.permission_id
					 AND p.deleted_at IS NULL
					WHERE om.organization_id = o.id
					  AND om.deleted_at IS NULL
					  AND (om.expires_at IS NULL OR om.expires_at > NOW())
					ORDER BY p.code
				),
				ARRAY[]::varchar[]
			) AS permissions
		FROM al_user_organization_role uor
		JOIN al_organization_role r ON r.id = uor.role_id
		JOIN al_organization o ON o.id = uor.organization_id
		WHERE uor.user_id = $1
		  AND COALESCE(uor.active, true)
		  AND COALESCE(o.active, true)
		ORDER BY o.id, r.name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs := make([]MyOrganizationRow, 0, 4)
	for rows.Next() {
		var o MyOrganizationRow
		if err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.Description,
			&o.LogoPath,
			&o.BannerPath,
			&o.Active,
			&o.RoleName,
			&o.Permissions,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}
