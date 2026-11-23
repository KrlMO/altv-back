package teamgames

import (
	"time"

	"arena/pkg/dto"
)

type TeamGameMemberResponse struct {
	UserID      int64               `json:"user_id"`
	Nickname    string              `json:"nickname"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	Country     dto.CountryResponse `json:"country"`
	Role        string              `json:"role"`
	JerseyPhoto string              `json:"jersey_photo"`
}

type PublicTeamGameMembersResponse struct {
	Roaster []TeamGameMemberResponse `json:"roaster"`
	Coach   *TeamGameMemberResponse  `json:"coach"`
}

type TeamGameManageMemberResponse struct {
	UserID      int64               `json:"user_id"`
	Username    string              `json:"username"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	Avatar      string              `json:"avatar"`
	BirthDate   *time.Time          `json:"birth_date"`
	Country     dto.CountryResponse `json:"country"`
	CountryCode string              `json:"country_code"`
	TeamRole    string              `json:"team_role"`
	JerseyPhoto string              `json:"jersey_photo"`
}
