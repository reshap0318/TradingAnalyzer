package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterSignalRoutes registers all signal-related routes
func RegisterSignalRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	signalGroup := router.Group("/signal")
	{
		// POST /api/signal/raws - Get raw OHLCV data for multiple timeframes
		signalGroup.POST("/raws", ctrl.SignalRawIndex)

		// POST /api/signal/analyze - Analyze market and generate trading signal
		signalGroup.POST("/analyze", ctrl.SignalAnalyzeIndex)
	}
}
