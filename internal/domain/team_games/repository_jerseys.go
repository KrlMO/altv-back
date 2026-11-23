package teamgames

import "context"

func (r *Repository) UpsertTeamGamePlayerImage(
	ctx context.Context,
	userID, teamGameID, createdBy int64,
	jerseyImagePath string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE al_team_game_player_image
		SET is_current = FALSE
		WHERE user_id = $1 AND team_game_id = $2
	`, userID, teamGameID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO al_team_game_player_image
			(user_id, team_game_id, player_jersey_image_path, is_current, created_by)
		VALUES ($1, $2, $3, TRUE, $4)
	`, userID, teamGameID, jerseyImagePath, createdBy); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
