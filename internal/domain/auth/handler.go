package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login/local", m.Login)
	}
}

// Login godoc
// @Summary Login a user
// @Description Logs in a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User info"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/login/local [post]
func (m *Module) Login(c *gin.Context) {
	req := LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := m.service.LocalLogin(c, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, response)
}
