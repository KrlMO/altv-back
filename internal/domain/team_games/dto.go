package teamgames

import "arena/pkg/dto"

type PublicTeamGameMembersResponse struct {
	Roaster []TeamGameMemberResponse `json:"roaster"`
	Coach   *TeamGameMemberResponse  `json:"coach"`
}

type TeamGameMemberResponse struct {
	UserID    int64               `json:"user_id"`
	Nickname  string              `json:"nickname"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	Country   dto.CountryResponse `json:"country"`
	Role      string              `json:"role"`
}

type AddTeamGameMembersRequest struct {
	Members []ReqTeamGameMember `json:"members" binding:"required,min=1,max=50"`
	GameId  int32               `json:"game_id" binding:"required"`
}

type ReqTeamGameMember struct {
	UserID         int64
	TeamGameRoleID int32
}

type RemoveTeamGameMembersRequest struct {
	Members []int64 `json:"members" binding:"required,min=1,max=50"`
}

type TeamGameMembersRoleUpdate struct {
	MembersToUpdate []ReqTeamGameMember `json:"members_to_update" binding:"required,min=1,max=50"`
}

type TeamGameRoleDto struct {
	ID   int32
	Name string
}

type RemoveTeamGameRoleFromUserRequest struct {
	Users []TeamGameMemberDto `json:"users" binding:"required,min=1,max=50"`
}

type TeamGameMemberDto struct {
	UserID int64
	Roles  []TeamGameRoleDto
}

type AddTeamGameRoleToMemberRequest struct {
	Users []TeamGameMemberDto `json:"users" binding:"required,min=1,max=50"`
}
