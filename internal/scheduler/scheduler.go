package scheduler

import (
	"algebra-apr-backend/internal/config"
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/models"
	"algebra-apr-backend/internal/services"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Scheduler struct {
	db         *gorm.DB
	scheduler  *gocron.Scheduler
	aprService *services.APRService
	config     *config.Config
}

func NewScheduler(db *gorm.DB, cfg *config.Config, aprService *services.APRService) *Scheduler {
	s := gocron.NewScheduler(time.UTC)

	return &Scheduler{
		db:         db,
		scheduler:  s,
		aprService: aprService,
		config:     cfg,
	}
}

func (s *Scheduler) Start() {
	logger.Logger.Info("Starting scheduler...", zap.Int("apr_update_minutes", s.config.APRUpdateMinutes))

	// Schedule unified APR update task
	s.scheduler.Every(s.config.APRUpdateMinutes).Minutes().Do(s.updateAllAPR)

	// Start the scheduler
	s.scheduler.StartAsync()
}

func (s *Scheduler) Stop() {
	logger.Logger.Info("Stopping scheduler...")
	s.scheduler.Stop()
}

func (s *Scheduler) updateAllAPR() {
	logger.Logger.Info("Running unified APR update task")

	var networks []models.Network
	if err := s.db.Find(&networks).Error; err != nil {
		logger.Logger.Error("Failed to get networks", zap.Error(err))
		return
	}

	// Process networks in parallel
	var wg sync.WaitGroup
	for _, network := range networks {
		wg.Add(1)
		go func(net models.Network) {
			defer wg.Done()
			if err := s.aprService.UpdateAllAPR(net.ID); err != nil {
				logger.Logger.Error("Failed to update all APR",
					zap.String("network", net.Title),
					zap.Error(err))
			} else {
				logger.Logger.Info("Updated all APR", zap.String("network", net.Title))
			}
		}(network)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	logger.Logger.Info("Completed unified APR update task for all networks")
}
