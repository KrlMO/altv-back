package teams

import "time"

type Team struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Tag         string    `db:"tag"`
	Description string    `db:"description"`
	LogoPath    string    `db:"logo_path"`
	BannerPath  string    `db:"banner_path"`
	CountryId   int32     `db:"country_id"`
	Active      bool      `db:"active"`
	IsVirtual   bool      `db:"is_virtual"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedBy   int32     `db:"created_by"`
	UpdatedBy   int32     `db:"updated_by"`
}
