package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teamGames := r.Group("/team-games")
	{
		teamGames.GET("/:id/members", m.ListTeamGameMembers)
		teamGames.POST("/:id/members", m.AddMembersToTeamGame)
	}
}

// ListTeamGameMembers godoc
// @Summary      List Team Game Members
// @Description  Get public members of a team game by team game ID
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team Game ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  response.ResponseError
// @Failure      404  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router       /team-games/{id}/members [get]
func (m *Module) ListTeamGameMembers(c *gin.Context) {
	teamGameId := c.Param("id")
	resp, err := m.service.GetTeamGameMembers(c, teamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyRoaster) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
				Ok:    false,
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[PublicTeamGameMembersResponse]{
		Ok:   true,
		Data: resp,
		Msg:  "Public Team Game Members from Team Game ID fetched successfully: " + teamGameId,
	})
}

// AddMembersToTeamGame godoc
// @Summary      Add Members to Team Game
// @Description  Add members to a team game
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team Game ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  response.ResponseError
// @Failure      404  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router       /team-games/{id}/members [post]
func (m *Module) AddMembersToTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	req := AddTeamGameMemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}

	err := m.service.AddMembersToTeamGame(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
				Ok:    false,
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Ok:   true,
		Data: map[string]string{"team_name": paramTeamGameId},
	})
}
