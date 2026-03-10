package config

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

// ConfigService defines interface for getting config data
type ConfigService interface {
	ConfigGetAll(ctx *gin.Context) (res []dtos.ConfigData, err error)
	ThresholdGetAll(ctx *gin.Context) (res []dtos.ThresholdData, err error)
}

// AppConfig holds application configuration
type AppConfig struct {
	Host      string
	Port      string
	JWTSecret []byte
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       int
}

// MMConfig holds money management configuration
type MMConfig struct {
	MIN_CONFIDENCE         int8    // minimum confidence level to take a trade
	MAX_DAILY_TRADES       int8    // Soft limit trades per hari
	MAX_DAILY_LOSS_PERCENT float32 // ..% from balance (HARD LIMIT)
	MAX_DAILY_LOSS_COUNT   int8    //Max consecutive losses (HARD LIMIT)
	RISK_REWARD_RATIO      float32 // Minimum R:R required
	RISK_REWARD_TARGET     float32 // Target R:R untuk excellent setups
	RISK_ENTRY_BUFFER      float32 // ..% Buffer RISK and Entry
	MAX_POSITION_SIZE      float32 // ..% from balance per trade

	LEVERAGE               int8 // Leverage to use for futures trading
	IS_AGRESSIVE           bool // true = hybrid entry, false = conservative
	ORDER_EXPIRATION_HOURS int8 // Hours before pending orders expire (optimal for 15m timeframe: 2-4 hours)
}

// ThresholdConfig holds threshold configuration
type ThresholdConfig struct {
	CATEGORY string
	MAX      int
	MIN      int
	ACTION   string
}

// BinanceConfig holds Binance API configuration
type BinanceConfig struct {
	IsTestnet  bool
	APIKey     string
	SecretKey  string
	Timeout    int
	MaxRetries int
	RetryDelay int
}

// IndicatorMACD holds MACD indicator parameters
type IndicatorMACD struct {
	FAST   int
	SLOW   int
	SIGNAL int
}

// IndicatorRSI holds RSI indicator parameters
type IndicatorRSI struct {
	PERIOD     int
	OVERBOUGHT float64
	OVERSOLD   float64
}

// IndicatorStochastic holds Stochastic indicator parameters
type IndicatorStochastic struct {
	K_PERIOD   int
	D_PERIOD   int
	SMOOTH     int
	OVERBOUGHT float64
	OVERSOLD   float64
}

// IndicatorBollinger holds Bollinger Bands indicator parameters
type IndicatorBollinger struct {
	PERIOD  int
	STD_DEV float64
}

// IndicatorATR holds ATR indicator parameters
type IndicatorATR struct {
	PERIOD int
}

// IndicatorMovingAverage holds Moving Average indicator parameters
type IndicatorMovingAverage struct {
	SMA_PERIODS []int // e.g., []int{20, 50, 200}
	EMA_PERIODS []int // e.g., []int{12, 26}
}

// IndicatorVolume holds Volume indicator parameters
type IndicatorVolume struct {
	MA_PERIOD int
}

// SRConfig holds LuxAlgo Support & Resistance configuration
type SRConfig struct {
	LOOKBACK_PERIODS int     // Number of candles to analyze for swing points
	TOLERANCE        float64 // Price tolerance for clustering levels (e.g., 0.01 = 1%)
	MIN_TOUCHES      int     // Minimum touches required to validate a level
}

// IndicatorsConfig holds all indicators configuration
type IndicatorsConfig struct {
	MACD           IndicatorMACD
	RSI            IndicatorRSI
	STOCHASTIC     IndicatorStochastic
	BOLLINGER      IndicatorBollinger
	ATR            IndicatorATR
	VOLUME         IndicatorVolume
	MOVING_AVERAGE IndicatorMovingAverage
	SUPPORT_RESIST SRConfig // LuxAlgo Support & Resistance
}

