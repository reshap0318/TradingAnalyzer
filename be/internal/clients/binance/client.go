package binance

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/database"
)

// Client represents Binance Futures API client
type Client struct {
	apiClient *futures.Client
	config    *Config
	cache     *database.CacheClient
	cacheCfg  *CacheConfig
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled    bool
	DefaultTTL time.Duration
	PriceTTL   time.Duration // TTL for price data
	KlineTTL   time.Duration // TTL for kline data
	AccountTTL time.Duration // TTL for account data
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:    true,
		DefaultTTL: 10 * time.Second,
		PriceTTL:   5 * time.Second,
		KlineTTL:   30 * time.Second,
		AccountTTL: 10 * time.Second,
	}
}

// Close closes the client connection (if needed)
func (c *Client) Close() error {
	// Cleanup resources if needed
	return nil
}

// NewClient creates new Binance Futures client
func NewClient(cfg *config.BinanceConfig, cacheClient *database.CacheClient) *Client {
	client := futures.NewClient(cfg.APIKey, cfg.SecretKey)

	// Use Futures testnet or mainnet
	if cfg.IsTestnet {
		client.BaseURL = "https://testnet.binancefuture.com"
	}

	return &Client{
		apiClient: client,
		config: &Config{
			APIKey:     cfg.APIKey,
			SecretKey:  cfg.SecretKey,
			IsTestnet:  cfg.IsTestnet,
			BaseURL:    client.BaseURL,
			Timeout:    time.Duration(cfg.Timeout) * time.Second,
			MaxRetries: cfg.MaxRetries,
			RetryDelay: time.Duration(cfg.RetryDelay) * time.Second,
		},
		cache:    cacheClient,
		cacheCfg: DefaultCacheConfig(),
	}
}

// SetCacheConfig sets custom cache configuration
func (c *Client) SetCacheConfig(cacheCfg *CacheConfig) {
	c.cacheCfg = cacheCfg
}

// GetContextWithTimeout creates a context with timeout
func (c *Client) GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// IsCacheAvailable checks if cache is enabled and available
func (c *Client) IsCacheAvailable() bool {
	return c.cache != nil && c.cacheCfg.Enabled
}
