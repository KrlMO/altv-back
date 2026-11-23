package teams

import (
	"arena/pkg/storage"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
)

var allowedImageTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

const maxImageSize = 5 * 1024 * 1024

func validateTeamImage(fh *multipart.FileHeader, field string) (string, error) {
	if fh == nil {
		return "", nil
	}
	if fh.Size > maxImageSize {
		return "", errors.New("The " + field + " image cannot exceed 5MB")
	}
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", errors.New("The " + field + " image must be PNG, JPEG or WEBP")
	}
	return ext, nil
}

func (s *Service) saveTeamImage(ctx context.Context, fh *multipart.FileHeader, folder, field string) (string, error) {
	ext, err := validateTeamImage(fh, field)
	if err != nil {
		return "", err
	}
	filename := storage.RandomFilename(ext)

	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := s.storage.Save(ctx, folder, filename, f, fh.Size, fh.Header.Get("Content-Type")); err != nil {
		return "", err
	}
	return filename, nil
}
