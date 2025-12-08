package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	service *Service
}

func NewModule(db *pgxpool.Pool, jwtSecret string) *Module {
	repo := NewRepository(db)
	service := NewService(repo, jwtSecret)
	return &Module{service}
}

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.POST("/auth/register", m.Register)

}

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user in the system
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body users.RegisterUserRequest true "User info"
// @Success 201 {object} users.UserResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (m *Module) Register(c *gin.Context) {
	req := RegisterUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		NickName:     req.NickName,
		PasswordHash: req.Password,
		CountryId:    req.CountryId,
		BirthDate:    req.BirthDate,
	}

	if err := m.service.Register(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (m *Module) Login(c *gin.Context) {
	req := LoginRequest{}
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := m.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
