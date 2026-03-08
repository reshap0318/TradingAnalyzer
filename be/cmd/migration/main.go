package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get database connection string from environment
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Database configuration is incomplete. Please set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, and DB_NAME")
	}

	// Build MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigration(dsn, "up")
	case "down":
		runMigration(dsn, "down")
	case "seed":
		runSeed(dsn)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migration/main.go <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up       Apply all migrations (create/update tables)")
	fmt.Println("  down     Drop all migrated tables")
	fmt.Println("  seed     Insert default data")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migration/main.go up")
	fmt.Println("  go run cmd/migration/main.go down")
	fmt.Println("  go run cmd/migration/main.go seed")
}

func runMigration(dsn string, command string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch command {
	case "up":
		fmt.Println("Running migrations...")

		// AutoMigrate threshold, interval, config, watchlist, strategy, and signal tables
		err := db.AutoMigrate(
			// master
			&models.Threshold{},
			&models.Timeframe{},
			&models.Config{},
			&models.Indicators{},
			// strategy
			&models.Strategy{},
			&models.StrategyTimeframe{},
			&models.StrategyIndicator{},
			&models.StrategyMoneyMgmt{},
			// transactional
			&models.Watchlist{},
			&models.Signal{},
			&models.SignalEntry{},
			&models.Backtest{},
			&models.BacktestTrade{},
		)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		fmt.Println("Migration completed successfully!")

	case "down":
		fmt.Println("Rolling back migrations...")

		// Drop tables
		err := db.Migrator().DropTable(
			// master
			&models.Threshold{},
			&models.Timeframe{},
			&models.Config{},
			&models.Indicators{},
			// strategy
			&models.Strategy{},
			&models.StrategyTimeframe{},
			&models.StrategyIndicator{},
			&models.StrategyMoneyMgmt{},
			// transactional
			&models.Watchlist{},
			&models.Signal{},
			&models.SignalEntry{},
			&models.Backtest{},
			&models.BacktestTrade{},
		)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

		fmt.Println("Rollback completed successfully!")
	}
}

func runSeed(dsn string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Seeding default data...")

	seedThreshold(db)
	seedConfig(db)
	// Seed timeframes
	seedTimeframes(db)
	// Seed Indicator for each indicator
	seedIndicator(db)
	// Seed strategy
	seedStrategy(db)

	fmt.Println("Seeding completed!")
}

func seedThreshold(db *gorm.DB) {
	// Default threshold data
	defaultThresholds := []models.Threshold{
		{
			Category:     "STRONG_BUY",
			MinValue:     70,
			MaxValue:     100,
			Action:       "BUY",
			Color:        "green",
			OrderDisplay: 1,
		},
		{
			Category:     "BUY",
			MinValue:     45,
			MaxValue:     70,
			Action:       "BUY",
			Color:        "light-green",
			OrderDisplay: 2,
		},
		{
			Category:     "WAIT",
			MinValue:     -45,
			MaxValue:     45,
			Action:       "WAIT",
			Color:        "gray",
			OrderDisplay: 3,
		},
		{
			Category:     "SELL",
			MinValue:     -70,
			MaxValue:     -45,
			Action:       "SELL",
			Color:        "light-red",
			OrderDisplay: 4,
		},
		{
			Category:     "STRONG_SELL",
			MinValue:     -100,
			MaxValue:     -70,
			Action:       "SELL",
			Color:        "red",
			OrderDisplay: 5,
		},
	}

	for _, threshold := range defaultThresholds {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Threshold
		result := db.Where(models.Threshold{Category: threshold.Category, MinValue: threshold.MinValue}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&threshold).Error; err != nil {
				log.Printf("Failed to create threshold %s: %v", threshold.Category, err)
			}
		} else if result.Error != nil {
			log.Printf("Failed to check threshold %s: %v", threshold.Category, result.Error)
		}
	}
}

func seedConfig(db *gorm.DB) {
	// Default config data
	defaultConfigs := []models.Config{
		{
			ConfigKey: "MIN_CONFIDENCE",
			Value:     "45",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_TRADES",
			Value:     "10",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_LOSS_PERCENT",
			Value:     "0.05",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_LOSS_COUNT",
			Value:     "7",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_POSITION_SIZE",
			Value:     "0.15",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_RISK_PER_TRADE",
			Value:     "0.04",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "RISK_REWARD_RATIO",
			Value:     "1.5",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "RISK_REWARD_TARGET",
			Value:     "3.0",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "LEVERAGE",
			Value:     "5",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "IS_AGRESSIVE",
			Value:     "false", // true | false
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "ORDER_EXPIRATION_HOURS",
			Value:     "4", // in hour
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "BINANCE_TESTNET",
			Value:     "true", // true | false
			Category:  "BINANCE",
		},
	}

	for _, cfg := range defaultConfigs {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Config
		result := db.Where(models.Config{ConfigKey: cfg.ConfigKey, Category: cfg.Category}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&cfg).Error; err != nil {
				log.Printf("Failed to create config %s: %v", cfg.ConfigKey, err)
			}
		} else if result.Error != nil {
			log.Printf("Failed to check config %s: %v", cfg.ConfigKey, result.Error)
		}
	}
}

