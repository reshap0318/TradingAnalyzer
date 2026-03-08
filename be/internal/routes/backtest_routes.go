package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterBacktestRoutes registers all backtest-related routes
func RegisterBacktestRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	backtestGroup := router.Group("/backtests")
	{
		backtestGroup.GET("", ctrl.BacktestIndex)
		backtestGroup.GET("/:id", ctrl.BacktestDetail)
		backtestGroup.POST("", ctrl.BacktestCreate)
		backtestGroup.DELETE("/:id", ctrl.BacktestDelete)
	}
}
