package service

import (
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/database"
	"github.com/reshap/trading-bot/internal/repository"
)

// Services holds all application services
type Services struct {
	repo          *repository.Repositories
	cfg           *config.Config
	RedisClient   *database.CacheClient
	BinanceClient *binance.Client
}

// NewServicesMinimal creates services with minimal initialization (for config loading)
// Does NOT initialize BinanceClient or TradingEngine
func NewServicesMinimal(repo *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		repo: repo,
		cfg:  cfg,
		// BinanceClient and TradingEngine are nil
	}
}

// NewServices creates and initializes all services with full initialization
func NewServices(repo *repository.Repositories, cfg *config.Config, redisClient *database.CacheClient) (*Services, error) {
	// Initialize Binance Futures client with Redis cache
	binanceClient := binance.NewClient(&cfg.BINANCE, redisClient)

	return &Services{
		repo:          repo,
		cfg:           cfg,
		RedisClient:   redisClient,
		BinanceClient: binanceClient,
	}, nil
}
