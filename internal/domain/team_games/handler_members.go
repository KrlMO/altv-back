package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
// @Router /api/v1/team-games/{id}/members [get]
func (m *Module) ListTeamGameMembers(c *gin.Context) {
	teamGameId := c.Param("id")
	resp, err := m.service.GetTeamGameMembers(c, teamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyRoaster) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[PublicTeamGameMembersResponse]{
		Data: resp,
		Msg:  "Public Team Game Members from Team Game ID fetched successfully: " + teamGameId,
	})
}

// ListTeamGameManageMembers godoc
// @Summary      List Team Game Manage Members
// @Description  Get all members of a team game with roles and jersey info for management
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Team Game ID"
// @Success      200  {object}  response.ResponseOk[[]TeamGameManageMemberResponse]
// @Failure      400  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members/manage [get]
func (m *Module) ListTeamGameManageMembers(c *gin.Context) {
	teamGameId := c.Param("id")
	resp, err := m.service.GetTeamGameManageMembers(c.Request.Context(), teamGameId)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[[]TeamGameManageMemberResponse]{
		Data: resp,
	})
}

// AddMembersToTeamGame godoc
// @Summary Add Members to Team Game
// @Description Add members to a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body AddTeamGameMembersRequest true "Add Team Game Members Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members [post]
func (m *Module) AddMembersToTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	req := AddTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	err := m.service.AddMembersToTeamGame(c, userID, globalRole, paramTeamGameId, req)
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
		Msg:  "Team Game Members added successfully",
	})
}

// RemoveMembersFromTeamGame godoc
// @Summary Remove Members from Team Game
// @Description Remove members from a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body RemoveTeamGameMembersRequest true "Remove Team Game Members Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members [delete]
func (m *Module) RemoveMembersFromTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := RemoveTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	err := m.service.RemoveMembersFromTeamGame(c, userID, globalRole, paramTeamGameId, req)
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
		Msg:  "users removed from team_game successfully",
	})
}
