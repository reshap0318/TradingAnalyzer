package binance

import (
	"context"
	"log"
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
		// Only required for older go-binance versions, but safe to set
		futures.UseTestnet = true
	}

	log.Printf("Creating BinanceClient - Testnet: %v | baseURL: %s", cfg.IsTestnet, client.BaseURL)

	// Sync Server Time with Binance to prevent "Timestamp outside recvWindow (-1021)" error
	timeOffset, err := client.NewSetServerTimeService().Do(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to sync server time with Binance: %v", err)
	} else {
		log.Printf("Binance server time synced successfully. Offset: %d ms", timeOffset)
	}

	return &Client{
		apiClient: client,
		config: &Config{
			APIKey:         cfg.APIKey,
			SecretKey:      cfg.SecretKey,
			IsTestnet:      cfg.IsTestnet,
			BaseURL:        client.BaseURL,
			Timeout:        time.Duration(cfg.Timeout) * time.Second,
			MaxRetries:     cfg.MaxRetries,
			RetryDelay:     time.Duration(cfg.RetryDelay) * time.Second,
			RecvWindow:     5000, // Default 5 seconds
			ServerTimeOffset: timeOffset,
		},
		cache:    cacheClient,
		cacheCfg: DefaultCacheConfig(),
	}
}

// SetCacheConfig sets custom cache configuration
func (c *Client) SetCacheConfig(cacheCfg *CacheConfig) {
	c.cacheCfg = cacheCfg
}

// GetConfig returns the current Binance configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetContextWithTimeout creates a context with timeout
func (c *Client) GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// IsCacheAvailable checks if cache is enabled and available
func (c *Client) IsCacheAvailable() bool {
	return c.cache != nil && c.cacheCfg.Enabled
}
