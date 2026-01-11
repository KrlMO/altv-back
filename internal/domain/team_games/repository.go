package teamgames

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTeamGamePublicMembers(
	ctx context.Context,
	teamGameId int64) ([]TeamGameMember, error) {
	members := []TeamGameMember{}
	rows, err := r.db.Query(ctx, `
			SELECT 
				atg.id,
				atg.team_id,
				atg.game_id,

			FROM al_user au
			JOIN al_user_team_game_role autgr
				ON autgr.user_id = au.id

			JOIN al_team_game_role atgr
				ON autgr.role_id = atgr.id
				AND (atgr.name = "COACH" OR atgr.name = "PLAYER" OR atgr.name = "TEAM_CAPTAIN")

			JOIN al_team_game atg
				ON autgr.team_game_id = atg.id

			WHERE id = $1 
		`, teamGameId)
}
