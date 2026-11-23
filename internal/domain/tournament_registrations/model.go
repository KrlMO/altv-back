package registrations

import "time"

type RegistrationStatus string

const (
	StatusPending   RegistrationStatus = "PENDING"
	StatusAccepted  RegistrationStatus = "ACCEPTED"
	StatusRejected  RegistrationStatus = "REJECTED"
	StatusCancelled RegistrationStatus = "CANCELLED"
)

// Registration is a roster (team_game) signed up to a tournament.
// Rejected/cancelled registrations are hard-deleted (the roster cascades),
// so a team can re-apply afterwards; the UNIQUE(tournament_id, team_game_id)
// holds for live registrations.
type Registration struct {
	ID           int64      `db:"id"`
	TournamentID int64      `db:"tournament_id"`
	TeamGameID   int64      `db:"team_game_id"`
	Status       string     `db:"status"`
	RegisteredBy int64      `db:"registered_by"`
	ReviewedBy   *int64     `db:"reviewed_by"`
	ReviewedAt   *time.Time `db:"reviewed_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// TournamentInfo is the subset of the tournament needed by the rules.
type TournamentInfo struct {
	ID             int64      `db:"id"`
	GameID         int32      `db:"game_id"`
	Status         string     `db:"status"`
	OrganizationID int64      `db:"organization_id"`
	StartAt        time.Time  `db:"start_at"`
	EndAt          *time.Time `db:"end_at"`
}

// RosterMember is one player slot of a registration's roster.
type RosterMember struct {
	RegistrationID int64     `db:"registration_id"`
	UserID         int64     `db:"user_id"`
	Nickname       string    `db:"nickname"`
	FirstName      string    `db:"first_name"`
	LastName       string    `db:"last_name"`
	IsStarter      bool      `db:"is_starter"`
	JerseyImage    *string   `db:"jersey_image"`
	AddedAt        time.Time `db:"added_at"`
}
