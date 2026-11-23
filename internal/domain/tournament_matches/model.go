package matches

import "time"

type MatchStatus string

const (
	StatusPending    MatchStatus = "PENDING"
	StatusInProgress MatchStatus = "IN_PROGRESS"
	StatusCompleted  MatchStatus = "COMPLETED"
	StatusWalkover   MatchStatus = "WALKOVER"
)

// Match is one game slot inside a stage. Teams are registrations; NULL
// teams mean TBD (next bracket round) or a bye.
type Match struct {
	ID           int64      `db:"id"`
	StageID      int64      `db:"stage_id"`
	TournamentID int64      `db:"tournament_id"`
	RoundNumber  int32      `db:"round_number"`
	MatchNumber  int32      `db:"match_number"`
	TeamA        *int64     `db:"team_a_registration_id"`
	TeamB        *int64     `db:"team_b_registration_id"`
	BestOf       *int32     `db:"best_of"`
	Status       string     `db:"status"`
	ScoreA       *int32     `db:"score_a"`
	ScoreB       *int32     `db:"score_b"`
	Winner       *int64     `db:"winner_registration_id"`
	ScheduledAt  *time.Time `db:"scheduled_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// StageInfo is the subset of the stage needed to generate matches.
type StageInfo struct {
	ID           int64  `db:"id"`
	TournamentID int64  `db:"tournament_id"`
	FormatCode   string `db:"format_code"`
	Status       string `db:"status"`
	BestOf       *int32 `db:"best_of"`
	Config       []byte `db:"config"`
}

// Participant is an accepted registration competing in the stage.
type Participant struct {
	RegistrationID int64  `db:"id"`
	TeamName       string `db:"team_name"`
	GameName       string `db:"game_name"`
}

// TeamLabel resolves a registration id to a displayable team name.
type TeamLabel struct {
	RegistrationID int64  `db:"id"`
	TeamName       string `db:"team_name"`
}
