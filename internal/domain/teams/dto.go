package teams

import (
	"mime/multipart"

	"arena/pkg/dto"
)

type TeamResponse struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Tag         string              `json:"tag"`
	Description string              `json:"description"`
	Logo        string              `json:"logo"`
	Banner      string              `json:"banner"`
	Country     dto.CountryResponse `json:"country"`
	IsVirtual   bool                `json:"is_virtual"`
	Active      bool                `json:"active"`
}

type TeamFilter struct {
	CountryID *int32
	OwnerID   *int64
	Active    *bool
	NameLike  *string
	TagLike   *string
	Limit     int
	Offset    int
}

type PatchUpdateTeamRequest struct {
	Name        *string               `form:"name"`
	Tag         *string               `form:"tag"`
	Description *string               `form:"description"`
	Logo        *multipart.FileHeader `form:"logo"`
	Banner      *multipart.FileHeader `form:"banner"`
	CountryId   *int32                `form:"country_id"`
	IsVirtual   *bool                 `form:"is_virtual"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type UpdateTeamMemberRoleRequest struct {
	Role *string `json:"role" binding:"required"`
}

type CreateTeamGameRequest struct {
	GameSlug   string  `json:"game_slug" binding:"required"`
	Name       string  `json:"name" binding:"required,min=3,max=150"`
	LogoPath   *string `json:"logo_path" binding:"omitempty"`
	BannerPath *string `json:"banner_path" binding:"omitempty"`
	Active     bool    `json:"active" binding:"omitempty"`
}

type MyTeamListItemDto struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Tag        string  `json:"tag"`
	Logo       *string `json:"logo"`
	MemberRole string  `json:"member_role"`
}

type CreateTeamRequestDto struct {
	Name        string                `form:"name" binding:"required,min=3,max=50"`
	Tag         string                `form:"tag" binding:"required,min=2,max=15"`
	Description *string               `form:"description" binding:"omitempty,max=500"`
	CountryId   int32                 `form:"country_id" binding:"required"`
	Logo        *multipart.FileHeader `form:"logo" binding:"omitempty"`
	Banner      *multipart.FileHeader `form:"banner" binding:"omitempty"`
	Roasters    []RoasterEntryDto     `form:"roasters" binding:"omitempty,dive"`
}

type RoasterEntryDto struct {
	GameId  int32              `json:"game_id"`
	Members []RoasterMemberDto `json:"members"`
}

type RoasterMemberDto struct {
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
	TeamRole string `json:"team_role"`
}
