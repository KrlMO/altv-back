// @title Arena API
// @version 1.0
// @description API principal de Arena Latam
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	_ "arena/docs"
	"arena/internal/config"
	"arena/internal/database"
	"arena/internal/infrastructure/seeder"
	"arena/internal/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Invalid configuration:", err)
	}
	log.Println("Server port:", cfg.ServerPort)

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	adminCfg := seeder.LoadAdminConfig()
	if err := seeder.SeedAdmin(db, adminCfg); err != nil {
		log.Fatal("Failed to seed admin user:", err)
	}

	r := router.SetupRouter(db, cfg)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
