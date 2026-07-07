package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
)

func CorsConfig(origin string) cors.Config {
	if origin == "" {
		origin = "http://localhost:3000"
	}
	return cors.Config{
		AllowOrigins:     []string{origin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}
