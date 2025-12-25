package teams

import "arena/internal/domain/users"

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
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Tag         string               `json:"tag"`
	Description string               `json:"description"`
	LogoPath    string               `json:"logo_path"`
	BannerPath  string               `json:"banner_path"`
	Country     users.CountryReponse `json:"country"`
	IsVirtual   bool                 `json:"is_virtual"`
	Active      bool                 `json:"active"`
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
	Active      *bool   `json:"active"`
}
