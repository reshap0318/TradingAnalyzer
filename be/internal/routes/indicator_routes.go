package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterIndicatorRoutes registers all indicator-related routes
func RegisterIndicatorRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	indicatorGroup := router.Group("/indicators")
	{
		indicatorGroup.GET("", ctrl.IndicatorIndex)
		indicatorGroup.GET("/active", ctrl.IndicatorGetActive)
		indicatorGroup.GET("/:id", ctrl.IndicatorDetail)
		indicatorGroup.POST("", ctrl.IndicatorCreate)
		indicatorGroup.PUT("/:id", ctrl.IndicatorUpdate)
		indicatorGroup.DELETE("/:id", ctrl.IndicatorDelete)
	}
}
