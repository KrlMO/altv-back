package teams

import (
	"arena/internal/domain/teams/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateTeamGame godoc
// @Summary Add a game into a team
// @Description Add a game into a team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param request body CreateTeamGameRequest true "Team Game info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams/{id}/games [post]
func (m *Module) CreateTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamId := c.Param("id")

	req := CreateTeamGameRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "Invalid Request" + err.Error(),
		})
		return
	}
	teamGameId, err := m.service.AddTeamGame(c.Request.Context(), paramTeamId, userID, globalRole, req)
	if err != nil {
		if errors.Is(err, rules.ErrUserNotMember) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      teamGameId,
		"message": "Team Game for game: " + req.GameSlug + " has been added succesfuly",
	})
}
