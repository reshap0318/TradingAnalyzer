package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterTradeRoutes registers all trade-related routes
func RegisterTradeRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	tradeGroup := router.Group("/trade")
	{
		// POST /api/trade/auto - Analyze and Execute trade automatically
		tradeGroup.POST("/auto", ctrl.TradeExecute)
	}
}
