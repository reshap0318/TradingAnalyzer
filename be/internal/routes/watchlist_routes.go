package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterWatchlistRoutes registers all watchlist-related routes
func RegisterWatchlistRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	watchlistGroup := router.Group("/watchlists")
	{
		watchlistGroup.GET("", ctrl.WatchlistIndex)
		watchlistGroup.GET("/status", ctrl.WatchlistScannerGetStatus)
		watchlistGroup.GET("/:id", ctrl.WatchlistDetail)

		watchlistGroup.POST("/activate", ctrl.WatchlistScannerActivate)
		watchlistGroup.POST("/deactivate", ctrl.WatchlistScannerDeactivate)
		watchlistGroup.POST("", ctrl.WatchlistCreate)

		watchlistGroup.PUT("/:id", ctrl.WatchlistUpdate)
		watchlistGroup.DELETE("/:id", ctrl.WatchlistDelete)

	}
}
