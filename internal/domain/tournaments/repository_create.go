package tournaments

import (
	"context"
)

// OrgContext is what the creation rules need about an organization.
type OrgContext struct {
	Exists      bool
	Roles       []string
	Permissions []string
}

// GetOrgContext returns the actor's roles in the org and the permission codes
// granted by the org's active membership.
func (r *Repository) GetOrgContext(ctx context.Context, organizationID, actorID int64) (*OrgContext, error) {
	var oc OrgContext
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM al_organization
			WHERE id = $1 AND COALESCE(active, true)
		)
	`, organizationID).Scan(&oc.Exists)
	if err != nil {
		return nil, err
	}
	if !oc.Exists {
		return nil, ErrTournamentNotFound // reused as "organization not found" by the creation flow
	}

	oc.Roles = []string{}
	rows, err := r.db.Query(ctx, `
		SELECT r.name
		FROM al_user_organization_role uor
		JOIN al_organization_role r ON r.id = uor.role_id
		WHERE uor.user_id = $1 AND uor.organization_id = $2
		  AND COALESCE(uor.active, true)
	`, actorID, organizationID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			rows.Close()
			return nil, err
		}
		oc.Roles = append(oc.Roles, role)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	oc.Permissions = []string{}
	rows, err = r.db.Query(ctx, `
		SELECT DISTINCT p.code
		FROM al_organization_membership om
		JOIN al_membership m ON m.id = om.membership_id
			AND m.deleted_at IS NULL AND m.active
		JOIN al_membership_permission mp ON mp.membership_id = m.id
			AND mp.deleted_at IS NULL
		JOIN al_permission p ON p.id = mp.permission_id AND p.deleted_at IS NULL
		WHERE om.organization_id = $1
		  AND om.deleted_at IS NULL
		  AND (om.expires_at IS NULL OR om.expires_at > NOW())
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		oc.Permissions = append(oc.Permissions, code)
	}
	return &oc, rows.Err()
}

// CountActiveTournaments returns how many live tournaments of the org are not
// finished or cancelled yet.
func (r *Repository) CountActiveTournaments(ctx context.Context, organizationID int64) (total int, official int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE TRUE),
			count(*) FILTER (WHERE organizer_type = 'OFFICIAL')
		FROM al_tournament
		WHERE organization_id = $1
		  AND status NOT IN ('FINISHED', 'CANCELLED')
	`, organizationID).Scan(&total, &official)
	return total, official, err
}

// InsertTournament creates a DRAFT tournament.
func (r *Repository) InsertTournament(ctx context.Context, req *CreateTournamentRequest, createdBy int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO al_tournament
			(organization_id, game_id, name, description, start_at, end_at,
			 status, created_by, color_hex, country_id, type_id, organizer_type)
		VALUES ($1, $2, $3, $4, $5, $6, 'DRAFT', $7, $8, $9, $10, $11)
		RETURNING id
	`, req.OrganizationID, req.GameID, req.Name, req.Description, req.StartAt, req.EndAt,
		createdBy, req.ColorHex, req.CountryID, req.TypeID, req.OrganizerType,
	).Scan(&id)
	return id, err
}
