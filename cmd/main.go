package main

import (
	"algebra-apr-backend/internal/config"
	"algebra-apr-backend/internal/database"
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/router"
	"algebra-apr-backend/internal/scheduler"
	"algebra-apr-backend/internal/services"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger.InitLogger(true)

	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Logger.Fatal("Failed to load config", zap.Error(err))
	}
	logger.Logger.Info("cfg loaded", zap.Any("cfg", cfg))

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Note: Migrations are now handled separately via cmd/migrate
	// Run 'make migrate-and-run' to apply migrations and start the app
	logger.Logger.Info("Database connection established successfully")

	// Import networks from config
	for _, network := range cfg.Networks {
		err = database.ImportNetwork(db, network)
		if err != nil {
			logger.Logger.Fatal("failed to import network from config", zap.String("network", network.Title))
		}
	}

	// Initialize APR service without GraphQL clients (they will be created dynamically)
	aprService := services.NewAPRService(db)

	// Initialize scheduler for background tasks
	taskScheduler := scheduler.NewScheduler(db, cfg, aprService)
	taskScheduler.Start()

	// Initialize router
	r := router.SetupRouter(db)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down server...")

	// Stop scheduler
	taskScheduler.Stop()

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server exited")
}
