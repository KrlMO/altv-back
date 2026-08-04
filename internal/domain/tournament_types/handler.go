package tournamenttypes

import (
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/v1/tournament-types", m.ListTournamentTypes)
}

// ListTournamentTypes godoc
// @Summary      List Tournament Types
// @Description  Get the list of active tournament types
// @Tags         Tournament Types
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.ResponseOk[[]TournamentTypeListItemResponse]
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/tournament-types [get]
func (m *Module) ListTournamentTypes(c *gin.Context) {
	types, err := m.service.ListTournamentTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list tournament types: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]TournamentTypeListItemResponse]{
		Data: types,
	})
}
