package teams

import (
	"arena/pkg/dto"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teams := r.Group("/teams")
	{
		teams.GET("/:id", m.GetTeamByID)
		teams.GET("/", m.GetTeams)
		auth := teams.Group("/", middleware.Auth(jwtSecret))
		{
			auth.POST("/", m.Register)
		}
	}
}

// RegisterTeam godoc
// @Summary Register a new team
// @Description Creates a new team in the system
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RegisterTeamRequest true "Team info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /teams [post]
func (m *Module) Register(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	req := RegisterTeamRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := m.service.RegisterTeam(c, userID, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "team registered"})
}

// GetTeamByID godoc
// @Summary Get team by ID
// @Description Retrieves a team by its ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} TeamResponse
// @Failure 404 {object} map[string]string
// @Router /teams/{id} [get]
func (m *Module) GetTeamByID(c *gin.Context) {
	teamID := c.Param("id")
	team, err := m.service.GetTeamByID(c, teamID)
	if err != nil {
		c.JSON(404, gin.H{"error": "team not found"})
		return
	}
	c.JSON(200, team)
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
// @Success 200 {array} TeamResponse
// @Failure 400 {object} map[string]string
// @Router /teams [get]
func (m *Module) GetTeams(c *gin.Context) {
	var filter TeamFilter

	if v := c.Query("country_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid country_id"})
			return
		}
		id32 := int32(id)
		filter.CountryID = &id32
	}

	if v := c.Query("owner_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid owner_id"})
			return
		}
		filter.OwnerID = &id
	}

	if v := c.Query("active"); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid active"})
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

	paginatedTeams, err := m.service.GetTeams(c, filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	response := response.ResponseOk[dto.PaginatedResponse[TeamResponse]]{
		Data: paginatedTeams,
		Ok:   true,
	}
	c.JSON(200, response)
}
