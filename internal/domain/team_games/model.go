package teamgames

import "database/sql"

type TeamGameListItem struct {
	ID          int64          `db:"id"`
	GameID      int32          `db:"game_id"`
	LogoKey     sql.NullString `db:"logo_key"`
	TeamName    string         `db:"team_name"`
	GameName    string         `db:"game_name"`
	RegionName  sql.NullString `db:"region_name"`
	RegionISO   sql.NullString `db:"region_iso"`
	CountryName sql.NullString `db:"country_name"`
	Active      bool           `db:"active"`
	CreatedAt   sql.NullTime   `db:"created_at"`
}
