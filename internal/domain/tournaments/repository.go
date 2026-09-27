package tournaments

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FilterTournaments(
	ctx context.Context,
	filter TournamentListFilter,
	page, limit int,
) ([]TournamentListItem, int64, error) {
	where := []string{"true"}
	args := []any{}
	argIdx := 1

	if filter.Name != nil && *filter.Name != "" {
		where = append(where, fmt.Sprintf("at.name ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, *filter.Name)
		argIdx++
	}
	if filter.RegionID != nil {
		where = append(where, fmt.Sprintf("ac.region_id = $%d", argIdx))
		args = append(args, *filter.RegionID)
		argIdx++
	}
	if filter.CountryID != nil {
		where = append(where, fmt.Sprintf("at.country_id = $%d", argIdx))
		args = append(args, *filter.CountryID)
		argIdx++
	}
	if filter.GameID != nil {
		where = append(where, fmt.Sprintf("at.game_id = $%d", argIdx))
		args = append(args, *filter.GameID)
		argIdx++
	}
	if filter.OrganizerType != nil && *filter.OrganizerType != "" {
		where = append(where, fmt.Sprintf("at.organizer_type = $%d", argIdx))
		args = append(args, *filter.OrganizerType)
		argIdx++
	}
	if filter.TypeID != nil {
		where = append(where, fmt.Sprintf("at.type_id = $%d", argIdx))
		args = append(args, *filter.TypeID)
		argIdx++
	}
	if filter.Status != nil {
		switch *filter.Status {
		case "active":
			where = append(where, "at.start_at <= NOW() AND (at.end_at IS NULL OR at.end_at >= NOW())")
		case "upcoming":
			where = append(where, "at.start_at > NOW()")
		case "finished":
			where = append(where, "at.end_at IS NOT NULL AND at.end_at < NOW()")
		}
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM al_tournament at
		LEFT JOIN al_country ac ON at.country_id = ac.id
		WHERE %s
	`, whereClause)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "at.start_at DESC"
	if filter.SortBy != nil {
		switch *filter.SortBy {
		case "start_date_asc":
			orderBy = "at.start_at ASC"
		case "start_date_desc":
			orderBy = "at.start_at DESC"
		case "name_asc":
			orderBy = "at.name ASC"
		case "name_desc":
			orderBy = "at.name DESC"
		case "newest":
			orderBy = "at.created_at DESC"
		}
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT
			at.id,
			at.name,
			at.logo_path,
			at.game_id,
			g.name AS game_name,
			at.organizer_type,
			at.organization_id,
			ao.name AS organization_name,
			at.type_id,
			att.name AS type_name,
			ar.name AS region_name,
			ar.iso_code AS region_iso,
			ac.name AS country_name,
			at.start_at,
			at.end_at,
			at.status,
			at.created_at
		FROM al_tournament at
		JOIN al_game g ON at.game_id = g.id
		LEFT JOIN al_organization ao ON at.organization_id = ao.id
		LEFT JOIN al_tournament_type att ON at.type_id = att.id
		LEFT JOIN al_country ac ON at.country_id = ac.id
		LEFT JOIN al_region ar ON ac.region_id = ar.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []TournamentListItem
	for rows.Next() {
		var item TournamentListItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.LogoKey,
			&item.GameID,
			&item.GameName,
			&item.OrganizerType,
			&item.OrganizationID,
			&item.Organization,
			&item.TypeID,
			&item.TypeName,
			&item.RegionName,
			&item.RegionISO,
			&item.CountryName,
			&item.StartAt,
			&item.EndAt,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}
