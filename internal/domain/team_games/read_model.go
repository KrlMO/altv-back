package teamgames

import (
	"arena/pkg/model"
	"database/sql"
)

type TeamGameMember struct {
	UserID       int64          `db:"user_id"`
	TeamID       int64          `db:"team_id"`
	TeamGameID   int64          `db:"team_game_id"`
	GameID       int64          `db:"game_id"`
	FirstName    string         `db:"first_name"`
	LastName     string         `db:"last_name"`
	Nickname     string         `db:"nickname"`
	ProfilePhoto sql.NullString `db:"profile_photo"`
	Country      model.Country  `db:"country"`
	TeamGameRole string         `db:"team_game_role"`
	JoinedAt     sql.NullTime   `db:"joined_at"`
	TransferType sql.NullString `db:"transfer_type"`
	PlayerImage  sql.NullString `db:"player_jersey_image_path"`
}
