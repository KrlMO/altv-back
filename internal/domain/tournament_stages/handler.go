package stages

import (
	"errors"
	"net/http"

	"arena/internal/domain/tournament_stages/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")

	tournaments := v1.Group("/tournaments")
	{
		// Public: stage visualization for users (HLTV-like structure view).
		tournaments.GET("/:id/stages", m.ListTournamentStages)

		auth := tournaments.Group("", middleware.Auth(jwtSecret))
		{
			auth.POST("/:id/stages", m.CreateStage)
		}
	}

	stages := v1.Group("/stages")
	{
		auth := stages.Group("", middleware.Auth(jwtSecret))
		{
			auth.PATCH("/:id", m.UpdateStage)
			auth.PATCH("/:id/status", m.TransitionStageStatus)
			auth.DELETE("/:id", m.DeleteStage)
		}
	}
}

// mapError maps domain errors to HTTP status codes.
func mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, rules.ErrBadIdStage):
		c.JSON(http.StatusNotFound, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrUnknownFormat),
		errors.Is(err, rules.ErrInvalidStatusChange),
		errors.Is(err, rules.ErrPreviousStageOpen),
		errors.Is(err, rules.ErrInvalidConfig):
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrStageNotPending):
		c.JSON(http.StatusConflict, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrNoPermission):
		c.JSON(http.StatusForbidden, response.ResponseError{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, response.ResponseError{Error: "internal error"})
	}
}

// CreateStage godoc
// @Summary Create a tournament stage
// @Description Creates a stage for a tournament. Allowed: organization OWNER/ADMIN, tournament TOURNAMENT_OWNER/TOURNAMENT_ADMIN, or global staff. The config is validated against the stage format.
// @Tags TournamentStages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tournament ID"
// @Param body body CreateStageRequest true "Stage payload"
// @Success 201 {object} response.ResponseOk[StageResponse]
// @Failure 400 {object} response.ResponseError
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/tournaments/{id}/stages [post]
func (m *Module) CreateStage(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	tournamentID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid tournament id"})
		return
	}

	var req CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	id, err := m.service.CreateStage(c.Request.Context(), userID, globalRoles, tournamentID, req)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[gin.H]{Data: gin.H{"id": id}})
}

// ListTournamentStages godoc
// @Summary List the stages of a tournament
// @Description Returns the stages of a tournament ordered by stage_order. Public endpoint for visualization.
// @Tags TournamentStages
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} response.ResponseOk[[]StageResponse]
// @Router /api/v1/tournaments/{id}/stages [get]
func (m *Module) ListTournamentStages(c *gin.Context) {
	tournamentID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid tournament id"})
		return
	}

	items, err := m.service.ListTournamentStages(c.Request.Context(), tournamentID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]StageResponse]{Data: items})
}

// UpdateStage godoc
// @Summary Update a pending stage
// @Description Patches the name, order or config of a stage. Only allowed while the stage is PENDING. Config must remain valid for the stage's format.
// @Tags TournamentStages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Stage ID"
// @Param body body UpdateStageRequest true "Stage patch"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/stages/{id} [patch]
func (m *Module) UpdateStage(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	stageID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	var req UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	if err := m.service.UpdateStage(c.Request.Context(), userID, globalRoles, stageID, req); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[bool]{Data: true})
}

// TransitionStageStatus godoc
// @Summary Advance a stage's status
// @Description Transitions a stage's status (PENDING -> IN_PROGRESS -> COMPLETED). A stage cannot start while earlier stages are not completed.
// @Tags TournamentStages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Stage ID"
// @Param body body UpdateStageStatusRequest true "New status"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 400 {object} response.ResponseError
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/stages/{id}/status [patch]
func (m *Module) TransitionStageStatus(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	stageID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	var req UpdateStageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	if err := m.service.TransitionStageStatus(c.Request.Context(), userID, globalRoles, stageID, req.Status); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[bool]{Data: true})
}

// DeleteStage godoc
// @Summary Delete a pending stage
// @Description Soft-deletes a stage. Only allowed while the stage is PENDING.
// @Tags TournamentStages
// @Security BearerAuth
// @Param id path int true "Stage ID"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/stages/{id} [delete]
func (m *Module) DeleteStage(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	stageID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	if err := m.service.DeleteStage(c.Request.Context(), userID, globalRoles, stageID); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[bool]{Data: true})
}

// getStringSlice reads a string slice from the gin context safely.
func getStringSlice(c *gin.Context, key string) []string {
	v, ok := c.Get(key)
	if !ok {
		return nil
	}
	roles, ok := v.([]string)
	if !ok {
		return nil
	}
	return roles
}
