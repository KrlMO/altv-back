package registrations

import "time"

// CreateRegistrationRequest is the payload to register a roster to a tournament.
type CreateRegistrationRequest struct {
	TeamGameID int64 `json:"team_game_id" binding:"required"`
}

// ReviewRegistrationRequest is the payload to accept or reject a registration.
type ReviewRegistrationRequest struct {
	Decision string `json:"decision" binding:"required,oneof=ACCEPT REJECT"`
}

// AddRosterMemberRequest adds a player to a registration's roster.
type AddRosterMemberRequest struct {
	UserID    int64 `json:"user_id" binding:"required"`
	IsStarter bool  `json:"is_starter"`
}

// RegistrationResponse is a registration from the organizer/team perspective.
type RegistrationResponse struct {
	ID           int64      `json:"id"`
	TournamentID int64      `json:"tournament_id"`
	TeamGameID   int64      `json:"team_game_id"`
	TeamName     string     `json:"team_name"`
	GameName     string     `json:"game_name"`
	Status       string     `json:"status"`
	RegisteredBy int64      `json:"registered_by"`
	ReviewedBy   *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PublicRegistrationResponse is what anonymous users see: accepted rosters only.
type PublicRegistrationResponse struct {
	ID         int64  `json:"id"`
	TeamGameID int64  `json:"team_game_id"`
	TeamName   string `json:"team_name"`
	GameName   string `json:"game_name"`
	Status     string `json:"status"`
}

// RosterMemberResponse is a player of a registration's roster.
type RosterMemberResponse struct {
	UserID      int64     `json:"user_id"`
	Nickname    string    `json:"nickname"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	IsStarter   bool      `json:"is_starter"`
	JerseyPhoto *string   `json:"jersey_photo,omitempty"`
	AddedAt     time.Time `json:"added_at"`
}
