package organizations

import "time"

type Organization struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	LogoPath    *string   `db:"logo_path"`
	BannerPath  *string   `db:"banner_path"`
	CreatedBy   int64     `db:"created_by"`
	Active      bool      `db:"active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// MyOrganizationRow is one (organization, role) pair for the authenticated user.
// A user holding several roles in the same organization produces several rows.
type MyOrganizationRow struct {
	ID          int64    `db:"id"`
	Name        string   `db:"name"`
	Description *string  `db:"description"`
	LogoPath    *string  `db:"logo_path"`
	BannerPath  *string  `db:"banner_path"`
	Active      bool     `db:"active"`
	RoleName    string   `db:"role_name"`
	Permissions []string `db:"permissions"`
}
