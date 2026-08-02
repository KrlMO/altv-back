package teams

import "time"

type Team struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Tag         string    `db:"tag"`
	Description string    `db:"description"`
	LogoKey     string    `db:"logo_key"`
	BannerKey   string    `db:"banner_key"`
	CountryId   int32     `db:"country_id"`
	Active      bool      `db:"active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedBy   int64     `db:"created_by"`
	UpdatedBy   int64     `db:"updated_by"`
}

type MyTeamListItem struct {
	ID         int64   `db:"id"`
	Name       string  `db:"name"`
	Tag        string  `db:"tag"`
	LogoKey    *string `db:"logo_key"`
	MemberRole string  `db:"member_role"`
}
