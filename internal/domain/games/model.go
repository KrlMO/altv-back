package games

import "database/sql"

type Game struct {
	ID       int            `db:"id"`
	Name     string         `db:"name"`
	ColorHex sql.NullString `db:"color_hex"`
}
