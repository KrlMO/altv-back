package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/dto"
	"arena/pkg/storage"
	"context"
	"strconv"
)

func (s *Service) toTeamGameListItemResponse(item TeamGameListItem) TeamGameListItemResponse {
	resp := TeamGameListItemResponse{
		ID:       item.ID,
		GameID:   item.GameID,
		TeamName: item.TeamName,
		GameName: item.GameName,
		Active:   item.Active,
	}
	if item.LogoKey.Valid {
		u := s.storage.URL(storage.FolderTeamsLogo, item.LogoKey.String)
		resp.Logo = &u
	}
	if item.RegionName.Valid {
		resp.RegionName = &item.RegionName.String
	}
	if item.RegionISO.Valid {
		resp.RegionISO = &item.RegionISO.String
	}
	if item.CountryName.Valid {
		resp.CountryName = &item.CountryName.String
	}
	if item.CreatedAt.Valid {
		resp.CreatedAt = &item.CreatedAt.Time
	}
	return resp
}

func (s *Service) ListTeamGames(
	ctx context.Context,
	filter TeamGameListFilter,
	page int,
) (dto.PaginatedResponse[TeamGameListItemResponse], error) {
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	items, total, err := s.repo.FilterTeamGames(ctx, filter, page, limit)
	if err != nil {
		return dto.PaginatedResponse[TeamGameListItemResponse]{}, err
	}

	response := make([]TeamGameListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, s.toTeamGameListItemResponse(item))
	}

	return dto.PaginatedResponse[TeamGameListItemResponse]{
		Data: response,
		Meta: dto.PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func (s *Service) ListTeamGamesByTeam(ctx context.Context, teamId string) ([]TeamGameListItemResponse, error) {
	parsedTeamID, err := strconv.ParseInt(teamId, 10, 64)
	if err != nil {
		return nil, rules.ErrBadIdTeamGame
	}

	items, err := s.repo.GetTeamGamesByTeam(ctx, parsedTeamID)
	if err != nil {
		return nil, err
	}

	response := make([]TeamGameListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, s.toTeamGameListItemResponse(item))
	}

	return response, nil
}
