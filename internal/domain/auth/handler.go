package auth

import (
	"arena/internal/domain/auth/rules"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login/local", m.LoginLocal)
		auth.POST("/refresh", m.Refresh)
		auth.POST("/logout", m.Logout)
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

	responseData, refreshToken, err := m.service.LocalLogin(
		c,
		req.Email,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		if errors.Is(err, rules.ErrUserNotFound) || errors.Is(err, rules.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, response.ResponseError{
				Error: "invalid credentials",
				Ok:    false,
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToGenerateAccessToken) || errors.Is(err, rules.ErrFailedToGenerateRefreshToken) {
			c.JSON(http.StatusInternalServerError, response.ResponseError{
				Error: "failed to generate tokens",
				Ok:    false,
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToStoreRefreshToken) {
			c.JSON(http.StatusInternalServerError, response.ResponseError{
				Error: "failed to store refresh token",
				Ok:    false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "internal server error",
			Ok:    false,
		})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	c.JSON(http.StatusOK, response.ResponseOk[LoginResponse]{
		Ok:   true,
		Data: responseData,
	})
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
		if errors.Is(err, rules.ErrInvalidRefreshToken) || errors.Is(err, rules.ErrRefreshTokenReuseDetected) || errors.Is(err, rules.ErrRefreshTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, rules.ErrFailedToGenerateAccessToken) || errors.Is(err, rules.ErrFailedToGenerateRefreshToken) || errors.Is(err, rules.ErrFailedToStoreRefreshToken) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to refresh tokens",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   false, // true en producción
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	c.JSON(http.StatusOK, reponse)
}

// Logout godoc
// @Summary Logout a user
// @Description Logs out a user by revoking their refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (m *Module) Logout(c *gin.Context) {

	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		_ = m.service.Logout(c, cookie)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, response.ResponseOk[any]{
		Ok:  true,
		Msg: "logged out successfully",
	})
}
