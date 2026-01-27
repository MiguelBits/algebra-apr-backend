package main

import (
	"algebra-apr-backend/internal/config"
	"algebra-apr-backend/internal/database"
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/migrations"
	"fmt"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger.InitLogger(true)

	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Check command line arguments (default to "up" for Heroku release phase)
	action := "up"
	if len(os.Args) >= 2 {
		action = os.Args[1]
	} else {
		logger.Logger.Info("No argument provided, defaulting to 'up' migration")
	}
	switch action {
	case "up":
		runMigrations(db, "up")
	case "down":
		runMigrations(db, "down")
	case "reset":
		runMigrations(db, "reset")
	default:
		fmt.Println("Invalid argument. Use 'up', 'down', or 'reset'.")
		os.Exit(1)
	}
}

func runMigrations(database *gorm.DB, action string) {
	m := gormigrate.New(database, gormigrate.DefaultOptions, migrations.GetMigrations())

	switch action {
	case "up":
		logger.Logger.Info("Applying migrations...")
		err := m.Migrate()
		if err != nil {
			logger.Logger.Fatal("Failed to apply migrations", zap.Error(err))
		}
		logger.Logger.Info("Migrations applied successfully")
	case "down":
		logger.Logger.Info("Rolling back the last migration...")
		err := m.RollbackLast()
		if err != nil {
			logger.Logger.Fatal("Failed to roll back the last migration", zap.Error(err))
		}
		logger.Logger.Info("Last migration rolled back successfully")
	case "reset":
		logger.Logger.Info("Resetting all migrations...")
		err := m.RollbackTo("0")
		if err != nil {
			logger.Logger.Warn("Failed to rollback migrations (this is normal for first run)", zap.Error(err))
		}
		err = m.Migrate()
		if err != nil {
			logger.Logger.Fatal("Failed to apply migrations after reset", zap.Error(err))
		}
		logger.Logger.Info("All migrations reset and applied successfully")
	}
}
