package teamgames

import (
	"arena/pkg/dto"
	"arena/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListTeamGames godoc
// @Summary      List Team Games
// @Description  Get a paginated list of team games with optional filters
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        name       query     string  false  "Filter by team game name"
// @Param        region_id  query     int     false  "Filter by region ID"
// @Param        country_id query     int     false  "Filter by country ID"
// @Param        game_id    query     int     false  "Filter by game ID"
// @Param        sort_by    query     string  false  "Sort order: name_asc, name_desc, newest, oldest"
// @Param        page       query     int     false  "Page number (default 1)"
// @Success      200  {object}  TeamGameListSuccessResponse
// @Failure      400  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/team-games [get]
func (m *Module) ListTeamGames(c *gin.Context) {
	filter := TeamGameListFilter{}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if regionID := c.Query("region_id"); regionID != "" {
		v, err := strconv.ParseInt(regionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid region_id"})
			return
		}
		v32 := int32(v)
		filter.RegionID = &v32
	}
	if countryID := c.Query("country_id"); countryID != "" {
		v, err := strconv.ParseInt(countryID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid country_id"})
			return
		}
		v32 := int32(v)
		filter.CountryID = &v32
	}
	if gameID := c.Query("game_id"); gameID != "" {
		v, err := strconv.ParseInt(gameID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid game_id"})
			return
		}
		v32 := int32(v)
		filter.GameID = &v32
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = &sortBy
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

	result, err := m.service.ListTeamGames(c.Request.Context(), filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list team games: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[dto.PaginatedResponse[TeamGameListItemResponse]]{
		Data: result,
	})
}

// ListTeamGamesByTeam godoc
// @Summary      List Team Games by Team
// @Description  Get the list of team games belonging to a team
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team ID"
// @Success      200  {object}  response.ResponseOk[[]TeamGameListItemResponse]
// @Failure      400  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/teams/{id}/team-games [get]
func (m *Module) ListTeamGamesByTeam(c *gin.Context) {
	teamId := c.Param("id")

	resp, err := m.service.ListTeamGamesByTeam(c.Request.Context(), teamId)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]TeamGameListItemResponse]{
		Data: resp,
	})
}
