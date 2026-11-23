package teamgames

import (
	"arena/pkg/middleware"
	"arena/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UploadMemberJersey godoc
// @Summary      Upload a member jersey image
// @Description  Upload the jersey photo for a team game member (multipart form: user_id + file)
// @Tags         Team Games
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string true "Team Game ID"
// @Param        user_id formData   string true "Member User ID"
// @Param        file    formData   file   true "Jersey image file"
// @Success      200     {object}   response.ResponseOk[map[string]string]
// @Failure      400     {object}   response.ResponseError
// @Router /api/v1/team-games/{id}/members/jersey [post]
func (m *Module) UploadMemberJersey(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRole := c.MustGet(middleware.ContextRoles).([]string)
	teamGameId := c.Param("id")

	memberID, err := strconv.ParseInt(c.PostForm("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "invalid user_id",
		})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "file is required",
		})
		return
	}

	jerseyURL, err := m.service.UploadMemberJersey(c.Request.Context(), userID, globalRole, teamGameId, memberID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[map[string]string]{
		Data: map[string]string{"jersey_photo": jerseyURL},
	})
}
