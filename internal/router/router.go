package router

import (
	"arena/internal/config"
	"arena/internal/domain/auth"
	"arena/internal/domain/countries"
	teamgames "arena/internal/domain/team_games"
	"arena/internal/domain/teams"
	"arena/internal/domain/users"
	"arena/pkg/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(middleware.CorsConfig()))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// USERS
	usersRepo := users.NewRepository(db)
	usersService := users.NewService(usersRepo)
	usersModule := users.NewModule(usersService)
	usersModule.RegisterRoutes(r, cfg.JWTSecret)

	//TEAMS
	teamsRepo := teams.NewRepository(db)
	teamsService := teams.NewService(teamsRepo)
	teamsModule := teams.NewModule(teamsService)
	teamsModule.RegisterRoutes(r, cfg.JWTSecret)

	//TEAM GAMES
	teamGamesRepo := teamgames.NewRepository(db)
	teamGamesService := teamgames.NewService(teamGamesRepo)
	teamGamesModule := teamgames.NewModule(teamGamesService)
	teamGamesModule.RegisterRoutes(r, cfg.JWTSecret)

	// COUNTRIES
	countriesRepo := countries.NewRepository(db)
	countriesService := countries.NewService(countriesRepo)
	countriesModule := countries.NewModule(countriesService)
	countriesModule.RegisterRoutes(r)

	// AUTH
	authRepo := auth.NewRepository(db)
	authModule := auth.NewModule(
		usersService,
		authRepo,
		cfg.JWTSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
	authModule.RegisterRoutes(r)

	return r
}
