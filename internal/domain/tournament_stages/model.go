package stages

import "time"

type Stage struct {
	ID           int64      `db:"id"`
	TournamentID int64      `db:"tournament_id"`
	FormatID     int32      `db:"format_id"`
	FormatCode   string     `db:"format_code"`
	FormatName   string     `db:"format_name"`
	Name         string     `db:"name"`
	StageOrder   int32      `db:"stage_order"`
	Status       string     `db:"status"`
	Config       []byte     `db:"config"`
	CreatedBy    int64      `db:"created_by"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// TournamentInfo is the subset of the tournament needed by the rules.
type TournamentInfo struct {
	ID             int64 `db:"id"`
	OrganizationID int64 `db:"organization_id"`
}
