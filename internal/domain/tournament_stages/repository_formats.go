package stages

import "context"

// ListStageFormats returns the active stage formats of the catalog.
func (r *Repository) ListStageFormats(ctx context.Context) ([]StageFormatDTO, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, name, description
		FROM al_stage_format
		WHERE active AND deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	formats := []StageFormatDTO{}
	for rows.Next() {
		var f StageFormatDTO
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &f.Description); err != nil {
			return nil, err
		}
		formats = append(formats, f)
	}
	return formats, rows.Err()
}
