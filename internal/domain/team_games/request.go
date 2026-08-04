package teamgames

type AddTeamGameMemberRequestItem struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type AddTeamGameMembersRequest struct {
	Members []AddTeamGameMemberRequestItem `json:"members" binding:"required,min=1,max=50"`
	GameId  int32                          `json:"game_id" binding:"required"`
}

type RemoveTeamGameMembersRequest struct {
	Members []int64 `json:"members" binding:"required,min=1,max=50"`
}

type AddTeamGameRoleToMemberRequest struct {
	Users []TeamGameMemberDto `json:"users" binding:"required,min=1,max=50"`
}

type RemoveTeamGameRoleFromUserRequest struct {
	Users []TeamGameMemberDto `json:"users" binding:"required,min=1,max=50"`
}
