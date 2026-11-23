package tournaments

// CreateTournamentBody is the JSON payload for tournament creation.
type CreateTournamentBody struct {
	OrganizationID int64   `json:"organization_id" binding:"required"`
	GameID         int32   `json:"game_id" binding:"required"`
	Name           string  `json:"name" binding:"required,max=150"`
	Description    *string `json:"description"`
	StartAt        string  `json:"start_at" binding:"required"`
	EndAt          string  `json:"end_at"`
	ColorHex       *string `json:"color_hex"`
	CountryID      *int32  `json:"country_id"`
	TypeID         *int32  `json:"type_id"`
	OrganizerType  string  `json:"organizer_type" binding:"required,oneof=COMMUNITY OFFICIAL"`
}
