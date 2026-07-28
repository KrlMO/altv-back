package users

import "arena/pkg/dto"

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
	ID           int64                   `json:"id"`
	FirstName    string                  `json:"first_name"`
	LastName     string                  `json:"last_name"`
	NickName     string                  `json:"nickname"`
	Auth         []ProvidersAuthResponse `json:"auth"`
	ProfilePhoto string                  `json:"profile_photo"`
	BannerPhoto  string                  `json:"banner_photo"`
	BirthDate    string                  `json:"birth_date"`
	Country      dto.CountryResponse     `json:"country"`
}

type ProvidersAuthResponse struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
	ID          int64  `json:"id"`
	IsVerified  bool   `json:"is_verified"`
}

type UpdateUserProfileRequest struct {
	FirstName    string `json:"first_name" binding:"omitempty"`
	LastName     string `json:"last_name" binding:"omitempty"`
	NickName     string `json:"nickname" binding:"omitempty"`
	ProfilePhoto string `json:"profile_photo" binding:"omitempty"`
	BannerPhoto  string `json:"banner_photo" binding:"omitempty"`
	BirthDate    string `json:"birth_date" binding:"omitempty"`
	CountryId    int32  `json:"country_id" binding:"omitempty"`
}

type PatchUserProfileRequest struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	NickName     *string `json:"nickname"`
	BirthDate    *string `json:"birth_date"`
	CountryId    *int32  `json:"country_id"`
	ProfilePhoto *string `json:"profile_photo"`
	BannerPhoto  *string `json:"banner_photo"`
}

type UserSearchResult struct {
	ID           int64               `json:"id"`
	FirstName    string              `json:"first_name"`
	LastName     string              `json:"last_name"`
	NickName     string              `json:"nickname"`
	ProfilePhoto string              `json:"profile_photo"`
	BirthDate    string              `json:"birth_date"`
	Country      dto.CountryResponse `json:"country"`
}
