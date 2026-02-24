package countries

import "database/sql"

type Country struct {
	ID       int32          `db:"id"`
	Name     string         `db:"name"`
	FlagPath sql.NullString `db:"flag_path"`
	IsoCode  string         `db:"iso_code"`
}
