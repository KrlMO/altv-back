package teams

import (
	"arena/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teams := r.Group("/teams")
	{
		auth := teams.Group("/", middleware.Auth(jwtSecret))
		{
			auth.POST("/register", m.Register)
		}
	}
}

func (m *Module) Register(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	req := RegisterTeamRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := m.service.RegisterTeam(c, userID, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "team registered"})
}
