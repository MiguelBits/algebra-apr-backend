package handlers

import (
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// GET /api/pools/apr?network=Polygon
func (h *Handler) GetPoolsAPR(c *gin.Context) {
	networkName := c.DefaultQuery("network", "Polygon")

	var pools []models.Pool
	result := h.db.Preload("Network").Joins("JOIN networks ON pools.network_id = networks.id").Where("networks.title = ?", networkName).Find(&pools)
	if result.Error != nil {
		logger.Logger.Error("Failed to fetch pools", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pools"})
		return
	}

	response := make(map[string]interface{})
	for _, pool := range pools {
		if pool.LastAPR != nil {
			response[pool.Address] = *pool.LastAPR
		} else {
			response[pool.Address] = 0.0
		}
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/pools/max-apr?network=Polygon
func (h *Handler) GetPoolsMaxAPR(c *gin.Context) {
	networkName := c.DefaultQuery("network", "Polygon")

	var pools []models.Pool
	result := h.db.Preload("Network").Joins("JOIN networks ON pools.network_id = networks.id").Where("networks.title = ?", networkName).Find(&pools)
	if result.Error != nil {
		logger.Logger.Error("Failed to fetch pools", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pools"})
		return
	}

	response := make(map[string]interface{})
	for _, pool := range pools {
		if pool.MaxAPR != nil {
			response[pool.Address] = *pool.MaxAPR
		} else {
			response[pool.Address] = 0.0
		}
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/eternal-farmings/apr?network=Polygon
func (h *Handler) GetEternalFarmingsAPR(c *gin.Context) {
	networkName := c.DefaultQuery("network", "Polygon")

	var farmings []models.Farming
	result := h.db.Preload("Network").Joins("JOIN networks ON farmings.network_id = networks.id").Where("networks.title = ?", networkName).Find(&farmings)
	if result.Error != nil {
		logger.Logger.Error("Failed to fetch eternal farmings", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch eternal farmings"})
		return
	}

	response := make(map[string]interface{})
	for _, farming := range farmings {
		if farming.LastAPR != nil {
			response[farming.Hash] = *farming.LastAPR
		} else {
			response[farming.Hash] = 0.0
		}
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/eternal-farmings/max-apr?network=Polygon
func (h *Handler) GetFarmingsMaxAPR(c *gin.Context) {
	networkName := c.DefaultQuery("network", "Polygon")

	var farmings []models.Farming
	result := h.db.Preload("Network").Joins("JOIN networks ON farmings.network_id = networks.id").Where("networks.title = ?", networkName).Find(&farmings)
	if result.Error != nil {
		logger.Logger.Error("Failed to fetch eternal farmings", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch eternal farmings"})
		return
	}

	response := make(map[string]interface{})
	for _, farming := range farmings {
		if farming.MaxAPR != nil {
			response[farming.Hash] = *farming.MaxAPR
		} else {
			response[farming.Hash] = 0.0
		}
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/eternal-farmings/tvl?network=Polygon
func (h *Handler) GetFarmingsTVL(c *gin.Context) {
	networkName := c.DefaultQuery("network", "Polygon")

	var farmings []models.Farming
	result := h.db.Preload("Network").Joins("JOIN networks ON farmings.network_id = networks.id").Where("networks.title = ?", networkName).Find(&farmings)
	if result.Error != nil {
		logger.Logger.Error("Failed to fetch eternal farmings", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch eternal farmings"})
		return
	}

	response := make(map[string]interface{})
	for _, farming := range farmings {
		if farming.TVL != nil {
			response[farming.Hash] = *farming.TVL
		} else {
			response[farming.Hash] = 0.0
		}
	}

	c.JSON(http.StatusOK, response)
}
