package users

import (
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchUsers godoc
// @Summary Search users
// @Description Searches for users based on a query string. Returns a list of matching users.
// @Tags Users
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} response.ResponseOk[[]UserSearchResult]
// @Failure 400 {object} response.ResponseError
// @Failure 500 {object} response.ResponseError
// @Router /api/v1/users/search [get]
func (m *Module) SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: "query parameter 'q' is required",
		})
		return
	}

	users, err := m.service.SearchUsers(c, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to search users: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]UserSearchResult]{
		Data: users,
	})
}