// Config holds all application configuration
type Config struct {
	APP        AppConfig
	DB         DBConfig
	Redis      RedisConfig
	MM         MMConfig
	Threshold  []ThresholdConfig
	BINANCE    BinanceConfig
	INDICATORS IndicatorsConfig
}

func LoadConfig() *Config {
	config := Config{}

	// APP config
	config.APP.Host = helpers.GetEnv("APP_HOST", "0.0.0.0")
	config.APP.Port = helpers.GetEnv("APP_PORT", "8080")
	config.APP.JWTSecret = []byte(helpers.GetEnv("APP_JWT_SECRET", "your-secret-key-change-in-production"))

	// DB config
	config.DB.Host = helpers.GetEnv("DB_HOST", "127.0.0.1")
	config.DB.Port = helpers.GetEnv("DB_PORT", "3306")
	config.DB.User = helpers.GetEnv("DB_USER", "root")
	config.DB.Password = helpers.GetEnv("DB_PASSWORD", "")
	config.DB.DBName = helpers.GetEnv("DB_NAME", "trading_bot")

	// Redis config
	config.Redis.Host = helpers.GetEnv("REDIS_HOST", "localhost")
	config.Redis.Port = helpers.GetEnv("REDIS_PORT", "6379")
	config.Redis.User = helpers.GetEnv("REDIS_USER", "default")
	config.Redis.Password = helpers.GetEnv("REDIS_PASSWORD", "")
	config.Redis.DB = helpers.GetEnvInt("REDIS_DB", 0)

	// Binance config
	config.BINANCE.IsTestnet = true
	config.BINANCE.APIKey = helpers.GetEnv("TESTNET_API_KEY", "")
	config.BINANCE.SecretKey = helpers.GetEnv("TESTNET_SECRET_KEY", "")

	// Money Management config
	config.MM.MIN_CONFIDENCE = 65
	config.MM.MAX_DAILY_TRADES = 10
	config.MM.MAX_DAILY_LOSS_PERCENT = 0.05
	config.MM.MAX_DAILY_LOSS_COUNT = 5
	config.MM.RISK_REWARD_RATIO = 2.0
	config.MM.RISK_REWARD_TARGET = 3.0
	config.MM.RISK_ENTRY_BUFFER = 0.0075 // 0.75%
	config.MM.MAX_POSITION_SIZE = 0.15   // 15% of balance

	config.MM.LEVERAGE = 5
	config.MM.IS_AGRESSIVE = false

	config.MM.ORDER_EXPIRATION_HOURS = 4

	// Threshold config
	config.Threshold = []ThresholdConfig{
		{CATEGORY: "STRONG_BUY", MAX: 100, MIN: 70, ACTION: "BUY"},
		{CATEGORY: "BUY", MAX: 70, MIN: 45, ACTION: "BUY"},
		{CATEGORY: "WAIT", MAX: 45, MIN: -45, ACTION: "WAIT"},
		{CATEGORY: "SELL", MAX: -45, MIN: -70, ACTION: "SELL"},
		{CATEGORY: "STRONG_SELL", MAX: -70, MIN: -100, ACTION: "SELL"},
	}

	// Indicators config - default values
	config.INDICATORS.MACD = IndicatorMACD{
		FAST:   12,
		SLOW:   26,
		SIGNAL: 9,
	}

	config.INDICATORS.RSI = IndicatorRSI{
		PERIOD:     14,
		OVERBOUGHT: 70,
		OVERSOLD:   30,
	}

	config.INDICATORS.STOCHASTIC = IndicatorStochastic{
		K_PERIOD:   14,
		D_PERIOD:   3,
		SMOOTH:     3,
		OVERBOUGHT: 80,
		OVERSOLD:   20,
	}

	config.INDICATORS.BOLLINGER = IndicatorBollinger{
		PERIOD:  20,
		STD_DEV: 2.0,
	}

	config.INDICATORS.ATR = IndicatorATR{
		PERIOD: 14,
	}

	config.INDICATORS.VOLUME = IndicatorVolume{
		MA_PERIOD: 20,
	}

	config.INDICATORS.MOVING_AVERAGE = IndicatorMovingAverage{
		SMA_PERIODS: []int{20, 50, 200},
		EMA_PERIODS: []int{12, 26},
	}

	config.INDICATORS.SUPPORT_RESIST = SRConfig{
		LOOKBACK_PERIODS: 200,
		TOLERANCE:        0.005,
		MIN_TOUCHES:      2,
	}

	return &config
}

