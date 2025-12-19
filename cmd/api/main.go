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
	"arena/internal/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()
	log.Println("Server port:", cfg.ServerPort)

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	r := router.SetupRouter(db, cfg)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
