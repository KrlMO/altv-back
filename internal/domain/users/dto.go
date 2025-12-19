package users

type RegisterUserRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	NickName  string `json:"nickname" binding:"required"`
	Password  string `json:"password" binding:"required"`
	BirthDate string `json:"birth_date" binding:"required"`
	CountryId int32  `json:"country_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type FastRegisterUserRequest struct {
	NickName  string `json:"nickname" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	CountryId int32  `json:"country_id" binding:"required"`
}

type UserProfileResponse struct {
	ID           int32          `json:"id"`
	FirstName    string         `json:"first_name"`
	LastName     string         `json:"last_name"`
	NickName     string         `json:"nickname"`
	Email        string         `json:"email"`
	ProfilePhoto string         `json:"profile_photo"`
	BannerPhoto  string         `json:"banner_photo"`
	BirthDate    string         `json:"birth_date"`
	Country      CountryReponse `json:"country"`
}

type CountryReponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	FlagPath string `json:"flag_path"`
}
