package countries

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	countries := v1.Group("/countries")
	{
		countries.GET("/", m.GetAllCountries)
	}
}

// GetAllCountries godoc
// @Summary Get all countries
// @Description Retrieves a list of all countries
// @Tags Countries
// @Accept json
// @Produce json
// @Success 200 {array} CountryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/countries/ [get]
func (m *Module) GetAllCountries(c *gin.Context) {
	countries, err := m.service.GetAllCountries(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, countries)
}
