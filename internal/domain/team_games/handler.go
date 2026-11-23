package teamgames

import (
	"arena/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teamGames := r.Group("/api/v1/team-games")
	{
		teamGames.GET("", m.ListTeamGames)
		teamGames.GET("/:id/members", m.ListTeamGameMembers)
		auth := teamGames.Group("/", middleware.Auth(jwtSecret))
		{
			auth.GET("/:id/members/manage", m.ListTeamGameManageMembers)
			auth.POST("/:id/members", m.AddMembersToTeamGame)
			auth.DELETE("/:id/members", m.RemoveMembersFromTeamGame)
			auth.DELETE("/:id/members/roles", m.RemoveTeamGameRoleFromMember)
			auth.POST("/:id/members/roles", m.AddTeamGameRoleToMember)
			auth.POST("/:id/members/jersey", m.UploadMemberJersey)
			auth.POST("/:id/leave", m.LeaveTeamGame)
			auth.PATCH("/:id/deactivate", m.DeactivateTeamGame)
			auth.PATCH("/:id/activate", m.ActivateTeamGame)
		}
	}

	teamsTeamGames := r.Group("/api/v1/teams/:id/team-games")
	{
		teamsTeamGames.GET("", m.ListTeamGamesByTeam)
	}

	teamsTeamGamesAuth := r.Group("/api/v1/teams/:id/team-games", middleware.Auth(jwtSecret))
	{
		teamsTeamGamesAuth.POST("", m.CreateTeamGameWithMembers)
	}
}
