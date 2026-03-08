package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
)

// RegisterConfigRoutes registers all config-related routes
func RegisterConfigRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
	configGroup := router.Group("/configs")
	{
		configGroup.GET("", ctrl.ConfigIndex)
		configGroup.GET("/:id", ctrl.ConfigDetail)
		configGroup.GET("/category/:category", ctrl.ConfigGetByCategory)
		configGroup.POST("", ctrl.ConfigCreate)
		configGroup.PUT("/:id", ctrl.ConfigUpdate)
		configGroup.DELETE("/:id", ctrl.ConfigDelete)
	}
}
