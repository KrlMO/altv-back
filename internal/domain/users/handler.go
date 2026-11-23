package users

import (
	"arena/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")
	users := v1.Group("/users")
	{
		users.POST("/register/local", m.RegisterLocal)
		users.POST("/fast-register", m.FastRegister)
		users.GET("/search", m.SearchUsers)

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
