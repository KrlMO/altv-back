package matches

import (
	"errors"
	"net/http"
	"strconv"

	"arena/internal/domain/tournament_matches/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")

	// Stage-scoped routes.
	stages := v1.Group("/stages")
	{
		// Public: match visualization (HLTV-like).
		stages.GET("/:id/matches", m.ListStageMatches)
		stages.GET("/:id/standings", m.Standings)

		auth := stages.Group("", middleware.Auth(jwtSecret))
		{
			auth.POST("/:id/generate", m.GenerateMatches)
			auth.POST("/:id/next-round", m.NextSwissRound)
		}
	}

	matchesGroup := v1.Group("/matches")
	{
		auth := matchesGroup.Group("", middleware.Auth(jwtSecret))
		{
			auth.PATCH("/:id/result", m.SetMatchResult)
		}
	}
}

// mapError maps domain errors to HTTP status codes.
func mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, rules.ErrBadIdMatch),
		errors.Is(err, rules.ErrMatchHasNoTeams):
		c.JSON(http.StatusNotFound, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrUnknownSeeding),
		errors.Is(err, rules.ErrBadManualOrder),
		errors.Is(err, rules.ErrBadScore),
		errors.Is(err, rules.ErrBadWinner),
		errors.Is(err, rules.ErrNoNextRound):
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrStageNotInProgress),
		errors.Is(err, rules.ErrMatchesExist),
		errors.Is(err, rules.ErrMatchNotPlayable),
		errors.Is(err, rules.ErrNoParticipants):
		c.JSON(http.StatusConflict, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrNoPermission):
		c.JSON(http.StatusForbidden, response.ResponseError{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, response.ResponseError{Error: "internal error"})
	}
}

// GenerateMatches godoc
// @Summary Generate the matches of a stage
// @Description Seeds the stage with matches according to its format (single elimination builds the whole bracket; swiss builds round 1). Only allowed while the stage is IN_PROGRESS and it has no matches yet.
// @Tags TournamentMatches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Stage ID"
// @Param body body GenerateMatchesRequest false "Seeding options"
// @Success 201 {object} response.ResponseOk[bool]
// @Failure 400 {object} response.ResponseError
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/stages/{id}/generate [post]
func (m *Module) GenerateMatches(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	stageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	var req GenerateMatchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	n, err := m.service.GenerateMatches(c.Request.Context(), userID, globalRoles, stageID, req)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[gin.H]{Data: gin.H{"matches": n}})
}

// ListStageMatches godoc
// @Summary List the matches of a stage
// @Description Returns the matches of a stage ordered by round and match number. Public endpoint for visualization.
// @Tags TournamentMatches
// @Produce json
// @Param id path int true "Stage ID"
// @Success 200 {object} response.ResponseOk[[]MatchResponse]
// @Router /api/v1/stages/{id}/matches [get]
func (m *Module) ListStageMatches(c *gin.Context) {
	stageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	items, err := m.service.ListStageMatches(c.Request.Context(), stageID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]MatchResponse]{Data: items})
}

// Standings godoc
// @Summary Swiss standings of a stage
// @Description Returns the win/loss record of every accepted participant of the stage, with active flags per the swiss qualification/elimination config.
// @Tags TournamentMatches
// @Produce json
// @Param id path int true "Stage ID"
// @Success 200 {object} response.ResponseOk[[]StandingsRow]
// @Router /api/v1/stages/{id}/standings [get]
func (m *Module) Standings(c *gin.Context) {
	stageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	items, err := m.service.Standings(c.Request.Context(), stageID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]StandingsRow]{Data: items})
}

// NextSwissRound godoc
// @Summary Generate the next swiss round
// @Description Pairs the active participants by record for the next swiss round, avoiding rematches when possible. The previous round must be fully played.
// @Tags TournamentMatches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Stage ID"
// @Success 201 {object} response.ResponseOk[bool]
// @Failure 400 {object} response.ResponseError
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/stages/{id}/next-round [post]
func (m *Module) NextSwissRound(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	stageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid stage id"})
		return
	}

	var req NextRoundRequest
	_ = c.ShouldBindJSON(&req) // body optional

	n, err := m.service.NextSwissRound(c.Request.Context(), userID, globalRoles, stageID, req)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[gin.H]{Data: gin.H{"matches": n}})
}

// SetMatchResult godoc
// @Summary Report a match result
// @Description Sets the score and winner of a match. In single elimination brackets the winner propagates to the next round slot. Allowed: organization OWNER/ADMIN, tournament TOURNAMENT_OWNER/TOURNAMENT_ADMIN/MATCH_ADMIN/REFEREE, or global staff.
// @Tags TournamentMatches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param body body SetMatchResultRequest true "Result payload"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 400 {object} response.ResponseError
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/matches/{id}/result [patch]
func (m *Module) SetMatchResult(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	matchID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid match id"})
		return
	}

	var req SetMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	if err := m.service.SetMatchResult(c.Request.Context(), userID, globalRoles, matchID, req); err != nil {
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
