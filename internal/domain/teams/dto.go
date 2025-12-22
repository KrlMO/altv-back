package teams

type RegisterTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Tag         string `json:"tag" binding:"required,min=2,max=15"`
	Description string `json:"description" binding:"omitempty,max=500"`
	LogoPath    string `json:"logo_path" binding:"omitempty"`
	BannerPath  string `json:"banner_path" binding:"omitempty"`
	CountryId   int32  `json:"country_id" binding:"required"`
	IsVirtual   bool   `json:"is_virtual" binding:"omitempty"`
}
