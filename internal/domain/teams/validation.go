package teams

import (
	"errors"
	"net/url"
)

func ValidateTeamCreateOrUpdate(name, tag, description, LogoPath, bannerPath string) error {
	if len(name) < 3 || len(name) > 50 {
		return errors.New("Team name must be between 3 and 50 characters")
	}

	if len(tag) < 2 || len(tag) > 15 {
		return errors.New("Team tag must be between 2 and 15 characters")
	}

	if len(description) > 500 {
		return errors.New("Team description must be less than 500 characters")
	}

	if LogoPath != "" {
		_, err := url.ParseRequestURI(LogoPath)
		if err != nil {
			return errors.New("invalid LogoPath URL")
		}
	}

	if bannerPath != "" {
		_, err := url.ParseRequestURI(bannerPath)
		if err != nil {
			return errors.New("invalid bannerPath URL")
		}
	}

	return nil
}

func ValidateTeamPatch(name, tag, description, logoPath, bannerPath *string) error {
	if name != nil {
		if len(*name) < 3 || len(*name) > 50 {
			return errors.New("Name must be between 3 and 50 characters")
		}
	}
	if tag != nil {
		if len(*tag) < 2 || len(*tag) > 15 {
			return errors.New("Tag must be between 2 and 15 characters")
		}
	}
	if description != nil {
		if len(*description) > 500 {
			return errors.New("Description must be less than 500 characters")
		}
	}
	if logoPath != nil && *logoPath != "" {
		if _, err := url.ParseRequestURI(*logoPath); err != nil {
			return errors.New("invalid logo_path URL")
		}
	}
	if bannerPath != nil && *bannerPath != "" {
		if _, err := url.ParseRequestURI(*bannerPath); err != nil {
			return errors.New("invalid banner_path URL")
		}
	}
	return nil
}
