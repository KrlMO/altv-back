package users

import (
	"arena/pkg/utils"
	"context"
	"database/sql"
	"errors"
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
	t, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return errors.New("birth_date must be YYYY-MM-DD")
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		return err
	}

	u := User{
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		NickName:     dto.NickName,
		PasswordHash: hash,
		BirthDate: sql.NullTime{
			Time:  t,
			Valid: true,
		},
		CountryId: dto.CountryId,
		Email:     dto.Email,
		Active:    true,
	}

	return s.repo.CreateUser(ctx, u)
}

func (s *Service) FastRegister(ctx context.Context, dto FastRegisterUserRequest) error {
	t, err := time.Parse("2006-01-02", "1892-01-03")
	if err != nil {
		return errors.New("birth_date must be YYYY-MM-DD")
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		return err
	}

	u := User{
		FirstName:    "Balrog",
		LastName:     "Newbie-" + strconv.Itoa(int(dto.CountryId)),
		NickName:     dto.NickName,
		PasswordHash: hash,
		BirthDate: sql.NullTime{
			Time:  t,
			Valid: true,
		},
		CountryId: dto.CountryId,
		Email:     dto.Email,
		Active:    true,
	}

	return s.repo.CreateUser(ctx, u)
}

func (s *Service) GetByEmailOrNickname(ctx context.Context, value string) (*User, []string, error) {
	u, err := s.repo.GetUserByEmailOrNickname(ctx, value)
	if err != nil {
		return nil, nil, err
	}

	roles, err := s.repo.GetGlobalRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}

	return u, roles, nil
}

func (s *Service) GetProfileByID(ctx context.Context, userID int64) (*UserProfileResponse, error) {
	up, err := s.repo.GetUserProfileByID(ctx, int32(userID))
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
		Email:        up.User.Email,
		ProfilePhoto: up.User.ProfilePhoto,
		BannerPhoto:  up.User.BannerPhoto,
		BirthDate:    birth,
		Country: CountryReponse{
			ID:       up.Country.ID,
			Name:     up.Country.Name,
			FlagPath: flag,
		},
	}

	return &upr, nil
}
