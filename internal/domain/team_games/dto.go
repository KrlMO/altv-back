package teamgames

import (
	"arena/pkg/dto"
	"time"
)

type ReqTeamGameMember struct {
	UserID         int64
	TeamGameRoleID int32
}

type TeamGameMembersRoleUpdate struct {
	MembersToUpdate []ReqTeamGameMember `json:"members_to_update" binding:"required,min=1,max=50"`
}

type TeamGameRoleDto struct {
	ID   int32
	Name string
}

type TeamGameMemberDto struct {
	UserID int64
	Roles  []TeamGameRoleDto
}

type TeamGameListFilter struct {
	Name      *string `form:"name"`
	RegionID  *int32  `form:"region_id"`
	CountryID *int32  `form:"country_id"`
	GameID    *int32  `form:"game_id"`
	SortBy    *string `form:"sort_by"`
	Page      *int    `form:"page"`
}

type TeamGameListItemResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	LogoPath    *string    `json:"logo_path"`
	TeamName    string     `json:"team_name"`
	GameName    string     `json:"game_name"`
	RegionName  *string    `json:"region_name"`
	RegionISO   *string    `json:"region_iso"`
	CountryName *string    `json:"country_name"`
	CreatedAt   *time.Time `json:"created_at"`
}

type TeamGameListSuccessResponse struct {
	Data dto.PaginatedResponse[TeamGameListItemResponse] `json:"data"`
	Msg  string                                           `json:"msg"`
}
