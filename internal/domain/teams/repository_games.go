package teams

import "context"

func (r *Repository) GetGameBySlug(
	ctx context.Context,
	gameSlug string,
) (int32, error) {
	var gameId int32
	err := r.db.QueryRow(ctx, `
			SELECT ag.id FROM al_game ag
			WHERE ag.slug = $1
		`, gameSlug).Scan(&gameId)
	if err != nil {
		return -1, err
	}
	return gameId, nil
}

func (r *Repository) CreateTeamGame(
	ctx context.Context,
	teamId, createdBy int64,
	gameId int32,
) (int64, error) {
	query := `
		INSERT INTO al_team_game(team_id, game_id, created_by, updated_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var teamGameId int64
	if err := r.db.QueryRow(ctx, query, teamId, gameId, createdBy, createdBy).Scan(&teamGameId); err != nil {
		return 0, err
	}
	return teamGameId, nil
}
