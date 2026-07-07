package users

import (
	"arena/internal/domain/users/rules"
	"arena/internal/shared"
	"arena/pkg/dto"
	"arena/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u User, provider string) error
	GetUserByProviderUID(ctx context.Context, providerUID string, provider string) (*User, error)
	GetGlobalRolesByUserID(ctx context.Context, userID int64) ([]string, error)
	GetUserProfileByID(ctx context.Context, userID int64) (*UserProfile, error)
	GetUserByID(ctx context.Context, userId int64) (*User, error)
	GetUserProviders(ctx context.Context, userID int64) ([]ProviderAuth, error)
	UpdateUser(ctx context.Context, u User) error
	PatchUser(ctx context.Context, u User) error
	DeactivateUser(ctx context.Context, userID int64) error
	ActivateUser(ctx context.Context, userID int64) error
	SoftDeleteUser(ctx context.Context, userID int64) error
}

type Service struct {
	repo       UserRepository
	sharedRepo *shared.SharedRepository
}

func NewService(r UserRepository, sharedRepo *shared.SharedRepository) *Service {
	return &Service{repo: r, sharedRepo: sharedRepo}
}

func (s *Service) Register(ctx context.Context, dto RegisterUserRequest) error {
	t, err := ValidateBirthDate(dto.BirthDate)
	if err != nil {
		return rules.ErrInvalidBirthDate
	}
	err = ValidateRegisterOrUpdate(dto.FirstName, dto.LastName, dto.NickName, "", "")
	if err != nil {
		return err
	}

	exists, err := s.sharedRepo.CountryExists(ctx, dto.CountryId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("invalid country")
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		return rules.ErrInvalidPassword
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
		Auth: []ProviderAuth{
			{
				PasswordHash: hash,
				ProviderUID:  &dto.Email,
			},
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

	if err := ValidateNickname(dto.NickName); err != nil {
		return err
	}

	exists, err := s.sharedRepo.CountryExists(ctx, dto.CountryId)
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
		Auth: []ProviderAuth{
			{
				ProviderUID:  &dto.Email,
				PasswordHash: hash,
			},
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

	if len(u.Auth) == 0 {
		return nil, nil, errors.New("user has no auth methods")
	}

	if provider == "local" && !u.Auth[0].IsVerified {
		return nil, nil, rules.ErrEmailNotVerified
	}

	roles, err := s.repo.GetGlobalRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}

	return u, roles, nil
}

func (s *Service) GetUserByID(
	ctx context.Context,
	userId int64,
) (*User, []string, error) {

	u, err := s.repo.GetUserByID(ctx, userId)
	if err != nil {
		return nil, nil, err
	}

	auths, err := s.repo.GetUserProviders(ctx, userId)
	if err != nil {
		return nil, nil, err
	}

	u.Auth = auths

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
		Auth:         mapAuth(up.User.Auth),
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

	exists, err := s.sharedRepo.CountryExists(ctx, dto.CountryId)
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
		if err := ValidateOptionalField(dto.FirstName, 2, 50, "first_name"); err != nil {
			return err
		}
		u.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		if err := ValidateOptionalField(dto.LastName, 2, 50, "last_name"); err != nil {
			return err
		}
		u.LastName = *dto.LastName
	}
	if dto.NickName != nil {
		if err := ValidateNickname(*dto.NickName); err != nil {
			return err
		}
		u.NickName = *dto.NickName
	}
	if dto.ProfilePhoto != nil {
		if err := ValidateOptionalURL(dto.ProfilePhoto, "profile_photo"); err != nil {
			return err
		}
		u.ProfilePhoto = *dto.ProfilePhoto
	}
	if dto.BannerPhoto != nil {
		if err := ValidateOptionalURL(dto.BannerPhoto, "banner_photo"); err != nil {
			return err
		}
		u.BannerPhoto = *dto.BannerPhoto
	}
	if dto.CountryId != nil {
		exists, err := s.sharedRepo.CountryExists(ctx, *dto.CountryId)
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

func (s *Service) DeactivateUser(ctx context.Context, userID int64) error {
	return s.repo.DeactivateUser(ctx, userID)
}

func (s *Service) ActivateUser(ctx context.Context, userID int64) error {
	return s.repo.ActivateUser(ctx, userID)
}

func (s *Service) SoftDeleteUser(ctx context.Context, userID int64) error {
	return s.repo.SoftDeleteUser(ctx, userID)
}

func (s *Service) SelfDelete(ctx context.Context, userID int64) error {
	return s.repo.SoftDeleteUser(ctx, userID)
}
