package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/di"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/middleware"
	"github.com/reshap/trading-bot/internal/routes"
)

func main() {
	// Load .env file
	godotenv.Load()

	host := helpers.GetEnv("APP_HOST", "0.0.0.0")
	port := helpers.GetEnv("APP_PORT", "8000")

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.SetTrustedProxies(nil)

	cfg := config.LoadConfig()

	// Initialize engine
	engine, err := di.NewContainer(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize container: %v", err))
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Setup routes
	apiGroup := router.Group("/api")
	routes.RegisterAuthRoutes(apiGroup, engine.Ctrl, engine.Srvc)

	protected := apiGroup.Group("")
	protected.Use(middleware.AuthMiddleware(engine.Srvc))
	{
		routes.RegisterThresholdRoutes(protected, engine.Ctrl)
		routes.RegisterTimeframeRoutes(protected, engine.Ctrl)
		routes.RegisterIndicatorRoutes(protected, engine.Ctrl)
		routes.RegisterConfigRoutes(protected, engine.Ctrl)
		routes.RegisterWatchlistRoutes(protected, engine.Ctrl)
		routes.RegisterStrategyRoutes(protected, engine.Ctrl)
		routes.RegisterSignalRoutes(protected, engine.Ctrl)
		routes.RegisterBacktestRoutes(protected, engine.Ctrl)
	}

	router.Run(fmt.Sprintf("%s:%s", host, port))
}
