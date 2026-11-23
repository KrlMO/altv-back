package teams

import (
	"arena/pkg/dto"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Register godoc
// @Summary Create a new team
// @Description Create a new team
// @Tags Teams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Team name"
// @Param tag formData string true "Team tag"
// @Param description formData string false "Team description"
// @Param country_id formData int true "Country ID"
// @Param logo formData file false "Team logo image"
// @Param banner formData file false "Team banner image"
// @Success 201 {object} response.ResponseOk[TeamResponse]
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams [post]
func (m *Module) Register(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRoles := c.MustGet(middleware.ContextRoles).([]string)
	req := CreateTeamRequestDto{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, response.ResponseError{
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	team, err := m.service.RegisterTeam(c.Request.Context(), userID, globalRoles, &req)
	if err != nil {
		c.JSON(400, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(201, response.ResponseOk[TeamResponse]{
		Data: *team,
	})
}

func (m *Module) GetMyTeams(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRoles := c.MustGet(middleware.ContextRoles).([]string)

	teams, err := m.service.GetMyTeams(c.Request.Context(), userID, globalRoles)
	if err != nil {
		c.JSON(500, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(200, response.ResponseOk[[]MyTeamListItemDto]{
		Data: teams,
	})
}

// GetTeamByID godoc
// @Summary Get team by ID
// @Description Retrieves a team by its ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Failure 404 {object} map[string]string
// @Router /api/v1/teams/{id} [get]
func (m *Module) GetTeamByID(c *gin.Context) {
	teamID := c.Param("id")
	team, err := m.service.GetTeamByID(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(404, response.ResponseError{
			Error: "Team not found: " + err.Error(),
		})
		return
	}
	response := response.ResponseOk[TeamResponse]{
		Data: *team,
	}
	c.JSON(200, response)
}

// GetTeams godoc
// @Summary Get list of teams
// @Description Retrieves a list of teams with optional filters
// @Tags Teams
// @Accept json
// @Produce json
// @Param country_id query int false "Filter by country ID"
// @Param owner_id query int false "Filter by owner ID"
// @Param active query bool false "Filter by active status"
// @Param name_like query string false "Filter by name substring"
// @Param tag_like query string false "Filter by tag substring"
// @Param limit query int false "Limit number of results" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams [get]
func (m *Module) GetTeams(c *gin.Context) {
	var filter TeamFilter

	if v := c.Query("country_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(400, response.ResponseError{
				Error: "invalid country_id",
			})
			return
		}
		id32 := int32(id)
		filter.CountryID = &id32
	}

	if v := c.Query("owner_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(400, response.ResponseError{
				Error: "invalid owner_id",
			})
			return
		}
		filter.OwnerID = &id
	}

	if v := c.Query("active"); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(400, response.ResponseError{
				Error: "invalid active",
			})
			return
		}
		filter.Active = &val
	}

	if v := c.Query("name_like"); v != "" {
		filter.NameLike = &v
	}

	if v := c.Query("tag_like"); v != "" {
		filter.TagLike = &v
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}
	filter.Limit = limit

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	filter.Offset = offset

	paginatedTeams, err := m.service.GetTeams(c.Request.Context(), filter)
	if err != nil {
		c.JSON(500, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	response := response.ResponseOk[dto.PaginatedResponse[TeamResponse]]{
		Data: paginatedTeams,
	}
	c.JSON(200, response)
}

// PatchUpdateTeam godoc
// @Summary Update team details
// @Description Updates the details of an existing team
// @Tags Teams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param name formData string false "Team name"
// @Param tag formData string false "Team tag"
// @Param description formData string false "Team description"
// @Param country_id formData int false "Country ID"
// @Param logo formData file false "Team logo image"
// @Param banner formData file false "Team banner image"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams/{id} [patch]
func (m *Module) PatchUpdateTeam(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRoles := c.MustGet(middleware.ContextRoles).([]string)
	teamID := c.Param("id")
	req := PatchUpdateTeamRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, response.ResponseError{
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	if err := m.service.PatchUpdateTeam(c.Request.Context(), userID, globalRoles, teamID, req); err != nil {
		c.JSON(400, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"message": "team updated"})
}
