package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teamGames := r.Group("/team-games")
	{
		teamGames.GET("/:id/members", m.ListTeamGameMembers)
	}
}

func (m *Module) ListTeamGameMembers(c *gin.Context) {
	teamGameId := c.Param("id")
	resp, err := m.service.GetTeamGameMembers(c, teamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyRoaster) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
				Ok:    false,
			})
		}
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[PublicTeamGameMembersResponse]{
		Ok:   true,
		Data: resp,
		Msg:  "Public Team Game Members",
	})
}
