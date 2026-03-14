package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterTradeRoutes registers all trade-related routes
func RegisterTradeRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	tradeGroup := router.Group("/trade")
	{
		// POST /api/trade/ - Execute single trade manually
		tradeGroup.POST("/execute", ctrl.TradeExecute)

		// Trade Monitor endpoints (for debugging/manual trigger)
		monitorGroup := tradeGroup.Group("/monitor")
		{
			// POST /api/trade/monitor/all - Process all active trades
			monitorGroup.POST("/all", ctrl.TradeMonitorProcessAllActive)
			// POST /api/trade/monitor/:id - Process single trade by ID
			monitorGroup.POST("/:id", ctrl.TradeMonitorProcessSingle)
		}

		// Trade Bot control endpoints
		botGroup := tradeGroup.Group("/bot")
		{
			// GET /api/trade/bot/status - Get bot status
			botGroup.GET("/status", ctrl.TradeBotGetStatus)
			// GET /api/trade/bot/summary - Get session summary
			botGroup.GET("/summary", ctrl.TradeBotGetSessionSummary)
			// GET /api/trade/bot/active - Get active trades
			botGroup.GET("/active", ctrl.TradeBotGetActiveTrades)
			// GET /api/trade/bot/ - Get executed trades in current session
			botGroup.GET("/", ctrl.TradeBotGetExecutedTrades)
			// POST /api/trade/bot/activate - Activate bot
			botGroup.POST("/activate", ctrl.TradeBotActivate)
			// POST /api/trade/bot/deactivate - Deactivate bot
			botGroup.POST("/deactivate", ctrl.TradeBotDeactivate)
		}
	}
}
