package teamgames

import (
	"context"
	"fmt"
	"strings"
)

func (r *Repository) GetGameIdBySlug(ctx context.Context, gameSlug string) (int32, error) {
	var gameId int32
	err := r.db.QueryRow(ctx, `
		SELECT id
		FROM al_game
		WHERE slug = $1
	`, gameSlug).Scan(&gameId)
	if err != nil {
		return 0, err
	}
	return gameId, nil
}

func (r *Repository) FilterTeamGames(
	ctx context.Context,
	filter TeamGameListFilter,
	page, limit int,
) ([]TeamGameListItem, int64, error) {
	where := []string{"tg.active = true"}
	args := []any{}
	argIdx := 1

	if filter.Name != nil && *filter.Name != "" {
		where = append(where, fmt.Sprintf("t.name ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, *filter.Name)
		argIdx++
	}
	if filter.RegionID != nil {
		where = append(where, fmt.Sprintf("tg.region_id = $%d", argIdx))
		args = append(args, *filter.RegionID)
		argIdx++
	}
	if filter.CountryID != nil {
		where = append(where, fmt.Sprintf("tg.country_id = $%d", argIdx))
		args = append(args, *filter.CountryID)
		argIdx++
	}
	if filter.GameID != nil {
		where = append(where, fmt.Sprintf("tg.game_id = $%d", argIdx))
		args = append(args, *filter.GameID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		WHERE %s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "tg.created_at DESC"
	if filter.SortBy != nil {
		switch *filter.SortBy {
		case "name_asc":
			orderBy = "t.name ASC"
		case "name_desc":
			orderBy = "t.name DESC"
		case "newest":
			orderBy = "tg.created_at DESC"
		case "oldest":
			orderBy = "tg.created_at ASC"
		}
	}

	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT
			tg.id,
			tg.game_id,
			t.logo_key,
			t.name AS team_name,
			g.name AS game_name,
			r.name AS region_name,
			r.iso_code AS region_iso,
			c.name AS country_name,
			tg.active,
			tg.created_at
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		JOIN al_game g ON tg.game_id = g.id
		LEFT JOIN al_region r ON tg.region_id = r.id
		LEFT JOIN al_country c ON tg.country_id = c.id
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

	var items []TeamGameListItem
	for rows.Next() {
		var item TeamGameListItem
		if err := rows.Scan(
			&item.ID,
			&item.GameID,
			&item.LogoKey,
			&item.TeamName,
			&item.GameName,
			&item.RegionName,
			&item.RegionISO,
			&item.CountryName,
			&item.Active,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *Repository) GetTeamGamesByTeam(ctx context.Context, teamId int64) ([]TeamGameListItem, error) {
	query := `
		SELECT
			tg.id,
			tg.game_id,
			t.logo_key,
			t.name AS team_name,
			g.name AS game_name,
			r.name AS region_name,
			r.iso_code AS region_iso,
			c.name AS country_name,
			tg.active,
			tg.created_at
		FROM al_team_game tg
		JOIN al_team t ON tg.team_id = t.id
		JOIN al_game g ON tg.game_id = g.id
		LEFT JOIN al_region r ON tg.region_id = r.id
		LEFT JOIN al_country c ON tg.country_id = c.id
		WHERE tg.team_id = $1
		ORDER BY tg.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TeamGameListItem
	for rows.Next() {
		var item TeamGameListItem
		if err := rows.Scan(
			&item.ID,
			&item.GameID,
			&item.LogoKey,
			&item.TeamName,
			&item.GameName,
			&item.RegionName,
			&item.RegionISO,
			&item.CountryName,
			&item.Active,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if items == nil {
		items = []TeamGameListItem{}
	}

	return items, rows.Err()
}
