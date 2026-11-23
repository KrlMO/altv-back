package teamgames

import (
	"arena/internal/domain/team_games/rules"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateTeamGameWithMembers godoc
// @Summary      Create a team game (roster) with members and jerseys
// @Description  Creates a team game for a team and adds all its members with their roles in one request.
// @Description  Multipart form: game_slug (text), members (JSON text: [{"user_id":1,"role":"PLAYER"}]) and one jersey file per member named "jersey_<user_id>".
// @Tags         Team Games
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string true "Team ID"
// @Param        game_slug  formData  string true "Game slug"
// @Param        members    formData  string true "Members JSON: [{\"user_id\":1,\"role\":\"PLAYER\"}]"
// @Param        jersey_1   formData  file   false "Jersey image for user 1"
// @Success      200        {object}  response.ResponseOk[map[string]string]
// @Failure      400        {object}  response.ResponseError
// @Router /api/v1/teams/{id}/team-games [post]
func (m *Module) CreateTeamGameWithMembers(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamId := c.Param("id")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid multipart form",
		})
		return
	}

	gameSlug := ""
	if values := form.Value["game_slug"]; len(values) > 0 {
		gameSlug = values[0]
	}
	if gameSlug == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "game_slug is required",
		})
		return
	}

	var members []AddTeamGameMemberRequestItem
	if values := form.Value["members"]; len(values) > 0 {
		if err := json.Unmarshal([]byte(values[0]), &members); err != nil {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: "invalid members JSON: " + err.Error(),
			})
			return
		}
	}
	if len(members) == 0 {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "members are required",
		})
		return
	}

	jerseyFiles := make(map[int64]*multipart.FileHeader)
	for _, member := range members {
		files := form.File["jersey_"+strconv.FormatInt(member.UserID, 10)]
		if len(files) > 0 {
			jerseyFiles[member.UserID] = files[0]
		}
	}

	teamGameID, err := m.service.CreateTeamGameWithMembers(
		c.Request.Context(), userID, globalRole, paramTeamId, gameSlug, members, jerseyFiles,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"id": strconv.FormatInt(teamGameID, 10)},
		Msg:  "Team game created successfully",
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

	err := m.service.LeaveTeamGame(c, userID, globalRole, paramTeamGameId)

	if err != nil {
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		if errors.Is(err, rules.ErrUserDoesNotBelongToTeamGame) {
			c.JSON(http.StatusNotFound, response.ResponseError{
				Error: err.Error(),
			})
			return
		}
		if errors.Is(err, rules.ErrCannotLeaveTeamGame) {
			c.JSON(http.StatusBadRequest, response.ResponseError{
				Error: err.Error(),
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

// DeactivateTeamGame godoc
// @Summary Deactivate Team Game
// @Description Sets the active flag of a team game to false
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Success 200  {object}  response.ResponseOk[map[string]string]
// @Failure 400  {object}  response.ResponseError
// @Failure 404  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/deactivate [patch]
func (m *Module) DeactivateTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}

	err := m.service.DeactivateTeamGame(c, userID, globalRole, paramTeamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrBadIdTeamGame) {
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

	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game deactivated successfully",
	})
}

// ActivateTeamGame godoc
// @Summary Activate Team Game
// @Description Sets the active flag of a team game to true
// @Tags Team Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path      string  true  "Team Game ID"
// @Success 200  {object}  response.ResponseOk[map[string]string]
// @Failure 400  {object}  response.ResponseError
// @Failure 404  {object}  response.ResponseError
// @Router /api/v1/team-games/{id}/activate [patch]
func (m *Module) ActivateTeamGame(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	paramTeamGameId := c.Param("id")

	if paramTeamGameId == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "team game ID is required",
		})
		return
	}

	err := m.service.ActivateTeamGame(c, userID, globalRole, paramTeamGameId)
	if err != nil {
		if errors.Is(err, rules.ErrBadIdTeamGame) {
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

	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"team_game": paramTeamGameId},
		Msg:  "team game activated successfully",
	})
}
