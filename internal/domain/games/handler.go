package games

import (
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	games := v1.Group("/games")
	{
		games.GET("", m.ListGames)
	}
}

// ListGames godoc
// @Summary      List Games
// @Description  Get a list of games, optionally filtered by name
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        name   query     string  false  "Filter by game name"
// @Success      200    {object}  response.ResponseOk[[]GameListItemResponse]
// @Failure      500    {object}  response.ResponseError
// @Router /api/v1/games [get]
func (m *Module) ListGames(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Query("name")

	games, err := m.service.ListGames(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list games: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]GameListItemResponse]{
		Data: games,
	})
}
