package teams

import "arena/pkg/dto"

type RegisterTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Tag         string `json:"tag" binding:"required,min=2,max=15"`
	Description string `json:"description" binding:"omitempty,max=500"`
	LogoPath    string `json:"logo_path" binding:"omitempty"`
	BannerPath  string `json:"banner_path" binding:"omitempty"`
	CountryId   int32  `json:"country_id" binding:"required"`
	IsVirtual   bool   `json:"is_virtual" binding:"omitempty"`
}

type TeamResponse struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Tag         string             `json:"tag"`
	Description string             `json:"description"`
	LogoPath    string             `json:"logo_path"`
	BannerPath  string             `json:"banner_path"`
	Country     dto.CountryReponse `json:"country"`
	IsVirtual   bool               `json:"is_virtual"`
	Active      bool               `json:"active"`
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
	Name        *string `json:"name"`
	Tag         *string `json:"tag"`
	Description *string `json:"description"`
	LogoPath    *string `json:"logo_path"`
	BannerPath  *string `json:"banner_path"`
	CountryId   *int32  `json:"country_id"`
	IsVirtual   *bool   `json:"is_virtual"`
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
