package tournaments

import (
	"arena/pkg/dto"
	"arena/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/v1/tournaments", m.ListTournaments)
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
