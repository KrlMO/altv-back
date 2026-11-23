package users

import (
	"arena/pkg/middleware"
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
