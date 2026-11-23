package users

import (
	"arena/internal/domain/identity"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminDeleteUser godoc
// @Summary Delete a user (admin)
// @Description Soft-deletes a user by ID. Only SUPER_ADMIN or STAFF roles allowed.
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 404 {object} response.ResponseError
// @Router /api/v1/users/{id} [delete]
func (m *Module) AdminDeleteUser(c *gin.Context) {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !identity.IsStaffOrSuperAdmin(roles) {
		c.JSON(http.StatusForbidden, response.ResponseError{
			Error: "only administrators can delete users",
		})
		return
	}

	idStr := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid user ID",
		})
		return
	}

	if err := m.service.SoftDeleteUser(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to delete user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// AdminActivateUser godoc
// @Summary Activate a user (admin)
// @Description Activates a deactivated user by ID. Only SUPER_ADMIN or STAFF roles allowed.
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/users/{id}/activate [patch]
func (m *Module) AdminActivateUser(c *gin.Context) {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !identity.IsStaffOrSuperAdmin(roles) {
		c.JSON(http.StatusForbidden, response.ResponseError{
			Error: "only administrators can activate users",
		})
		return
	}

	idStr := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid user ID",
		})
		return
	}

	if err := m.service.ActivateUser(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to activate user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user activated successfully"})
}

// AdminDeactivateUser godoc
// @Summary Deactivate a user (admin)
// @Description Deactivates a user by ID. Only SUPER_ADMIN or STAFF roles allowed.
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/users/{id}/deactivate [patch]
func (m *Module) AdminDeactivateUser(c *gin.Context) {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !identity.IsStaffOrSuperAdmin(roles) {
		c.JSON(http.StatusForbidden, response.ResponseError{
			Error: "only administrators can deactivate users",
		})
		return
	}

	idStr := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid user ID",
		})
		return
	}

	if err := m.service.DeactivateUser(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to deactivate user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}
