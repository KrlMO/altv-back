package teamgames

import "arena/pkg/dto"

type TeamGameMemberResponse struct {
	UserID    int64               `json:"user_id"`
	Nickname  string              `json:"nickname"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	Country   dto.CountryResponse `json:"country"`
	Role      string              `json:"role"`
}

type PublicTeamGameMembersResponse struct {
	Roaster []TeamGameMemberResponse `json:"roaster"`
	Coach   *TeamGameMemberResponse  `json:"coach"`
}
