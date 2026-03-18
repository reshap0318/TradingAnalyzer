package binance

import "time"

// Config holds Binance Futures configuration
type Config struct {
	APIKey         string
	SecretKey      string
	IsTestnet      bool
	BaseURL        string
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	RecvWindow     int64 // milliseconds, default 5000
	ServerTimeOffset int64 // milliseconds offset from server time
}

// DefaultConfig returns default configuration for Binance Futures
func DefaultConfig() *Config {
	return &Config{
		IsTestnet:  true,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		RecvWindow: 5000, // 5 seconds default
		ServerTimeOffset: 0,
	}
}

// Validate validates configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return ErrAPIKeyRequired
	}
	if c.SecretKey == "" {
		return ErrSecretKeyRequired
	}
	return nil
}
