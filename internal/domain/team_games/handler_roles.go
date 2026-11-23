package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RemoveTeamGameRoleFromMember godoc
// @Summary Remove Team Game Role from Member
// @Description Remove a specific role from members in a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body RemoveTeamGameRoleFromUserRequest true "Remove Team Game Role From User Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members/roles [delete]
func (m *Module) RemoveTeamGameRoleFromMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := RemoveTeamGameRoleFromUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	err := m.service.RemoveTeamGameRoleFromMember(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role removed from users successfully",
	})
}

// AddTeamGameRoleToMember godoc
// @Summary Add Team Game Role to Member
// @Description Add a specific role to members in a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body AddTeamGameRoleToMemberRequest true "Add Team Game Role To Member Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members/roles [post]
func (m *Module) AddTeamGameRoleToMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := AddTeamGameRoleToMemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	err := m.service.AddTeamGameRoleToMember(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role added to users successfully",
	})
}
