package database

import (
	"algebra-apr-backend/internal/config"
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/models"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ImportNetwork(db *gorm.DB, networkConfig config.Network) error {
	var network models.Network
	result := db.Where("title = ?", networkConfig.Title).First(&network)

	if result.Error != nil {
		// Create new network
		network = models.Network{
			Title:                networkConfig.Title,
			AnalyticsSubgraphURL: networkConfig.AnalyticsSubgraphURL,
			FarmingSubgraphURL:   networkConfig.FarmingSubgraphURL,
			APIKey:               networkConfig.APIKey,
		}
		if err := db.Create(&network).Error; err != nil {
			return err
		}
		logger.Logger.Info("Created network", zap.String("title", network.Title))
	} else {
		// Update existing network
		network.AnalyticsSubgraphURL = networkConfig.AnalyticsSubgraphURL
		network.FarmingSubgraphURL = networkConfig.FarmingSubgraphURL
		network.APIKey = networkConfig.APIKey
		if err := db.Save(&network).Error; err != nil {
			return err
		}
		logger.Logger.Info("Updated network", zap.String("title", network.Title))
	}
	return nil
}
