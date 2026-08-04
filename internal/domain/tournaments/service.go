package tournaments

import (
	"arena/pkg/dto"
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTournaments(
	ctx context.Context,
	filter TournamentListFilter,
	page int,
) (dto.PaginatedResponse[TournamentListItemResponse], error) {
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	items, total, err := s.repo.FilterTournaments(ctx, filter, page, limit)
	if err != nil {
		return dto.PaginatedResponse[TournamentListItemResponse]{}, err
	}

	response := make([]TournamentListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, s.toListItemResponse(item))
	}

	return dto.PaginatedResponse[TournamentListItemResponse]{
		Data: response,
		Meta: dto.PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func (s *Service) toListItemResponse(item TournamentListItem) TournamentListItemResponse {
	resp := TournamentListItemResponse{
		ID:        item.ID,
		Name:      item.Name,
		GameID:    item.GameID,
		GameName:  item.GameName,
		StartAt:   item.StartAt,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
	}
	if item.LogoKey.Valid {
		resp.Logo = &item.LogoKey.String
	}
	if item.OrganizerType.Valid {
		resp.OrganizerType = &item.OrganizerType.String
	}
	if item.OrganizationID.Valid {
		resp.OrganizationID = &item.OrganizationID.Int64
	}
	if item.Organization.Valid {
		resp.Organization = &item.Organization.String
	}
	if item.TypeID.Valid {
		resp.TypeID = &item.TypeID.Int32
	}
	if item.TypeName.Valid {
		resp.TypeName = &item.TypeName.String
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
	if item.EndAt.Valid {
		resp.EndAt = &item.EndAt.Time
	}
	return resp
}
