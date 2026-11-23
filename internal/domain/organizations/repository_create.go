package organizations

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MembershipSlugFreeOrganizer is the plan every new organization starts with.
const MembershipSlugFreeOrganizer = "organizer-basic"

// RoleOwner is the role assigned to the creator of an organization.
const RoleOwner = "OWNER"

type CreateRepository struct {
	db *pgxpool.Pool
}

func NewCreateRepository(db *pgxpool.Pool) *CreateRepository {
	return &CreateRepository{db}
}

// CreateOrganization creates the organization, assigns the OWNER role to the
// creator and attaches the free organizer membership — all in one transaction.
func (r *CreateRepository) CreateOrganization(
	ctx context.Context,
	name string,
	description *string,
	createdBy int64,
) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var orgID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO al_organization (name, description, created_by)
		VALUES ($1, $2, $3)
		RETURNING id
	`, name, description, createdBy).Scan(&orgID)
	if err != nil {
		return 0, err
	}

	var roleID int
	err = tx.QueryRow(ctx, `
		SELECT id FROM al_organization_role WHERE name = $1
	`, RoleOwner).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errors.New("organization role OWNER is not seeded")
	}
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_user_organization_role (user_id, organization_id, role_id, assigned_by)
		VALUES ($1, $2, $3, $1)
	`, createdBy, orgID, roleID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO al_organization_membership (organization_id, membership_id, created_by, updated_by)
		SELECT $1, m.id, $2, $2
		FROM al_membership m
		WHERE m.slug = $3 AND m.deleted_at IS NULL AND m.active
	`, orgID, createdBy, MembershipSlugFreeOrganizer)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return orgID, nil
}
