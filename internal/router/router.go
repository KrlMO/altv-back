package router

import (
	_ "arena/docs"
	"arena/internal/config"
	"arena/internal/domain/users"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	usersModule := users.NewModule(db, cfg.JWTSecret)
	usersModule.RegisterRoutes(r)

	return r
}
