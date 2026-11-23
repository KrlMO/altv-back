package storage

import (
	"context"
	"errors"
	"io"
	"strings"
)

const (
	DriverLocal = "local"
	DriverVPS   = "vps"
	DriverS3    = "s3"
)

const (
	FolderTeamsLogo      = "teams/logo"
	FolderTeamsBanner    = "teams/banner"
	FolderTeamGameJersey = "team_games/jersey"
	FolderOrgsLogo       = "organizations/logo"
	FolderOrgsBanner     = "organizations/banner"
)

var ErrNotImplemented = errors.New("storage driver not implemented yet")

type Storage interface {
	Save(ctx context.Context, folder, filename string, src io.Reader, size int64, contentType string) error
	URL(folder, filename string) string
	Delete(ctx context.Context, folder, filename string) error
}

func New(driver, mediaURL, mediaRoot string) (Storage, error) {
	switch strings.ToLower(driver) {
	case DriverLocal, DriverVPS:
		if mediaRoot == "" {
			return nil, errors.New("MEDIA_ROOT is required for " + driver + " storage driver")
		}
		return NewLocal(mediaRoot, mediaURL), nil
	case DriverS3:
		return NewS3(mediaURL), nil
	default:
		return nil, errors.New("unknown STORAGE_DRIVER: " + driver)
	}
}