func seedTimeframes(db *gorm.DB) {
	// Using slice to maintain order as defined in the original map
	timeframeData := []struct {
		name      string
		inMinutes int
	}{
		// Scalping (sangat cepat, 1-15 menit hold)
		{"1m", 1},
		{"3m", 3},
		{"5m", 5},
		// Short-term scalping (15-30 menit hold)
		{"15m", 15},
		{"30m", 30},
		// Day trading (beberapa jam hold)
		{"1h", 60},
		{"2h", 120},
		{"4h", 240},
		{"6h", 360},
		{"8h", 480},
		{"12h", 720},
		// Swing trading (beberapa hari hold)
		{"1d", 1440},
		{"3d", 4320},
		// Position trading (mingguan/bulanan)
		{"1w", 10080},
		{"1M", 43200},
	}

	for _, data := range timeframeData {
		// Create or get interval
		var interval models.Timeframe
		result := db.Where(models.Timeframe{Name: data.name}).First(&interval)

		if result.Error == gorm.ErrRecordNotFound {
			interval = models.Timeframe{
				Name:      data.name,
				InMinutes: data.inMinutes,
			}
			if err := db.Create(&interval).Error; err != nil {
				log.Printf("Failed to create interval %s: %v", data.name, err)
				continue
			}
		} else if result.Error != nil {
			log.Printf("Failed to check interval %s: %v", data.name, result.Error)
			continue
		}
	}
}

func seedIndicator(db *gorm.DB) {
	// Default Indicator for each indicator with parameters
	// Scoring is hardcoded in indicator functions (internal/bot/indicators/)
	// Total weight normalized to 1.0 (100%)
	//
	// Weight Distribution:
	// - Core Indicators (65%): MA, MACD, RSI - Primary signals
	// - Secondary Indicators (25%): Stoch, BB, Volume - Confirmation
	// - Special Indicators (10%): Patterns, ATR, Trend Bonus - Context
	defaultIndicator := []struct {
		Name        string
		Indicator   string
		Description string
		Params      string
		Weight      float64
		OrderView   int
	}{
		// Core Indicators (65%)
		{
			Name:        "Moving Average",
			Indicator:   "moving_average",
			Description: "Moving Average - Trend indicator using multiple SMAs and EMAs",
			Params:      `{"sma_periods": [20, 50, 200], "ema_periods": [12, 26]}`,
			Weight:      0.30,
			OrderView:   1,
		},
		{
			Name:        "MACD",
			Indicator:   "macd",
			Description: "Moving Average Convergence Divergence - Trend-following momentum indicator",
			Params:      `{"fast_period": 12, "slow_period": 26, "signal_period": 9}`,
			Weight:      0.22,
			OrderView:   2,
		},
		{
			Name:        "RSI",
			Indicator:   "rsi",
			Description: "Relative Strength Index - Momentum oscillator measuring speed and magnitude of price changes",
			Params:      `{"period": 14, "overbought": 70, "oversold": 30}`,
			Weight:      0.13,
			OrderView:   3,
		},

		// Secondary Indicators (25%)
		{
			Name:        "Stochastic",
			Indicator:   "stochastic",
			Description: "Stochastic Oscillator - Momentum indicator comparing closing price to price range (with trend neutralization)",
			Params:      `{"k_period": 14, "d_period": 3, "smooth": 3, "overbought": 80, "oversold": 20}`,
			Weight:      0.10,
			OrderView:   4,
		},
		{
			Name:        "Bollinger Bands",
			Indicator:   "bollinger_bands",
			Description: "Bollinger Bands - Volatility bands at standard deviations above/below moving average",
			Params:      `{"period": 20, "std_dev": 2.0}`,
			Weight:      0.10,
			OrderView:   5,
		},
		{
			Name:        "Volume",
			Indicator:   "volume",
			Description: "Volume analysis - Comparing current volume to moving average",
			Params:      `{"ma_period": 20}`,
			Weight:      0.05,
			OrderView:   6,
		},

		// Special Indicators (10%)
		{
			Name:        "Candle Patterns",
			Indicator:   "candle_patterns",
			Description: "Candlestick Patterns - Pattern recognition for reversal and continuation patterns (last 5 candles)",
			Params:      `{}`,
			Weight:      0.04,
			OrderView:   7,
		},
		{
			Name:        "ATR",
			Indicator:   "atr",
			Description: "Average True Range - Volatility indicator measuring price movement magnitude",
			Params:      `{"period": 14}`,
			Weight:      0.02,
			OrderView:   8,
		},
		{
			Name:        "Trend Bonus",
			Indicator:   "trend_bonus",
			Description: "Trend alignment bonus - Rewards when MA and MACD are aligned (strong trend). Returns ±20 signal",
			Params:      `{}`,
			Weight:      0.04,
			OrderView:   9,
		},
	}

	for _, indicatorData := range defaultIndicator {
		var existing models.Indicators
		result := db.Where(models.Indicators{Name: indicatorData.Name}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			indicator := models.Indicators{
				Name:        indicatorData.Name,
				Indicator:   indicatorData.Indicator,
				Description: indicatorData.Description,
				IsActive:    true,
				Weight:      indicatorData.Weight,
				OrderView:   indicatorData.OrderView,
			}
			if indicatorData.Params != "" {
				indicator.Params = []byte(indicatorData.Params)
			}

			if err := db.Create(&indicator).Error; err != nil {
				log.Printf("Failed to create indicator %s: %v", indicatorData.Name, err)
			} else {
				fmt.Printf("✓ Created indicator: %s (%s) - Weight: %.2f\n", indicatorData.Name, indicatorData.Indicator, indicatorData.Weight)
			}
		} else if result.Error != nil {
			log.Printf("Failed to check indicator %s: %v", indicatorData.Name, result.Error)
		} else {
			// Indicator exists, update if needed
			existing.Description = indicatorData.Description
			existing.Params = []byte(indicatorData.Params)
			existing.Weight = indicatorData.Weight
			existing.OrderView = indicatorData.OrderView
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("Failed to update indicator %s: %v", indicatorData.Name, err)
			} else {
				fmt.Printf("✓ Updated indicator: %s (%s) - Weight: %.2f\n", indicatorData.Name, indicatorData.Indicator, indicatorData.Weight)
			}
		}
	}
}

