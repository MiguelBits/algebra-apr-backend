package router

import (
	"algebra-apr-backend/internal/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Configure CORS to allow specific origins
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://localhost:8080",
			"https://satsuma.exchange",
			"https://www.satsuma.exchange",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
