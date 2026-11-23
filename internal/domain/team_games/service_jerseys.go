package teamgames

import (
	"arena/internal/domain/team_games/rules"
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

func validateJerseyImage(fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "", errors.New("file is required")
	}
	if fh.Size > maxImageSize {
		return "", errors.New("the jersey image cannot exceed 5MB")
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
		return "", errors.New("the jersey image must be PNG, JPEG or WEBP")
	}
	return ext, nil
}

func (s *Service) UploadMemberJersey(
	ctx context.Context,
	currentUserId int64,
	currentUserGlobalRoles []string,
	teamGameId string,
	memberUserId int64,
	file *multipart.FileHeader,
) (string, error) {
	tgc, parsedTeamGameID, err := s.GetTeamGameActionContext(ctx, currentUserId, currentUserGlobalRoles, teamGameId)
	if err != nil {
		return "", err
	}
	if err := rules.CanAddMembersToTeamGame(tgc); err != nil {
		return "", err
	}

	ext, err := validateJerseyImage(file)
	if err != nil {
		return "", err
	}
	filename := storage.RandomFilename(ext)

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := s.storage.Save(ctx, storage.FolderTeamGameJersey, filename, f, file.Size, file.Header.Get("Content-Type")); err != nil {
		return "", err
	}

	if err := s.repo.UpsertTeamGamePlayerImage(ctx, memberUserId, parsedTeamGameID, currentUserId, filename); err != nil {
		return "", err
	}

	return s.storage.URL(storage.FolderTeamGameJersey, filename), nil
}
