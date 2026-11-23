package registrations

import (
	"errors"
	"net/http"
	"strconv"

	"arena/internal/domain/tournament_registrations/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")

	// Tournament-scoped routes.
	tournaments := v1.Group("/tournaments")
	{
		// Public: only accepted registrations are shown to anonymous users.
		tournaments.GET("/:id/registrations", m.ListTournamentRegistrations)

		auth := tournaments.Group("", middleware.Auth(jwtSecret))
		{
			auth.POST("/:id/registrations", m.CreateRegistration)
		}
	}

	// Registration-scoped routes.
	registrations := v1.Group("/registrations")
	{
		registrations.GET("/:id/roster", m.ListRegistrationRoster)

		auth := registrations.Group("", middleware.Auth(jwtSecret))
		{
			auth.DELETE("/:id", m.CancelRegistration)
			auth.POST("/:id/review", m.ReviewRegistration)
			auth.POST("/:id/roster", m.AddRosterMember)
			auth.DELETE("/:id/roster/:userId", m.RemoveRosterMember)
		}
	}
}

// mapError maps domain errors to HTTP status codes.
func mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, rules.ErrBadIdRegistration),
		errors.Is(err, rules.ErrTeamGameInactive),
		errors.Is(err, rules.ErrRosterMemberNotFound):
		c.JSON(http.StatusNotFound, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrTournamentNotOpen),
		errors.Is(err, rules.ErrGameMismatch),
		errors.Is(err, rules.ErrAlreadyRegistered),
		errors.Is(err, rules.ErrRegistrationNotPending),
		errors.Is(err, rules.ErrUserInAnotherRoster),
		errors.Is(err, rules.ErrNotTeamGameMember):
		c.JSON(http.StatusConflict, response.ResponseError{Error: err.Error()})
	case errors.Is(err, rules.ErrNoPermissionToRegister),
		errors.Is(err, rules.ErrNoPermissionToReview),
		errors.Is(err, rules.ErrNoPermissionToCancel),
		errors.Is(err, rules.ErrNoPermissionToManageRoster):
		c.JSON(http.StatusForbidden, response.ResponseError{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, response.ResponseError{Error: "internal error"})
	}
}

// CreateRegistration godoc
// @Summary Register a roster to a tournament
// @Description Registers a team_game (roster) to a tournament. Only the team owner/manager of the roster's team can register it. The roster's game must match the tournament's game.
// @Tags TournamentRegistrations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tournament ID"
// @Param body body CreateRegistrationRequest true "Registration payload"
// @Success 201 {object} response.ResponseOk[RegistrationResponse]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/tournaments/{id}/registrations [post]
func (m *Module) CreateRegistration(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	tournamentID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid tournament id"})
		return
	}

	var req CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	id, err := m.service.CreateRegistration(c.Request.Context(), userID, globalRoles, tournamentID, req.TeamGameID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[gin.H]{
		Data: gin.H{"id": id},
	})
}

// ListTournamentRegistrations godoc
// @Summary List the registrations of a tournament
// @Description Lists the registrations of a tournament. Anonymous and non-privileged users only see ACCEPTED registrations; organizers and staff see all of them.
// @Tags TournamentRegistrations
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} response.ResponseOk[[]RegistrationResponse]
// @Router /api/v1/tournaments/{id}/registrations [get]
func (m *Module) ListTournamentRegistrations(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	tournamentID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid tournament id"})
		return
	}

	items, err := m.service.ListTournamentRegistrations(c.Request.Context(), userID, globalRoles, tournamentID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]RegistrationResponse]{Data: items})
}

// ReviewRegistration godoc
// @Summary Accept or reject a registration
// @Description Reviews a pending registration. Allowed: organization OWNER/ADMIN, tournament TOURNAMENT_OWNER/TOURNAMENT_ADMIN, or global staff. Rejecting deletes the registration and frees its roster players.
// @Tags TournamentRegistrations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Registration ID"
// @Param body body ReviewRegistrationRequest true "Review decision (ACCEPT | REJECT)"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/registrations/{id}/review [post]
func (m *Module) ReviewRegistration(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	registrationID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid registration id"})
		return
	}

	var req ReviewRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	if err := m.service.ReviewRegistration(c.Request.Context(), userID, globalRoles, registrationID, req.Decision); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[bool]{Data: true})
}

// CancelRegistration godoc
// @Summary Cancel a registration
// @Description Cancels a registration. Allowed: the user who registered it, the roster's team owner/manager, or global staff. The registration is deleted and its roster is freed.
// @Tags TournamentRegistrations
// @Security BearerAuth
// @Param id path int true "Registration ID"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/registrations/{id} [delete]
func (m *Module) CancelRegistration(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	registrationID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid registration id"})
		return
	}

	if err := m.service.CancelRegistration(c.Request.Context(), userID, globalRoles, registrationID); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[bool]{Data: true})
}

// ListRegistrationRoster godoc
// @Summary List the roster of a registration
// @Description Returns the players competing through a registration (starters first).
// @Tags TournamentRegistrations
// @Produce json
// @Param id path int true "Registration ID"
// @Success 200 {object} response.ResponseOk[[]RosterMemberResponse]
// @Router /api/v1/registrations/{id}/roster [get]
func (m *Module) ListRegistrationRoster(c *gin.Context) {
	registrationID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid registration id"})
		return
	}

	members, err := m.service.ListRegistrationRoster(c.Request.Context(), registrationID)
	if err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]RosterMemberResponse]{Data: members})
}

// AddRosterMember godoc
// @Summary Add a player to a registration's roster
// @Description Adds a member of the roster to the registration's tournament roster. A user can only be in one roster per tournament.
// @Tags TournamentRegistrations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Registration ID"
// @Param body body AddRosterMemberRequest true "Roster member payload"
// @Success 201 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/registrations/{id}/roster [post]
func (m *Module) AddRosterMember(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	registrationID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid registration id"})
		return
	}

	var req AddRosterMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	if err := m.service.AddRosterMember(c.Request.Context(), userID, globalRoles, registrationID, req.UserID, req.IsStarter); err != nil {
		mapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[bool]{Data: true})
}

// RemoveRosterMember godoc
// @Summary Remove a player from a registration's roster
// @Description Removes a player from the registration's tournament roster.
// @Tags TournamentRegistrations
// @Security BearerAuth
// @Param id path int true "Registration ID"
// @Param userId path int true "User ID"
// @Success 200 {object} response.ResponseOk[bool]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 404 {object} response.ResponseError
// @Router /api/v1/registrations/{id}/roster/{userId} [delete]
func (m *Module) RemoveRosterMember(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := getStringSlice(c, middleware.ContextRoles)

	registrationID, err := ParseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid registration id"})
		return
	}
	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid user id"})
		return
	}

	if err := m.service.RemoveRosterMember(c.Request.Context(), userID, globalRoles, registrationID, targetUserID); err != nil {
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
