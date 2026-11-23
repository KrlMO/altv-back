package tournaments

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"arena/internal/domain/tournaments/rules"
	"arena/pkg/dto"
	"arena/pkg/middleware"
	"arena/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/v1/tournaments", m.ListTournaments)
	r.GET("/api/v1/tournaments/:id", m.GetTournamentByID)

	auth := r.Group("/api/v1/tournaments", middleware.Auth(m.jwtSecret))
	{
		auth.POST("", m.CreateTournament)
	}
}

// CreateTournament godoc
// @Summary Create a tournament
// @Description Creates a DRAFT tournament for an organization. Requires the CREATE_TOURNAMENT permission from the organization's active membership, plus the OWNER/ADMIN/CREATOR role. OFFICIAL tournaments need CREATE_OFFICIAL_TOURNAMENT. Free plans can only keep one active tournament.
// @Tags Tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateTournamentBody true "Tournament payload"
// @Success 201 {object} response.ResponseOk[bool]
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 409 {object} response.ResponseError
// @Router /api/v1/tournaments [post]
func (m *Module) CreateTournament(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	globalRoles := c.GetStringSlice(middleware.ContextRoles)

	var body CreateTournamentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid request: " + err.Error()})
		return
	}

	req := CreateTournamentRequest{
		OrganizationID: body.OrganizationID,
		GameID:         body.GameID,
		Name:           body.Name,
		Description:    body.Description,
		ColorHex:       body.ColorHex,
		CountryID:      body.CountryID,
		TypeID:         body.TypeID,
		OrganizerType:  body.OrganizerType,
	}
	startAt, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "start_at must be RFC3339"})
		return
	}
	req.StartAt = startAt
	if body.EndAt != "" {
		endAt, err := time.Parse(time.RFC3339, body.EndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "end_at must be RFC3339"})
			return
		}
		req.EndAt = &endAt
	}

	id, err := m.service.CreateTournament(c.Request.Context(), userID, globalRoles, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTournamentNotFound):
			c.JSON(http.StatusNotFound, response.ResponseError{Error: "organization or game not found"})
		case errors.Is(err, rules.ErrNoPermission),
			errors.Is(err, rules.ErrOrgNoCreatePermission),
			errors.Is(err, rules.ErrOfficialNotAllowed):
			c.JSON(http.StatusForbidden, response.ResponseError{Error: err.Error()})
		case errors.Is(err, rules.ErrActiveLimitReached),
			errors.Is(err, rules.ErrOfficialLimitReached):
			c.JSON(http.StatusConflict, response.ResponseError{Error: err.Error()})
		default:
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, response.ResponseOk[gin.H]{Data: gin.H{"id": id}})
}

// GetTournamentByID godoc
// @Summary      Tournament detail
// @Description  Get a single tournament by id.
// @Tags         Tournaments
// @Produce      json
// @Param        id   path  int  true  "Tournament ID"
// @Success      200  {object}  response.ResponseOk[TournamentListItemResponse]
// @Failure      404  {object}  response.ResponseError
// @Router       /api/v1/tournaments/{id} [get]
func (m *Module) GetTournamentByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid tournament id"})
		return
	}

	tournament, err := m.service.GetTournamentByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTournamentNotFound) {
			c.JSON(http.StatusNotFound, response.ResponseError{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, response.ResponseError{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[TournamentListItemResponse]{Data: tournament})
}

// ListTournaments godoc
// @Summary      List Tournaments
// @Description  Get a paginated list of tournaments with optional filters
// @Tags         Tournaments
// @Accept       json
// @Produce      json
// @Param        name            query     string  false  "Filter by tournament name"
// @Param        region_id       query     int     false  "Filter by region ID"
// @Param        country_id      query     int     false  "Filter by country ID"
// @Param        game_id         query     int     false  "Filter by game ID"
// @Param        organizer_type  query     string  false  "Filter by organizer type: OFFICIAL, COMMUNITY, INDEPENDENT"
// @Param        type_id         query     int     false  "Filter by tournament type ID"
// @Param        status          query     string  false  "Filter by time window: active, upcoming, finished"
// @Param        sort_by         query     string  false  "Sort order: start_date_asc, start_date_desc, name_asc, name_desc, newest"
// @Param        page            query     int     false  "Page number (default 1)"
// @Success      200  {object}  TournamentListSuccessResponse
// @Failure      400  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/tournaments [get]
func (m *Module) ListTournaments(c *gin.Context) {
	filter := TournamentListFilter{}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if v := c.Query("region_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid region_id"})
			return
		}
		id32 := int32(id)
		filter.RegionID = &id32
	}
	if v := c.Query("country_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid country_id"})
			return
		}
		id32 := int32(id)
		filter.CountryID = &id32
	}
	if v := c.Query("game_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid game_id"})
			return
		}
		id32 := int32(id)
		filter.GameID = &id32
	}
	if v := c.Query("organizer_type"); v != "" {
		filter.OrganizerType = &v
	}
	if v := c.Query("type_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid type_id"})
			return
		}
		id32 := int32(id)
		filter.TypeID = &id32
	}
	if v := c.Query("status"); v != "" {
		filter.Status = &v
	}
	if v := c.Query("sort_by"); v != "" {
		filter.SortBy = &v
	}

	page := 1
	if p := c.Query("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid page"})
			return
		}
		page = v
	}

	result, err := m.service.ListTournaments(c.Request.Context(), filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list tournaments: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[dto.PaginatedResponse[TournamentListItemResponse]]{
		Data: result,
	})
}
