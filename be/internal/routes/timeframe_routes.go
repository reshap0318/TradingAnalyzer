package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterTimeframeRoutes registers all timeframe-related routes
func RegisterTimeframeRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	timeframeGroup := router.Group("/timeframes")
	{
		timeframeGroup.GET("", ctrl.TimeframeIndex)
		timeframeGroup.GET("/:name", ctrl.TimeframeDetail)
		timeframeGroup.POST("", ctrl.TimeframeCreate)
		timeframeGroup.PUT("/:name", ctrl.TimeframeUpdate)
		timeframeGroup.DELETE("/:name", ctrl.TimeframeDelete)
	}
}
