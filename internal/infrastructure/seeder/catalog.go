package seeder

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type membershipCategory struct {
	Code        string
	Description string
}

type membership struct {
	CategoryCode string
	Name         string
	Slug         string
	Price        float64
	Level        int
}

type permission struct {
	Code        string
	Description string
}

var membershipCategories = []membershipCategory{
	{Code: "USER", Description: ""},
	{Code: "TEAM", Description: ""},
	{Code: "ORGANIZER", Description: ""},
}

var memberships = []membership{
	{CategoryCode: "USER", Name: "Free", Slug: "free", Price: 0.00, Level: 0},
	{CategoryCode: "USER", Name: "VIP", Slug: "vip", Price: 0.99, Level: 1},
	{CategoryCode: "USER", Name: "PRO", Slug: "pro", Price: 4.99, Level: 2},
	{CategoryCode: "TEAM", Name: "Free", Slug: "team-free", Price: 0.00, Level: 0},
	{CategoryCode: "TEAM", Name: "Leader", Slug: "team-leader", Price: 19.99, Level: 1},
	{CategoryCode: "TEAM", Name: "Master", Slug: "team-master", Price: 49.99, Level: 2},
	{CategoryCode: "ORGANIZER", Name: "Free", Slug: "organizer-basic", Price: 0.00, Level: 0},
	{CategoryCode: "ORGANIZER", Name: "Organizer", Slug: "organizer-pro", Price: 39.99, Level: 1},
	{CategoryCode: "ORGANIZER", Name: "Enterprise", Slug: "organizer-enterprise", Price: 99.99, Level: 2},
}

var permissions = []permission{
	{Code: "CREATE_USER", Description: "Permission to create a new user"},
	{Code: "READ_USER", Description: "Permission to read user information"},
	{Code: "UPDATE_USER", Description: "Permission to update user information"},
	{Code: "DELETE_USER", Description: "Permission to delete a user"},
	{Code: "CREATE_TEAM", Description: "Permission to create a new team"},
	{Code: "READ_TEAM", Description: "Permission to read team information"},
	{Code: "UPDATE_TEAM", Description: "Permission to update team information"},
	{Code: "DELETE_TEAM", Description: "Permission to delete a team"},
	{Code: "CREATE_TOURNAMENT", Description: "Permission to create a new tournament"},
	{Code: "CREATE_OFFICIAL_TOURNAMENT", Description: "Permission to create an official tournament"},
	{Code: "VIEW_ADVANCED_ANALYTICS", Description: "Permission to view advanced analytics"},
	{Code: "CUSTOM_PROFILE", Description: "Permission to customize profile"},
	{Code: "CUSTOM_NAME_COLOR", Description: "Permission to customize name color"},
	{Code: "CUSTOM_AVATAR", Description: "Permission to customize avatar"},
	{Code: "CUSTOM_BANNER", Description: "Permission to customize banner"},
	{Code: "CUSTOM_BADGE", Description: "Permission to customize badge"},
	{Code: "CUSTOM_EMOTE", Description: "Permission to customize emote"},
	{Code: "CUSTOM_TAGLINE", Description: "Permission to customize tagline"},
	{Code: "CUSTOM_THEME", Description: "Permission to customize theme"},
	{Code: "CREATE_MULTIPLE_TEAMS", Description: "Permission to create multiple teams"},
	{Code: "CREATE_MULTIPLE_TOURNAMENTS", Description: "Permission to create multiple tournaments"},
	{Code: "CREATE_MULTIPLE_OFFICIAL_TOURNAMENTS", Description: "Permission to create multiple official tournaments"},
}

func SeedCatalog(db *pgxpool.Pool, adminUserID int64) error {
	if adminUserID == 0 {
		log.Println("[seeder] no super admin available, skipping catalog seed")
		return nil
	}

	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	categoryIDs := make(map[string]int, len(membershipCategories))

	for _, c := range membershipCategories {
		var id int

		err = tx.QueryRow(ctx, `
			INSERT INTO al_membership_category (code, description, created_by, updated_by)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (code) DO UPDATE
			SET description = EXCLUDED.description, updated_by = EXCLUDED.updated_by, updated_at = NOW()
			RETURNING id
		`, c.Code, c.Description, adminUserID).Scan(&id)

		if err != nil {
			return err
		}

		categoryIDs[c.Code] = id
	}

	for _, m := range memberships {
		categoryID, ok := categoryIDs[m.CategoryCode]
		if !ok {
			return fmt.Errorf("seeder: membership %q references unknown category %q", m.Slug, m.CategoryCode)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO al_membership (membership_category_id, name, slug, price, level, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			ON CONFLICT (slug) DO NOTHING
		`, categoryID, m.Name, m.Slug, m.Price, m.Level, adminUserID)

		if err != nil {
			return err
		}
	}

	for _, p := range permissions {
		_, err = tx.Exec(ctx, `
			INSERT INTO al_permission (code, description, created_by, updated_by)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (code) DO NOTHING
		`, p.Code, p.Description, adminUserID)

		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Printf("[seeder] catalog seeded: %d categories, %d memberships, %d permissions (actor: %d)\n",
		len(membershipCategories), len(memberships), len(permissions), adminUserID)
	return nil
}
