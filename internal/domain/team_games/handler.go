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
		auth := teamGames.Group("/", middleware.Auth(jwtSecret))
		{
			auth.POST("/:id/members", m.AddMembersToTeamGame)
			auth.DELETE("/:id/members", m.RemoveMembersFromTeamGame)
			auth.DELETE("/:id/members/roles", m.RemoveTeamGameRoleFromMember)
			auth.POST("/:id/members/roles", m.AddTeamGameRoleToMember)
			auth.POST("/:id/leave", m.LeaveTeamGame)
		}
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
// @Router /team-games/{id}/members [post]
func (m *Module) AddMembersToTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	req := AddTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}

	teamGameName, err := m.service.AddMembersToTeamGame(c, userID, globalRole, paramTeamGameId, req)
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
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "Team Game Members added successfully in: " + teamGameName,
	})
}

// LeaveTeamGame godoc
// @Summary Leave Team Game
// @Description Allows the current user to leave a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Success 200  {object}  map[string]string
// @Failure 400  {object}  response.ResponseError
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /team-games/{id}/leave [post]
func (m *Module) LeaveTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
			Ok:    false,
		})
		return
	}

	teamGameName, err := m.service.LeaveTeamGame(c, userID, globalRole, paramTeamGameId)

	if err != nil {
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
				Ok:    false,
			})
			return
		}
		if errors.Is(err, rules.ErrUserDoesNotBelongToTeamGame) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error() + ": " + teamGameName,
				Ok:    false,
			})
			return
		}
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error() + ": " + teamGameName,
				Ok:    false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Ok:   true,
		Data: map[string]string{"user has left team_game successfully": paramTeamGameId},
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
// @Router /team-games/{id}/members [delete]
func (m *Module) RemoveMembersFromTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
			Ok:    false,
		})
		return
	}
	req := RemoveTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}

	teamGameName, err := m.service.RemoveMembersFromTeamGame(c, userID, globalRole, paramTeamGameId, req)
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
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "users removed from team_game successfully in: " + teamGameName,
	})
}

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
// @Router /team-games/{id}/members/roles [delete]
func (m *Module) RemoveTeamGameRoleFromMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
			Ok:    false,
		})
		return
	}
	req := RemoveTeamGameRoleFromUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}

	teamGameName, err := m.service.RemoveTeamGameRoleFromMember(c, userID, globalRole, paramTeamGameId, req)
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
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role removed from users successfully in: " + teamGameName,
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
// @Router /team-games/{id}/members/roles [post]
func (m *Module) AddTeamGameRoleToMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
			Ok:    false,
		})
		return
	}
	req := AddTeamGameRoleToMemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
			Ok:    false,
		})
		return
	}

	teamGameName, err := m.service.AddTeamGameRoleToMember(c, userID, globalRole, paramTeamGameId, req)
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
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role added to users successfully in: " + teamGameName,
	})
}
