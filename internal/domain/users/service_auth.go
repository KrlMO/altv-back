package users

import (
	"arena/internal/domain/users/rules"
	"arena/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

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
