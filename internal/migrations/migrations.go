package migrations

import (
	"algebra-apr-backend/internal/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "202507060001_create_networks_table",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Network{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&models.Network{})
			},
		},
		{
			ID: "202507060002_create_pools_table",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Pool{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&models.Pool{})
			},
		},
		{
			ID: "202507060003_create_eternal_farmings_table",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Farming{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&models.Farming{})
			},
		},
	}
}
