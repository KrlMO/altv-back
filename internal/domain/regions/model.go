package regions

import "database/sql"

type Region struct {
	ID       int32          `db:"id"`
	Name     string         `db:"name"`
	IsoCode  string         `db:"iso_code"`
	FlagPath sql.NullString `db:"flag_path"`
}
