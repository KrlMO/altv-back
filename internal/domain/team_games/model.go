package teamgames

import "database/sql"

type TeamGame struct {
	ID        int    `db:"id"`
	TeamID    int    `db:"team_id"`
	GameID    int    `db:"game_id"`
	Name      string `db:"name"`
	LogoPath  string `db:"logo_path"`
	RegionID  *int32 `db:"region_id"`
	CountryID *int32 `db:"country_id"`
}

type TeamGameListItem struct {
	ID           int64          `db:"id"`
	Name         string         `db:"name"`
	LogoPath     sql.NullString `db:"logo_path"`
	TeamName     string         `db:"team_name"`
	GameName     string         `db:"game_name"`
	RegionName   sql.NullString `db:"region_name"`
	RegionISO    sql.NullString `db:"region_iso"`
	CountryName  sql.NullString `db:"country_name"`
	CreatedAt    sql.NullTime   `db:"created_at"`
}
