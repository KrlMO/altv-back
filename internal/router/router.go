package router

import (
	"arena/internal/config"
	"arena/internal/domain/auth"
	"arena/internal/domain/countries"
	"arena/internal/domain/games"
	"arena/internal/domain/regions"
	teamgames "arena/internal/domain/team_games"
	"arena/internal/domain/teams"
	tournamenttypes "arena/internal/domain/tournament_types"
	"arena/internal/domain/tournaments"
	"arena/internal/domain/users"
	"arena/internal/infrastructure/mail"
	"arena/internal/shared"
	"arena/pkg/middleware"
	"arena/pkg/storage"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(middleware.CorsConfig(cfg.CORSOrigin)))

	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// LOCAL MEDIA
	if cfg.StorageDriver == storage.DriverLocal && cfg.MediaRoot != "" {
		r.Static("/media", cfg.MediaRoot)
	}

	// SHARED
	sharedRepo := shared.NewSharedRepository(db)

	// USERS
	usersRepo := users.NewRepository(db)
	usersService := users.NewService(usersRepo, sharedRepo)
	usersModule := users.NewModule(usersService)
	usersModule.RegisterRoutes(r, cfg.JWTSecret)

	//TEAMS
	teamsRepo := teams.NewRepository(db)
	mediaStorage, err := storage.New(cfg.StorageDriver, cfg.MediaURL, cfg.MediaRoot)
	if err != nil {
		panic("invalid storage configuration: " + err.Error())
	}
	teamsService := teams.NewService(teamsRepo, sharedRepo, mediaStorage)
	teamsModule := teams.NewModule(teamsService)
	teamsModule.RegisterRoutes(r, cfg.JWTSecret)

	//TEAM GAMES
	teamGamesRepo := teamgames.NewRepository(db, sharedRepo)
	teamGamesService := teamgames.NewService(teamGamesRepo, mediaStorage)
	teamGamesModule := teamgames.NewModule(teamGamesService)
	teamGamesModule.RegisterRoutes(r, cfg.JWTSecret)

	// COUNTRIES
	countriesRepo := countries.NewRepository(db)
	countriesService := countries.NewService(countriesRepo)
	countriesModule := countries.NewModule(countriesService)
	countriesModule.RegisterRoutes(r)

	// GAMES
	gamesRepo := games.NewRepository(db)
	gamesService := games.NewService(gamesRepo)
	gamesModule := games.NewModule(gamesService)
	gamesModule.RegisterRoutes(r)

	// REGIONS
	regionsRepo := regions.NewRepository(db)
	regionsService := regions.NewService(regionsRepo)
	regionsModule := regions.NewModule(regionsService)
	regionsModule.RegisterRoutes(r)

	// TOURNAMENT TYPES
	tournamentTypesRepo := tournamenttypes.NewRepository(db)
	tournamentTypesService := tournamenttypes.NewService(tournamentTypesRepo)
	tournamentTypesModule := tournamenttypes.NewModule(tournamentTypesService)
	tournamentTypesModule.RegisterRoutes(r)

	// TOURNAMENTS
	tournamentsRepo := tournaments.NewRepository(db)
	tournamentsService := tournaments.NewService(tournamentsRepo)
	tournamentsModule := tournaments.NewModule(tournamentsService)
	tournamentsModule.RegisterRoutes(r)

	mailer := mail.NewSMTPMailer(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		cfg.SMTPFrom,
		cfg.BaseURL,
	)
	// AUTH
	authRepo := auth.NewRepository(db)
	authModule := auth.NewModule(
		usersService,
		authRepo,
		mailer,
		cfg.JWTSecret,
		cfg.FrontendURL,
		15*time.Minute,
		7*24*time.Hour,
	)
	authModule.RegisterRoutes(r)

	return r
}