// LoadConfigDB loads config from database, overriding default values
// This function should be called after services are initialized
func LoadConfigDB(srvc ConfigService) *Config {
	cfg := LoadConfig()

	// Map config values from database
	mmConfigs, _ := srvc.ConfigGetAll(nil)
	for _, data := range mmConfigs {
		switch data.ConfigKey {
		case "MIN_CONFIDENCE":
			if val, err := helpers.ParseFloat(data.Value, 8); err == nil {
				cfg.MM.MIN_CONFIDENCE = int8(val)
			}
		case "MAX_DAILY_TRADES":
			if val, err := helpers.ParseFloat(data.Value, 8); err == nil {
				cfg.MM.MAX_DAILY_TRADES = int8(val)
			}
		case "MAX_DAILY_LOSS_PERCENT":
			if val, err := helpers.ParseFloat(data.Value, 8); err == nil {
				cfg.MM.MAX_DAILY_LOSS_PERCENT = float32(val)
			}
		case "MAX_DAILY_LOSS_COUNT":
			if val, err := helpers.ParseFloat(data.Value, 8); err == nil {
				cfg.MM.MAX_DAILY_LOSS_COUNT = int8(val)
			}
		case "RISK_REWARD_RATIO":
			if val, err := helpers.ParseFloat(data.Value, 32); err == nil {
				cfg.MM.RISK_REWARD_RATIO = float32(val)
			}
		case "RISK_REWARD_TARGET":
			if val, err := helpers.ParseFloat(data.Value, 32); err == nil {
				cfg.MM.RISK_REWARD_TARGET = float32(val)
			}
		case "MAX_POSITION_SIZE":
			if val, err := helpers.ParseFloat(data.Value, 32); err == nil {
				cfg.MM.MAX_POSITION_SIZE = float32(val)
			}
		case "RISK_ENTRY_BUFFER":
			if val, err := helpers.ParseFloat(data.Value, 32); err == nil {
				cfg.MM.RISK_ENTRY_BUFFER = float32(val)
			}
		case "LEVERAGE":
			if val, err := helpers.ParseFloat(data.Value, 8); err == nil {
				cfg.MM.LEVERAGE = int8(val)
			}
		case "ORDER_EXPIRATION_HOURS":
			if val, err := helpers.ParseFloat(data.Value, 16); err == nil {
				cfg.MM.ORDER_EXPIRATION_HOURS = int8(val)
			}
		case "IS_AGRESSIVE":
			cfg.MM.IS_AGRESSIVE = data.Value == "true"
		case "BINANCE_TESTNET":
			cfg.BINANCE.IsTestnet = data.Value == "true"
			if cfg.BINANCE.IsTestnet {
				cfg.BINANCE.APIKey = helpers.GetEnv("TESTNET_API_KEY", "")
				cfg.BINANCE.SecretKey = helpers.GetEnv("TESTNET_SECRET_KEY", "")
			} else {
				cfg.BINANCE.APIKey = helpers.GetEnv("BINANCE_API_KEY", "")
				cfg.BINANCE.SecretKey = helpers.GetEnv("BINANCE_SECRET_KEY", "")
			}
		}
	}

	threshold, err := srvc.ThresholdGetAll(nil)
	if err == nil {
		cfg.Threshold = make([]ThresholdConfig, len(threshold))
		for i, th := range threshold {
			cfg.Threshold[i] = ThresholdConfig{
				CATEGORY: th.Category,
				MAX:      th.MaxValue,
				MIN:      th.MinValue,
				ACTION:   th.Action,
			}
		}
	}

	return cfg
}
