package users

import (
	"arena/internal/domain/identity"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")
	users := v1.Group("/users")
	{
		users.POST("/register/local", m.RegisterLocal)
		users.POST("/fast-register", m.FastRegister)

		auth := users.Group("/", middleware.Auth(jwtSecret))
		{
			auth.GET("/me", m.GetMe)
			auth.PUT("/me", m.UpdateMe)
			auth.PATCH("/me", m.PatchUpdateMe)
			auth.DELETE("/me", m.SelfDelete)
		}

		admin := users.Group("/", middleware.Auth(jwtSecret))
		{
			admin.DELETE("/:id", m.AdminDeleteUser)
			admin.PATCH("/:id/activate", m.AdminActivateUser)
			admin.PATCH("/:id/deactivate", m.AdminDeactivateUser)
		}
	}
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "User info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/register/local [post]
func (m *Module) RegisterLocal(c *gin.Context) {
	req := RegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	if err := m.service.Register(c, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

// FaastRegisterUser godoc
// @Summary Fast register a new user
// @Description Quickly creates a new user in the system with minimal info
// @Tags Users
// @Accept json
// @Produce json
// @Param request body FastRegisterUserRequest true "User info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/fast-register [post]
func (m *Module) FastRegister(c *gin.Context) {
	req := FastRegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}

	if err := m.service.FastRegister(c, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

// GetMe godoc
// @Summary Get current user profile
// @Description Retrieves the profile of the currently authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me [get]
func (m *Module) GetMe(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)

	user, err := m.service.GetProfile(c, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ResponseError{
			Error: "unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[UserProfileResponse]{
		Data: *user,
	})
}

// UpdateMe godoc
// @Summary Update current user profile
// @Description Updates the profile of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserProfileRequest true "Updated user info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me [put]
func (m *Module) UpdateMe(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	req := UpdateUserProfileRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}
	if req == (UpdateUserProfileRequest{}) {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "empty request"})
		return
	}
	if err := m.service.UpdateProfile(c, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})

}

// PatchUpdateMe godoc
// @Summary Partially update current user profile
// @Description Partially updates the profile of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PatchUserProfileRequest true "Partially updated user info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me [patch]
func (m *Module) PatchUpdateMe(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	req := PatchUserProfileRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}
	if req == (PatchUserProfileRequest{}) {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: "empty request"})
		return
	}
	if err := m.service.PatchProfile(c, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// SelfDelete godoc
// @Summary Delete current user account
// @Description Soft-deletes the current user account (GDPR). Anonymizes PII and revokes all tokens.
// @Tags Users
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} response.ResponseError
// @Router /api/v1/users/me [delete]
func (m *Module) SelfDelete(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)

	if err := m.service.SelfDelete(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to delete account: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// AdminDeleteUser godoc
// @Summary Delete a user (admin)
// @Description Soft-deletes a user by ID. Only ILUVATAR or AINUR roles allowed.
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
	if !identity.IsIluvatarOrAinur(roles) {
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
// @Description Activates a deactivated user by ID. Only ILUVATAR or AINUR roles allowed.
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/users/{id}/activate [patch]
func (m *Module) AdminActivateUser(c *gin.Context) {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !identity.IsIluvatarOrAinur(roles) {
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
// @Description Deactivates a user by ID. Only ILUVATAR or AINUR roles allowed.
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Router /api/v1/users/{id}/deactivate [patch]
func (m *Module) AdminDeactivateUser(c *gin.Context) {
	roles := c.MustGet(middleware.ContextRoles).([]string)
	if !identity.IsIluvatarOrAinur(roles) {
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
