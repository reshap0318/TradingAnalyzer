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
			// POST /api/trade/monitor/:id/close - Manual close trade by ID
			monitorGroup.POST("/:id/close", ctrl.TradeManualClose)
		}

		// Trade Bot control endpoints
		botGroup := tradeGroup.Group("/bot")
		{
			// GET /api/trade/bot - Get all trades with optional filters (?status=&symbol=&interval=&side=&min_confidence=&date_start=&date_end=&limit=)
			botGroup.GET("", ctrl.TradeBotGetAll)
			// GET /api/trade/bot/status - Get bot status
			botGroup.GET("/status", ctrl.TradeBotGetStatus)
			// GET /api/trade/bot/summary - Get session summary statistics
			botGroup.GET("/summary", ctrl.TradeBotGetSessionSummary)
			// GET /api/trade/bot/session - Get trades executed in current bot session
			botGroup.GET("/session", ctrl.TradeBotGetExecutedTrades)
			// POST /api/trade/bot/activate - Activate bot
			botGroup.POST("/activate", ctrl.TradeBotActivate)
			// POST /api/trade/bot/deactivate - Deactivate bot
			botGroup.POST("/deactivate", ctrl.TradeBotDeactivate)
		}
	}
}
