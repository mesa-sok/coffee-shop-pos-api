package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coffee-shop-pos/configs"
	httpdelivery "coffee-shop-pos/internal/delivery/http"
	"coffee-shop-pos/internal/delivery/http/handler"
	"coffee-shop-pos/internal/repository/postgres"
	"coffee-shop-pos/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Set Gin mode before creating the engine so debug route printing is suppressed
	// in production (defaults to "release").
	gin.SetMode(cfg.GinMode)

	// Initialize database connection
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Repository
	menuRepo := postgres.NewMenuItemRepository(db)
	orderRepo := postgres.NewOrderRepository(db)

	// Initialize Usecase
	menuUsecase := usecase.NewMenuUsecase(menuRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, menuRepo)

	// Initialize Handler
	menuHandler := handler.NewMenuHandler(menuUsecase)
	orderHandler := handler.NewOrderHandler(orderUsecase)

	// Initialize Gin Engine
	r := gin.Default()

	// Setup Router (also registers global middleware)
	httpdelivery.NewRouter(r, menuHandler, orderHandler)

	// Use a custom http.Server with timeouts to protect against slow-loris
	// and other slow-connection attacks.
	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Wait for OS shutdown signal (SIGINT or SIGTERM).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give in-flight requests up to 10 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
