package regions

import (
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/v1/regions", m.GetAllRegions)
}

// GetAllRegions godoc
// @Summary      Get all regions
// @Description  Retrieves a list of all registered regions
// @Tags         Regions
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.ResponseOk[[]RegionResponse]
// @Failure      500  {object}  response.ResponseError
// @Router /api/v1/regions [get]
func (m *Module) GetAllRegions(c *gin.Context) {
	regions, err := m.service.GetAllRegions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "failed to list regions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]RegionResponse]{
		Data: regions,
	})
}
