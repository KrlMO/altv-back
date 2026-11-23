package organizations

import (
	"arena/internal/domain/identity"
	"arena/pkg/middleware"
	"arena/pkg/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(r *gin.Engine, jwtSecret string) {
	v1 := r.Group("/api/v1")
	organizations := v1.Group("/organizations")
	{
		auth := organizations.Group("", middleware.Auth(jwtSecret))
		{
			auth.GET("/my", m.GetMyOrganizations)
		}
	}
}

// GetMyOrganizations godoc
// @Summary List the organizations of the authenticated user
// @Description Retrieves every organization the authenticated user belongs to, grouped by the organization role the user holds in each one. Each organization also reports the permission codes granted by its currently valid plans. BANNED users are rejected.
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.ResponseOk[[]OrganizationRoleGroupDto]
// @Failure 401 {object} response.ResponseError
// @Failure 403 {object} response.ResponseError
// @Failure 500 {object} response.ResponseError
// @Router /api/v1/organizations/my [get]
func (m *Module) GetMyOrganizations(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserID).(int64)
	globalRoles := c.MustGet(middleware.ContextRoles).([]string)

	organizations, err := m.service.GetMyOrganizations(c.Request.Context(), userID, globalRoles)
	if err != nil {
		if errors.Is(err, identity.ErrBannedFromOrganizations) {
			c.JSON(http.StatusForbidden, response.ResponseError{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.ResponseError{
			Error: "Internal error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ResponseOk[[]OrganizationRoleGroupDto]{
		Data: organizations,
	})
}
