package main

import (
	"log"

	"coffee-shop-pos/configs"
	"coffee-shop-pos/internal/delivery/http"
	"coffee-shop-pos/internal/delivery/http/handler"
	"coffee-shop-pos/internal/repository/postgres"
	"coffee-shop-pos/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Initialize database connection
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Repository
	menuRepo := postgres.NewMenuItemRepository(db)

	// Initialize Usecase
	menuUsecase := usecase.NewMenuUsecase(menuRepo)

	// Initialize Handler
	menuHandler := handler.NewMenuHandler(menuUsecase)

	// Initialize Gin Engine
	r := gin.Default()

	// Setup Router
	http.NewRouter(r, menuHandler)

	// Start Server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
