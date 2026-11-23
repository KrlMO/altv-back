package teams

import (
	"arena/internal/domain/teams/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddMemberToTeam godoc
// @Summary Add a member to a team
// @Description Adds a member to a team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param request body AddMemberRequest true "Member info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams/{id}/members [post]
func (m *Module) AddMemberToTeam(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	teamID := c.Param("id")
	req := AddMemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.ResponseError{
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	if err := m.service.AddMemberToTeam(c.Request.Context(), userID, teamID, globalRole, req); err != nil {
		c.JSON(500, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added to team"})
}

// UpdateTeamMemberRole godoc
// @Summary Updates team member role role
// @Description Updates team member role role
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Param request body UpdateTeamMemberRoleRequest true "Role"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams/{id}/members/{userId} [patch]
func (m *Module) UpdateTeamMemberRole(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamId := c.Param("id")
	paramUserID := c.Param("userId")

	req := UpdateTeamMemberRoleRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.ResponseError{
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	if err := m.service.UpdateUserTeamRole(c.Request.Context(), userID, globalRole, paramTeamId, paramUserID, req); err != nil {
		c.JSON(500, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated team member role"})
}

// DeleteUserFromTeam godoc
// @Summary Delete member from a team
// @Description Delete member from a team
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/teams/{id}/members/{userId} [delete]
func (m *Module) DeleteUserFromTeam(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamId := c.Param("id")
	paramUserID := c.Param("userId")

	err := m.service.DeleteUserFromTeam(c.Request.Context(), userID, globalRole, paramTeamId, paramUserID)
	if err != nil {

		if errors.Is(err, rules.ErrUserNotMember) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User " + paramUserID + " was deleted succesfuly from team: " + paramTeamId,
	})

}
