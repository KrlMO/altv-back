package teamgames

import "arena/pkg/dto"

type PublicTeamGameMembersResponse struct {
	Roaster []TeamGameMemberResponse `json:"roaster"`
	Coach   *TeamGameMemberResponse  `json:"coach"`
}

type TeamGameMemberResponse struct {
	UserID    int64              `json:"user_id"`
	Nickname  string             `json:"nickname"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	Country   dto.CountryReponse `json:"country"`
	Role      string             `json:"role"`
}

type AddTeamGameMemberRequest struct {
	Members []NewTeamGameMember `json:"members" binding:"required,min=1,max=50"`
	GameId  int32               `json:"game_id" binding:"required"`
}

type NewTeamGameMember struct {
	UserID         int64
	TeamGameRoleID int32
}
