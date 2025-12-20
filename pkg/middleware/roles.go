package middleware

import "github.com/gin-gonic/gin"

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := c.MustGet(ContextRoles).([]string)

		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
