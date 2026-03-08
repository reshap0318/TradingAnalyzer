package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/controller"
	"github.com/reshap/trading-bot/internal/middleware"
	"github.com/reshap/trading-bot/internal/service"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(router *gin.RouterGroup, ctrl *controller.Controller, srvc *service.Services) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", ctrl.Login)

		// Protected routes (require authentication)
		protected := authGroup.Group("")
		protected.Use(middleware.AuthMiddleware(srvc))
		{
			protected.POST("/logout", ctrl.Logout)
			protected.GET("/me", ctrl.Me)
		}
	}
}
