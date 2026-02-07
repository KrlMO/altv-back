package users

import (
	"arena/pkg/dto"
	"arena/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Register(ctx context.Context, dto RegisterUserRequest) error {
	t, err := ValidateBirthDate(dto.BirthDate)
	if err != nil {
		return err
	}
	err = ValidateRegisterOrUpdate(dto.FirstName, dto.LastName, dto.NickName, "", "")
	if err != nil {
		return err
	}

	exists, err := s.repo.CountryExists(ctx, dto.CountryId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("invalid country")
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		return err
	}

	u := User{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		NickName:  dto.NickName,
		BirthDate: sql.NullTime{
			Time:  t,
			Valid: true,
		},
		CountryId: dto.CountryId,
		Auth: UserAuth{
			PasswordHash: hash,
			ProviderUID:  dto.Email,
		},
		Active: true,
	}

	return s.repo.CreateUser(ctx, u, "local")
}

func (s *Service) FastRegister(ctx context.Context, dto FastRegisterUserRequest) error {
	t, err := time.Parse("2006-01-02", "1892-01-03")
	if err != nil {
		return errors.New("birth_date must be YYYY-MM-DD")
	}

	if len(dto.NickName) < 3 || len(dto.NickName) > 30 {
		return errors.New("nickname must be between 3 and 30 characters")
	}

	ok, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, dto.NickName)
	if !ok {
		return errors.New("nickname contains invalid characters")
	}

	exists, err := s.repo.CountryExists(ctx, dto.CountryId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("invalid country")
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		return err
	}

	u := User{
		FirstName: "Balrog",
		LastName:  "Newbie-" + strconv.Itoa(int(dto.CountryId)),
		NickName:  dto.NickName,
		BirthDate: sql.NullTime{
			Time:  t,
			Valid: true,
		},
		CountryId: dto.CountryId,
		Active:    true,
		Auth: UserAuth{
			ProviderUID:  dto.Email,
			PasswordHash: hash,
		},
	}

	return s.repo.CreateUser(ctx, u, "local")
}

func (s *Service) GetUser(ctx context.Context, providerUID, provider string) (*User, []string, error) {
	u, err := s.repo.GetUserByProviderUID(ctx, providerUID, provider)
	if err != nil {
		return nil, nil, err
	}

	if u.Active == false {
		return nil, nil, errors.New("user does not exist")
	}

	roles, err := s.repo.GetGlobalRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}

	return u, roles, nil
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (*UserProfileResponse, error) {
	up, err := s.repo.GetUserProfileByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	birth := ""
	if up.User.BirthDate.Valid {
		birth = up.User.BirthDate.Time.Format("2006-01-02")
	}

	flag := ""
	if up.Country.FlagPath.Valid {
		flag = up.Country.FlagPath.String
	}

	upr := UserProfileResponse{
		ID:           up.User.ID,
		FirstName:    up.User.FirstName,
		LastName:     up.User.LastName,
		NickName:     up.User.NickName,
		Email:        up.User.Auth.ProviderUID,
		ProfilePhoto: up.User.ProfilePhoto,
		BannerPhoto:  up.User.BannerPhoto,
		BirthDate:    birth,
		Country: dto.CountryResponse{
			ID:       up.Country.ID,
			Name:     up.Country.Name,
			FlagPath: flag,
		},
	}

	return &upr, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, dto UpdateUserProfileRequest) error {
	t, err := ValidateBirthDate(dto.BirthDate)
	if err != nil {
		return err
	}

	err = ValidateRegisterOrUpdate(dto.FirstName, dto.LastName, dto.NickName, dto.ProfilePhoto, dto.BannerPhoto)
	if err != nil {
		return err
	}

	exists, err := s.repo.CountryExists(ctx, dto.CountryId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("invalid country")
	}

	u := User{
		ID:           userID,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		NickName:     dto.NickName,
		ProfilePhoto: dto.ProfilePhoto,
		BannerPhoto:  dto.BannerPhoto,
		BirthDate: sql.NullTime{
			Time:  t,
			Valid: true,
		},
		CountryId: dto.CountryId,
	}

	return s.repo.UpdateUser(ctx, u)
}

func (s *Service) PatchProfile(ctx context.Context, userID int64, dto PatchUserProfileRequest) error {
	u := User{ID: userID}
	if dto.BirthDate != nil {
		t, err := ValidateBirthDate(*dto.BirthDate)
		if err != nil {
			return err
		}
		u.BirthDate = sql.NullTime{
			Time:  t,
			Valid: true,
		}
	}
	if dto.FirstName != nil {
		if len(*dto.FirstName) > 50 || len(*dto.FirstName) < 2 {
			return errors.New("first_name must be between 2 and 50 characters")
		}
		u.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		if len(*dto.LastName) > 50 || len(*dto.LastName) < 2 {
			return errors.New("last_name must be between 2 and 50 characters")
		}
		u.LastName = *dto.LastName
	}
	if dto.NickName != nil {
		if len(*dto.NickName) < 3 || len(*dto.NickName) > 30 {
			return errors.New("nickname must be between 3 and 30 characters")
		}
		ok, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, *dto.NickName)
		if !ok {
			return errors.New("nickname contains invalid characters")
		}
		u.NickName = *dto.NickName
	}
	if dto.ProfilePhoto != nil {
		_, err := regexp.MatchString(`^https?://`, *dto.ProfilePhoto)
		if err != nil {
			return errors.New("invalid profile_photo URL")
		}
		u.ProfilePhoto = *dto.ProfilePhoto
	}
	if dto.BannerPhoto != nil {
		_, err := regexp.MatchString(`^https?://`, *dto.BannerPhoto)
		if err != nil {
			return errors.New("invalid banner_photo URL")
		}
		u.BannerPhoto = *dto.BannerPhoto
	}
	if dto.CountryId != nil {
		exists, err := s.repo.CountryExists(ctx, *dto.CountryId)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("invalid country")
		}
		u.CountryId = *dto.CountryId
	}

	return s.repo.PatchUser(ctx, u)
}
