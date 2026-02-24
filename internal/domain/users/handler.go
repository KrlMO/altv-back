package users

import (
	"arena/pkg/middleware"
	"arena/pkg/response"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.service.Register(c, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := m.service.FastRegister(c, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			Ok:    false,
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[UserProfileResponse]{
		Ok:   true,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req == (UpdateUserProfileRequest{}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty request"})
		return
	}
	if err := m.service.UpdateProfile(c, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req == (PatchUserProfileRequest{}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty request"})
		return
	}
	if err := m.service.PatchProfile(c, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
