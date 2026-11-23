package teams

import (
	"arena/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")
	teams := v1.Group("/teams")
	{
		teams.GET("/:id", m.GetTeamByID)
		teams.GET("", m.GetTeams)
		auth := teams.Group("", middleware.Auth(jwtSecret))
		{
			auth.POST("", m.Register)
			auth.PATCH("/:id", m.PatchUpdateTeam)
			auth.POST("/:id/members", m.AddMemberToTeam)
			auth.PATCH("/:id/members/:userId", m.UpdateTeamMemberRole)
			auth.DELETE("/:id/members/:userId", m.DeleteUserFromTeam)
			auth.POST("/:id/games", m.CreateTeamGame)
			auth.GET("/my", m.GetMyTeams)
		}
	}
}