func seedStrategy(db *gorm.DB) {
	// Check if strategy already exists
	var existing models.Strategy
	result := db.Where(models.Strategy{Name: "Day Trading Pro"}).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// 1. Create main strategy
		strategy := models.Strategy{
			Name:      "Day Trading Pro",
			PrimaryTF: "15m",
			IsActive:  true,
		}

		if err := db.Create(&strategy).Error; err != nil {
			log.Printf("Failed to create strategy: %v", err)
			return
		}

		// 2. Create strategy timeframes
		timeframes := []models.StrategyTimeframe{
			{StrategyID: strategy.ID, TimeframeName: "15m", Weight: 0.50},
			{StrategyID: strategy.ID, TimeframeName: "30m", Weight: 0.30},
			{StrategyID: strategy.ID, TimeframeName: "1h", Weight: 0.20},
		}

		for _, tf := range timeframes {
			if err := db.Create(&tf).Error; err != nil {
				log.Printf("Failed to create strategy timeframe: %v", err)
			}
		}

		// 3. Create strategy indicators
		indicators := []models.StrategyIndicator{
			{StrategyID: strategy.ID, IndicatorID: 1, Weight: 0.30}, // Moving Average
			{StrategyID: strategy.ID, IndicatorID: 2, Weight: 0.22}, // MACD
			{StrategyID: strategy.ID, IndicatorID: 3, Weight: 0.13}, // RSI
			{StrategyID: strategy.ID, IndicatorID: 4, Weight: 0.10}, // Stochastic
			{StrategyID: strategy.ID, IndicatorID: 5, Weight: 0.10}, // Bollinger Bands
			{StrategyID: strategy.ID, IndicatorID: 6, Weight: 0.05}, // Volume
			{StrategyID: strategy.ID, IndicatorID: 7, Weight: 0.04}, // Candle Patterns
			{StrategyID: strategy.ID, IndicatorID: 8, Weight: 0.02}, // ATR
			{StrategyID: strategy.ID, IndicatorID: 9, Weight: 0.04}, // Trend Bonus
		}

		for _, ind := range indicators {
			if err := db.Create(&ind).Error; err != nil {
				log.Printf("Failed to create strategy indicator: %v", err)
			}
		}

		// 4. Create strategy money management
		moneyMgmt := []models.StrategyMoneyMgmt{
			{StrategyID: strategy.ID, Parameter: "MIN_CONFIDENCE", Value: "45"},
			{StrategyID: strategy.ID, Parameter: "MAX_DAILY_TRADES", Value: "10"},
			{StrategyID: strategy.ID, Parameter: "MAX_DAILY_LOSS_PERCENT", Value: "0.05"},
			{StrategyID: strategy.ID, Parameter: "MAX_POSITION_SIZE", Value: "0.15"},
			{StrategyID: strategy.ID, Parameter: "RISK_REWARD_RATIO", Value: "1.5"},
			{StrategyID: strategy.ID, Parameter: "LEVERAGE", Value: "5"},
			{StrategyID: strategy.ID, Parameter: "ORDER_EXPIRATION_HOURS", Value: "4"},
		}

		for _, mm := range moneyMgmt {
			if err := db.Create(&mm).Error; err != nil {
				log.Printf("Failed to create strategy money management: %v", err)
			}
		}

		fmt.Println("✓ Created strategy: Day Trading Pro")
	} else if result.Error != nil {
		log.Printf("Failed to check strategy: %v", result.Error)
	}
}
