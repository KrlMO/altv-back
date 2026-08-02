package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/dto"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	teamGames := r.Group("/api/v1/team-games")
	{
		teamGames.GET("", m.ListTeamGames)
		teamGames.GET("/:id/members", m.ListTeamGameMembers)
		auth := teamGames.Group("/", middleware.Auth(jwtSecret))
		{
			auth.POST("/:id/members", m.AddMembersToTeamGame)
			auth.DELETE("/:id/members", m.RemoveMembersFromTeamGame)
			auth.DELETE("/:id/members/roles", m.RemoveTeamGameRoleFromMember)
			auth.POST("/:id/members/roles", m.AddTeamGameRoleToMember)
			auth.POST("/:id/leave", m.LeaveTeamGame)
		}
	}

	teamsTeamGames := r.Group("/api/v1/teams/:id/team-games")
	{
		teamsTeamGames.GET("", m.ListTeamGamesByTeam)
	}
}

// ListTeamGames godoc
// @Summary      List Team Games
// @Description  Get a paginated list of team games with optional filters
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        name       query     string  false  "Filter by team game name"
// @Param        region_id  query     int     false  "Filter by region ID"
// @Param        country_id query     int     false  "Filter by country ID"
// @Param        game_id    query     int     false  "Filter by game ID"
// @Param        sort_by    query     string  false  "Sort order: name_asc, name_desc, newest, oldest"
// @Param        page       query     int     false  "Page number (default 1)"
// @Success      200  {object}  TeamGameListSuccessResponse
// @Failure      400  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/team-games [get]
func (m *Module) ListTeamGames(c *gin.Context) {
	filter := TeamGameListFilter{}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if regionID := c.Query("region_id"); regionID != "" {
		v, err := strconv.ParseInt(regionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid region_id"})
			return
		}
		v32 := int32(v)
		filter.RegionID = &v32
	}
	if countryID := c.Query("country_id"); countryID != "" {
		v, err := strconv.ParseInt(countryID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid country_id"})
			return
		}
		v32 := int32(v)
		filter.CountryID = &v32
	}
	if gameID := c.Query("game_id"); gameID != "" {
		v, err := strconv.ParseInt(gameID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid game_id"})
			return
		}
		v32 := int32(v)
		filter.GameID = &v32
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = &sortBy
	}

	page := 1
	if p := c.Query("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, response.ResponseError{Error: "invalid page"})
			return
		}
		page = v
	}

	result, err := m.service.ListTeamGames(c.Request.Context(), filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list team games: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[dto.PaginatedResponse[TeamGameListItemResponse]]{
		Data: result,
	})
}

// ListTeamGamesByTeam godoc
// @Summary      List Team Games by Team
// @Description  Get the list of team games belonging to a team
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team ID"
// @Success      200  {object}  response.ResponseOk[[]TeamGameListItemResponse]
// @Failure      400  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/teams/{id}/team-games [get]
func (m *Module) ListTeamGamesByTeam(c *gin.Context) {
	teamId := c.Param("id")

	resp, err := m.service.ListTeamGamesByTeam(c.Request.Context(), teamId)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]TeamGameListItemResponse]{
		Data: resp,
	})
}

// ListTeamGameMembers godoc
// @Summary      List Team Game Members
// @Description  Get public members of a team game by team game ID
// @Tags         Team Games
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team Game ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  response.ResponseError
// @Failure      404  {object}  response.ResponseError
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members [get]
func (m *Module) ListTeamGameMembers(c *gin.Context) {
	teamGameId := c.Param("id")
	resp, err := m.service.GetTeamGameMembers(c, teamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyRoaster) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[PublicTeamGameMembersResponse]{
		Data: resp,
		Msg:  "Public Team Game Members from Team Game ID fetched successfully: " + teamGameId,
	})
}

// AddMembersToTeamGame godoc
// @Summary Add Members to Team Game
// @Description Add members to a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body AddTeamGameMembersRequest true "Add Team Game Members Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members [post]
func (m *Module) AddMembersToTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	req := AddTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	teamGameName, err := m.service.AddMembersToTeamGame(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "Team Game Members added successfully in: " + teamGameName,
	})
}

// LeaveTeamGame godoc
// @Summary Leave Team Game
// @Description Allows the current user to leave a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Success 200  {object}  map[string]string
// @Failure 400  {object}  response.ResponseError
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/leave [post]
func (m *Module) LeaveTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}

	teamGameName, err := m.service.LeaveTeamGame(c, userID, globalRole, paramTeamGameId)

	if err != nil {
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		if errors.Is(err, rules.ErrUserDoesNotBelongToTeamGame) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error() + ": " + teamGameName,
			})
			return
		}
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error() + ": " + teamGameName,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"user has left team_game successfully": paramTeamGameId},
	})
}

// RemoveMembersFromTeamGame godoc
// @Summary Remove Members from Team Game
// @Description Remove members from a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body RemoveTeamGameMembersRequest true "Remove Team Game Members Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members [delete]
func (m *Module) RemoveMembersFromTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := RemoveTeamGameMembersRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	teamGameName, err := m.service.RemoveMembersFromTeamGame(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "users removed from team_game successfully in: " + teamGameName,
	})
}

// RemoveTeamGameRoleFromMember godoc
// @Summary Remove Team Game Role from Member
// @Description Remove a specific role from members in a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body RemoveTeamGameRoleFromUserRequest true "Remove Team Game Role From User Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members/roles [delete]
func (m *Module) RemoveTeamGameRoleFromMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := RemoveTeamGameRoleFromUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	teamGameName, err := m.service.RemoveTeamGameRoleFromMember(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role removed from users successfully in: " + teamGameName,
	})
}

// AddTeamGameRoleToMember godoc
// @Summary Add Team Game Role to Member
// @Description Add a specific role to members in a team game
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Param  request body AddTeamGameRoleToMemberRequest true "Add Team Game Role To Member Request"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  response.ResponseError
// @Failure 500  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/members/roles [post]
func (m *Module) AddTeamGameRoleToMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")
	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}
	req := AddTeamGameRoleToMemberRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	teamGameName, err := m.service.AddTeamGameRoleToMember(c, userID, globalRole, paramTeamGameId, req)
	if err != nil {
		if errors.Is(err, rules.ErrEmptyMembers) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game role added to users successfully in: " + teamGameName,
	})
}
