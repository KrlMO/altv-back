package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login/local", m.LoginLocal)
		auth.POST("/refresh", m.Refresh)
	}
}

// LoginLocal godoc
// @Summary Login a user
// @Description Logs in a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User info"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/login/local [post]
func (m *Module) LoginLocal(c *gin.Context) {

	req := LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, refreshToken, err := m.service.LocalLogin(
		c,
		req.Email,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
	})

	c.JSON(http.StatusOK, response)
}

// Refresh godoc
// @Summary Refresh user token
// @Description Refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (m *Module) Refresh(c *gin.Context) {

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "missing refresh token",
		})
		return
	}

	reponse, newRefreshToken, err := m.service.Refresh(
		c,
		cookie,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid refresh token",
		})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   false, // true en producción
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
	})

	c.JSON(http.StatusOK, reponse)
}
