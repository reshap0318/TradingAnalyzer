package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterSignalRoutes registers all signal-related routes
func RegisterSignalRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	signalGroup := router.Group("/signal")
	{
		// GET /api/signal - List signals with pagination and filters
		signalGroup.GET("", ctrl.SignalIndex)

		// GET /api/signal/:id - Get signal details with all snapshots
		signalGroup.GET("/:id", ctrl.SignalDetail)

		// DELETE /api/signal/:id - Delete a signal by ID
		signalGroup.DELETE("/:id", ctrl.SignalDelete)

		// POST /api/signal/cleanup - Cleanup old signals
		signalGroup.POST("/cleanup", ctrl.SignalCleanup)

		// POST /api/signal/raws - Get raw OHLCV data for multiple timeframes
		signalGroup.POST("/raws", ctrl.SignalRawIndex)

		// POST /api/signal/analyze - Analyze market and generate trading signal
		signalGroup.POST("/analyze", ctrl.SignalAnalyzeIndex)
	}
}
