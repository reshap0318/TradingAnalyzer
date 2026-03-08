package di

import (
	"fmt"

	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/controller"
	"github.com/reshap/trading-bot/internal/database"
	"github.com/reshap/trading-bot/internal/repository"
	"github.com/reshap/trading-bot/internal/service"
	"gorm.io/gorm"
)

type Container struct {
	Cfg   *config.Config
	DB    *gorm.DB
	Redis *database.CacheClient
	Repo  *repository.Repositories
	Srvc  *service.Services
	Ctrl  *controller.Controller
}

func NewContainer(cfg *config.Config) (*Container, error) {

	// 1. Initialize MySQL
	db, err := database.NewMySQLConnection(&cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	// 2. Initialize Redis
	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		// Log error but continue (Redis is optional for some features)
		fmt.Printf("Warning: Redis connection failed: %v\n", err)
		fmt.Println("Continuing without Redis. Some features may not be available.")
	}

	// 3. Create CacheClient wrapper
	var cacheClient *database.CacheClient
	if redisClient != nil {
		cacheClient = database.NewCacheClient(redisClient)
		fmt.Println("Redis connected successfully")
	}

	// 4. Initialize repositories
	repo, _ := repository.NewRepositories(db)

	// 5. Create minimal services for config loading (before full initialization)
	tempSrvc := service.NewServicesMinimal(repo, cfg)

	// 6. Load config from database to override environment values
	cfg = config.LoadConfigDB(tempSrvc)

	// 7. Initialize full services with updated config (includes BinanceClient & TradingEngine)
	srvc, _ := service.NewServices(repo, cfg, cacheClient)

	// 8. Initialize controller
	ctrl := controller.NewController(srvc, cfg)

	return &Container{
		Cfg:   cfg,
		DB:    db,
		Redis: cacheClient,
		Repo:  repo,
		Srvc:  srvc,
		Ctrl:  ctrl,
	}, nil
}
