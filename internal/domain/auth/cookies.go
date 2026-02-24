package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   false, // true en producción
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 días
	})
}

func clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
