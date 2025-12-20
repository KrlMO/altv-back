package users

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int32        `db:id`
	FirstName    string       `db:first_name`
	LastName     string       `db:last_name`
	NickName     string       `db:nick_name`
	PasswordHash string       `db:password_hash`
	BirthDate    sql.NullTime `db:birth_date`
	RegisteredAt time.Time    `db:registered_at`
	CountryId    int32        `db:country_id`
	Email        string       `db:email`
	ProfilePhoto string       `db:profile_photo`
	BannerPhoto  string       `db:banner_photo`
	Active       bool         `db:active`
	UpdatedAt    time.Time    `db:updated_at`
}

type Country struct {
	ID       int32          `db:id`
	Name     string         `db:name`
	FlagPath sql.NullString `db:flag_path`
	IsoCode  string         `db:iso_code"`
}
