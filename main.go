package main

import (
	"log"

	"bodda/internal/config"
	"bodda/internal/database"
	"bodda/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Start server
	srv := server.New(cfg, db)
	log.Printf("Server starting on port %s", cfg.Port)
	if err := srv.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}