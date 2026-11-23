package users

import "context"

func (r *Repository) DeactivateUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_user
		SET active = false, updated_at = NOW()
		WHERE id = $1
	`, userID)
	return err
}

func (r *Repository) ActivateUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE al_user
		SET active = true, updated_at = NOW(), deleted_at = NULL
		WHERE id = $1
	`, userID)
	return err
}

func (r *Repository) SoftDeleteUser(ctx context.Context, userID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE al_user
		SET active = false,
			deleted_at = NOW(),
			updated_at = NOW(),
			first_name = 'Deleted',
			last_name = 'User',
			nickname = 'deleted_' || $1::text,
			profile_photo = '',
			banner_photo = ''
		WHERE id = $1
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE al_user_auth
		SET password_hash = '', updated_at = NOW()
		WHERE user_id = $1 AND provider = 'local'
	`, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM al_refresh_tokens WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM al_refresh_tokens WHERE user_id = $1
	`, userID)
	return err
}
