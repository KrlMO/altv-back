package users

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int64        `db:"id"`
	FirstName    string       `db:"first_name"`
	LastName     string       `db:"last_name"`
	NickName     string       `db:"nick_name"`
	BirthDate    sql.NullTime `db:"birth_date"`
	RegisteredAt time.Time    `db:"registered_at"`
	CountryId    int32        `db:"country_id"`
	ProfilePhoto string       `db:"profile_photo"`
	BannerPhoto  string       `db:"banner_photo"`
	Active       bool         `db:"active"`
	UpdatedAt    time.Time    `db:"updated_at"`
	Auth         []UserAuth   `db:"auth"`
}

type UserAuth struct {
	ID           int64   `db:"id"`
	ProviderName *string `db:"provider_name"`
	ProviderUID  *string `db:"provider_uid"`
	PasswordHash string  `db:"password_hash"`
}
