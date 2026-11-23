package users

import (
	"arena/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "User info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/register/local [post]
func (m *Module) RegisterLocal(c *gin.Context) {
	req := RegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	if err := m.service.Register(c, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

// FaastRegisterUser godoc
// @Summary Fast register a new user
// @Description Quickly creates a new user in the system with minimal info
// @Tags Users
// @Accept json
// @Produce json
// @Param request body FastRegisterUserRequest true "User info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/fast-register [post]
func (m *Module) FastRegister(c *gin.Context) {
	req := FastRegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}

	if err := m.service.FastRegister(c, req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}
