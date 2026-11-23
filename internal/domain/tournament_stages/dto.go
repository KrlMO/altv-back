package stages

import "time"

// CreateStageRequest creates a stage for a tournament.
type CreateStageRequest struct {
	FormatCode string                 `json:"format_code" binding:"required"`
	Name       string                 `json:"name" binding:"required,max=150"`
	StageOrder int32                  `json:"stage_order" binding:"required,min=1"`
	Config     map[string]interface{} `json:"config"`
}

// UpdateStageRequest patches a pending stage.
type UpdateStageRequest struct {
	Name       *string                `json:"name" binding:"omitempty,max=150"`
	StageOrder *int32                 `json:"stage_order" binding:"omitempty,min=1"`
	Config     map[string]interface{} `json:"config"`
}

// UpdateStageStatusRequest transitions a stage's status.
type UpdateStageStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=IN_PROGRESS COMPLETED"`
}

// StageResponse is the public visualization of a stage.
type StageResponse struct {
	ID           int64                  `json:"id"`
	TournamentID int64                  `json:"tournament_id"`
	FormatCode   string                 `json:"format_code"`
	FormatName   string                 `json:"format_name"`
	Name         string                 `json:"name"`
	StageOrder   int32                  `json:"stage_order"`
	Status       string                 `json:"status"`
	Config       map[string]interface{} `json:"config"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}
