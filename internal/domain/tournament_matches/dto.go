package matches

// SetMatchTeamsRequest fills the team slots of a match. Zero/omitted slots
// are left untouched.
type SetMatchTeamsRequest struct {
	TeamARegistrationID int64 `json:"team_a_registration_id"`
	TeamBRegistrationID int64 `json:"team_b_registration_id"`
}

// GenerateMatchesRequest seeds a stage's matches.
type GenerateMatchesRequest struct {
	// Seeding strategy: "registration_order" (default). Future: "random",
	// "manual", "stats".
	Seeding string `json:"seeding" binding:"omitempty,oneof=registration_order manual"`
	// ManualOrder is the explicit participant order (seed 1 first). Required
	// when seeding = "manual".
	ManualOrder []int64 `json:"manual_order"`
}

// NextRoundRequest generates the next swiss round.
type NextRoundRequest struct {
	// Seeding strategy for the pairing fallbacks within equal records.
	Seeding string `json:"seeding" binding:"omitempty,oneof=registration_order"`
}

// SetMatchResultRequest reports a match result.
type SetMatchResultRequest struct {
	ScoreA int32 `json:"score_a" binding:"min=0"`
	ScoreB int32 `json:"score_b" binding:"min=0"`
}

// MatchResponse is the public visualization of a match.
type MatchResponse struct {
	ID           int64     `json:"id"`
	StageID      int64     `json:"stage_id"`
	TournamentID int64     `json:"tournament_id"`
	RoundNumber  int32     `json:"round_number"`
	MatchNumber  int32     `json:"match_number"`
	TeamA        *TeamInfo `json:"team_a,omitempty"`
	TeamB        *TeamInfo `json:"team_b,omitempty"`
	BestOf       *int32    `json:"best_of,omitempty"`
	Status       string    `json:"status"`
	ScoreA       *int32    `json:"score_a,omitempty"`
	ScoreB       *int32    `json:"score_b,omitempty"`
	Winner       *TeamInfo `json:"winner,omitempty"`
	ScheduledAt  *string   `json:"scheduled_at,omitempty"`
}

type TeamInfo struct {
	RegistrationID int64  `json:"registration_id"`
	TeamName       string `json:"team_name"`
}

// StandingsRow is one line of the swiss standings.
type StandingsRow struct {
	RegistrationID int64  `json:"registration_id"`
	TeamName       string `json:"team_name"`
	Wins           int    `json:"wins"`
	Losses         int    `json:"losses"`
	Active         bool   `json:"active"` // still playing (not qualified/eliminated)
}
