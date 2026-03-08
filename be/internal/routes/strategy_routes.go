package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterStrategyRoutes registers all strategy-related routes
func RegisterStrategyRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	strategyGroup := router.Group("/strategies")
	{
		strategyGroup.GET("", ctrl.StrategyIndex)
		strategyGroup.GET("/active", ctrl.StrategyActive)
		strategyGroup.GET("/:id", ctrl.StrategyDetail)
		strategyGroup.POST("", ctrl.StrategyCreate)
		strategyGroup.PUT("/:id", ctrl.StrategyUpdate)
		strategyGroup.DELETE("/:id", ctrl.StrategyDelete)
	}
}
