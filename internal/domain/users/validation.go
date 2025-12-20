package users

import (
	"errors"
	"net/url"
	"regexp"
	"time"
)

func ValidateRegisterOrUpdate(firstName, lastName, nickName, profilePhoto, banner_photo string) error {

	if len(firstName) < 2 || len(firstName) > 50 {
		return errors.New("first_name must be between 2 and 50 characters")
	}

	if len(lastName) < 2 || len(lastName) > 50 {
		return errors.New("last_name must be between 2 and 50 characters")
	}

	if len(nickName) < 3 || len(nickName) > 30 {
		return errors.New("nickname must be between 3 and 30 characters")
	}

	ok, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, nickName)
	if !ok {
		return errors.New("nickname contains invalid characters")
	}

	if profilePhoto != "" {
		_, err := url.ParseRequestURI(profilePhoto)
		if err != nil {
			return errors.New("invalid profile_photo URL")
		}
	}

	if banner_photo != "" {
		_, err := url.ParseRequestURI(banner_photo)
		if err != nil {
			return errors.New("invalid banner_photo URL")
		}
	}

	return nil
}

func ValidateBirthDate(birthDate string) (time.Time, error) {
	if birthDate == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return time.Time{}, errors.New("birth_date must be YYYY-MM-DD")
	}

	now := time.Now()
	if t.After(now) {
		return time.Time{}, errors.New("birth_date cannot be in the future")
	}

	if t.Before(now.AddDate(-120, 0, 0)) {
		return time.Time{}, errors.New("birth_date unrealistically old")
	}
	return t, nil
}
