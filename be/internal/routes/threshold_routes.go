package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterThresholdRoutes registers all threshold-related routes
func RegisterThresholdRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	thresholdGroup := router.Group("/thresholds")
	{
		thresholdGroup.GET("", ctrl.ThresholdIndex)
		thresholdGroup.GET("/:id", ctrl.ThresholdDetail)
		thresholdGroup.POST("", ctrl.ThresholdCreate)
		thresholdGroup.PUT("/:id", ctrl.ThresholdUpdate)
		thresholdGroup.DELETE("/:id", ctrl.ThresholdDelete)
	}
}