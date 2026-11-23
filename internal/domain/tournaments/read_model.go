package tournaments

import (
	"database/sql"
	"time"
)

type TournamentListItem struct {
	ID             int64
	Name           string
	LogoKey        sql.NullString
	GameID         int32
	GameName       string
	OrganizerType  sql.NullString
	OrganizationID sql.NullInt64
	Organization   sql.NullString
	TypeID         sql.NullInt32
	TypeName       sql.NullString
	RegionName     sql.NullString
	RegionISO      sql.NullString
	CountryName    sql.NullString
	StartAt        time.Time
	EndAt          sql.NullTime
	Status         string
	CreatedAt      time.Time
}
