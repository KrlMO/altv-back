package stages

import (
	"net/http"

	"arena/pkg/response"

	"github.com/gin-gonic/gin"
)

type StageFormatDTO struct {
	ID          int32   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// ListStageFormats godoc
// @Summary List the available stage formats
// @Description Returns the active stage formats of the catalog, for stage creation forms.
// @Tags TournamentStages
// @Produce json
// @Success 200 {object} response.ResponseOk[[]StageFormatDTO]
// @Router /api/v1/stage-formats [get]
func (m *Module) ListStageFormats(c *gin.Context) {
	formats, err := m.service.ListStageFormats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{Error: "internal error"})
		return
	}
	c.JSON(http.StatusOK, response.ResponseOk[[]StageFormatDTO]{Data: formats})
}
