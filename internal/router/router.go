package router

import (
	"algebra-apr-backend/internal/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	handler := handlers.NewHandler(db)

	api := r.Group("/api")
	{
		pools := api.Group("/pools")
		{
			pools.GET("/apr", handler.GetPoolsAPR)
			pools.GET("/max-apr", handler.GetPoolsMaxAPR)
		}

		eternalFarmings := api.Group("/eternal-farmings")
		{
			eternalFarmings.GET("/apr", handler.GetEternalFarmingsAPR)
			eternalFarmings.GET("/max-apr", handler.GetFarmingsMaxAPR)
			eternalFarmings.GET("/tvl", handler.GetFarmingsTVL)
		}
	}

	return r
}
