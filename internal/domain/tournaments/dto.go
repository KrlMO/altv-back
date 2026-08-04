package tournaments

import (
	"arena/pkg/dto"
	"time"
)

type TournamentListFilter struct {
	Name          *string `form:"name"`
	RegionID      *int32  `form:"region_id"`
	CountryID     *int32  `form:"country_id"`
	GameID        *int32  `form:"game_id"`
	OrganizerType *string `form:"organizer_type"`
	TypeID        *int32  `form:"type_id"`
	Status        *string `form:"status"`
	SortBy        *string `form:"sort_by"`
}

type TournamentListItemResponse struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Logo           *string    `json:"logo"`
	GameID         int32      `json:"game_id"`
	GameName       string     `json:"game_name"`
	OrganizerType  *string    `json:"organizer_type"`
	OrganizationID *int64     `json:"organization_id"`
	Organization   *string    `json:"organization"`
	TypeID         *int32     `json:"type_id"`
	TypeName       *string    `json:"type_name"`
	RegionName     *string    `json:"region_name"`
	RegionISO      *string    `json:"region_iso"`
	CountryName    *string    `json:"country_name"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          *time.Time `json:"end_at"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

type TournamentListSuccessResponse struct {
	Data dto.PaginatedResponse[TournamentListItemResponse] `json:"data"`
	Msg  string                                            `json:"msg"`
}
