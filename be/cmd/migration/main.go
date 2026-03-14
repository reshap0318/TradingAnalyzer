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
			&models.Trade{},
			&models.TradeEntry{},
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
			&models.Trade{},
			&models.TradeEntry{},
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
	// Seed watchlist
	seedWatchlist(db)

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
			Color:        "red",
			OrderDisplay: 4,
		},
		{
			Category:     "STRONG_SELL",
			MinValue:     -100,
			MaxValue:     -70,
			Action:       "SELL",
			Color:        "dark-red",
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
			ConfigKey: "RISK_ENTRY_BUFFER",
			Value:     "0.0075",
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
	// Indicator IDs (from seedIndicator):
	// 1=MA, 2=MACD, 3=RSI, 4=Stochastic, 5=BB, 6=Volume, 7=Candle Patterns, 8=ATR, 9=Trend Bonus

	type tfSeed struct {
		Name   string
		Weight float64
	}
	type indSeed struct {
		ID     uint
		Weight float64
	}
	type mmSeed struct {
		Param string
		Value string
	}
	type strategySeed struct {
		Name       string
		PrimaryTF  string
		Timeframes []tfSeed
		Indicators []indSeed
		MoneyMgmt  []mmSeed
	}

	strategies := []strategySeed{
		// ──────────────────────────────────────────────
		// 1. Scalper Pro - Ultra short-term, high frequency
		// ──────────────────────────────────────────────
		{
			Name:      "Scalper Pro",
			PrimaryTF: "1m",
			Timeframes: []tfSeed{
				{"1m", 0.50}, {"5m", 0.30}, {"15m", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.20}, {2, 0.15}, {3, 0.18}, {4, 0.12},
				{5, 0.15}, {6, 0.10}, {7, 0.05}, {8, 0.02}, {9, 0.03},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "20"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_POSITION_SIZE", "0.10"},
				{"RISK_REWARD_RATIO", "1.5"}, {"LEVERAGE", "10"},
				{"ORDER_EXPIRATION_HOURS", "1"}, {"IS_AGRESSIVE", "true"},
				{"MAX_RISK_PER_TRADE", "0.02"},
			},
		},

		// ──────────────────────────────────────────────
		// 2. Day Trading Pro - Balanced intraday
		// Fixed: 15m/1h/4h instead of 15m/30m/1h
		// ──────────────────────────────────────────────
		{
			Name:      "Day Trading Pro",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.10},
				{"15m", 0.40},
				{"1h", 0.30},
				{"4h", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.25},
				{2, 0.20},
				{3, 0.10},
				{4, 0.10},
				{5, 0.10},
				{6, 0.15},
				{7, 0.05},
				{8, 0.05},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "3"}, // Hemat peluru!
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.20"}, // 20% dari $37 = $7.4
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"}, // Wajib antri di S/R
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ──────────────────────────────────────────────
		// 3. Momentum Hunter - Catches strong moves
		// High MACD+RSI+Volume weight
		// ──────────────────────────────────────────────
		{
			Name:      "Momentum Hunter",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"15m", 0.40}, {"1h", 0.35}, {"4h", 0.25},
			},
			Indicators: []indSeed{
				{1, 0.18}, {2, 0.25}, {3, 0.20}, {4, 0.08},
				{5, 0.08}, {6, 0.12}, {7, 0.03}, {8, 0.02}, {9, 0.04},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_POSITION_SIZE", "0.12"},
				{"RISK_REWARD_RATIO", "2.5"}, {"LEVERAGE", "7"},
				{"ORDER_EXPIRATION_HOURS", "3"}, {"IS_AGRESSIVE", "true"},
				{"MAX_RISK_PER_TRADE", "0.03"},
			},
		},

		// ──────────────────────────────────────────────
		// 4. Swing Trader - Multi-day holds
		// High MA+Trend weight, low leverage, wide SL
		// ──────────────────────────────────────────────
		{
			Name:      "Swing Trader",
			PrimaryTF: "4h",
			Timeframes: []tfSeed{
				{"4h", 0.40}, {"1d", 0.35}, {"1w", 0.25},
			},
			Indicators: []indSeed{
				{1, 0.30}, {2, 0.22}, {3, 0.12}, {4, 0.08},
				{5, 0.08}, {6, 0.06}, {7, 0.04}, {8, 0.04}, {9, 0.06},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.06"}, {"MAX_POSITION_SIZE", "0.20"},
				{"RISK_REWARD_RATIO", "3.0"}, {"LEVERAGE", "3"},
				{"ORDER_EXPIRATION_HOURS", "24"}, {"IS_AGRESSIVE", "false"},
				{"MAX_RISK_PER_TRADE", "0.04"},
			},
		},

		// ──────────────────────────────────────────────
		// 5. Trend Follower - Ride the trend
		// MA dominant (35%), strong trend bonus
		// ──────────────────────────────────────────────
		{
			Name:      "Trend Follower",
			PrimaryTF: "1h",
			Timeframes: []tfSeed{
				{"1h", 0.40}, {"4h", 0.35}, {"1d", 0.25},
			},
			Indicators: []indSeed{
				{1, 0.35}, {2, 0.20}, {3, 0.10}, {4, 0.05},
				{5, 0.07}, {6, 0.08}, {7, 0.03}, {8, 0.04}, {9, 0.08},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_POSITION_SIZE", "0.18"},
				{"RISK_REWARD_RATIO", "2.5"}, {"LEVERAGE", "4"},
				{"ORDER_EXPIRATION_HOURS", "8"}, {"IS_AGRESSIVE", "false"},
				{"MAX_RISK_PER_TRADE", "0.04"},
			},
		},

		// ──────────────────────────────────────────────
		// 6. Breakout Scalper - Volatility breakout catches
		// BB+Volume dominant, aggressive entries
		// ──────────────────────────────────────────────
		{
			Name:      "Breakout Scalper",
			PrimaryTF: "5m",
			Timeframes: []tfSeed{
				{"5m", 0.50}, {"15m", 0.30}, {"1h", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.15}, {2, 0.15}, {3, 0.12}, {4, 0.08},
				{5, 0.20}, {6, 0.18}, {7, 0.05}, {8, 0.04}, {9, 0.03},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "15"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_POSITION_SIZE", "0.10"},
				{"RISK_REWARD_RATIO", "2.0"}, {"LEVERAGE", "8"},
				{"ORDER_EXPIRATION_HOURS", "2"}, {"IS_AGRESSIVE", "true"},
				{"MAX_RISK_PER_TRADE", "0.02"},
			},
		},

		// ──────────────────────────────────────────────
		// 7. Conservative Intraday - Low risk, steady returns
		// Wider TF gaps (30m/2h/8h), low leverage
		// ──────────────────────────────────────────────
		{
			Name:      "Conservative Intraday",
			PrimaryTF: "30m",
			Timeframes: []tfSeed{
				{"30m", 0.35}, {"2h", 0.35}, {"8h", 0.30},
			},
			Indicators: []indSeed{
				{1, 0.28}, {2, 0.18}, {3, 0.15}, {4, 0.12},
				{5, 0.10}, {6, 0.07}, {7, 0.04}, {8, 0.03}, {9, 0.03},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_POSITION_SIZE", "0.10"},
				{"RISK_REWARD_RATIO", "2.0"}, {"LEVERAGE", "3"},
				{"ORDER_EXPIRATION_HOURS", "6"}, {"IS_AGRESSIVE", "false"},
				{"MAX_RISK_PER_TRADE", "0.02"},
			},
		},

		// ──────────────────────────────────────────────
		// 8. Meme Coin Sniper - Optimized for volatile meme coins
		// Volume+RSI heavy (meme = volume driven), aggressive
		// ──────────────────────────────────────────────
		{
			Name:      "Meme Coin Sniper",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"15m", 0.45}, {"1h", 0.35}, {"4h", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.15}, {2, 0.18}, {3, 0.18}, {4, 0.10},
				{5, 0.12}, {6, 0.15}, {7, 0.06}, {8, 0.03}, {9, 0.03},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.06"}, {"MAX_POSITION_SIZE", "0.12"},
				{"RISK_REWARD_RATIO", "2.0"}, {"LEVERAGE", "5"},
				{"ORDER_EXPIRATION_HOURS", "3"}, {"IS_AGRESSIVE", "true"},
				{"MAX_RISK_PER_TRADE", "0.04"},
			},
		},
	}

	for _, s := range strategies {
		var existing models.Strategy
		result := db.Where(models.Strategy{Name: s.Name}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			strategy := models.Strategy{
				Name:      s.Name,
				PrimaryTF: s.PrimaryTF,
				IsActive:  s.PrimaryTF == "15m",
			}
			if err := db.Create(&strategy).Error; err != nil {
				log.Printf("Failed to create strategy '%s': %v", s.Name, err)
				continue
			}

			for _, tf := range s.Timeframes {
				if err := db.Create(&models.StrategyTimeframe{
					StrategyID:    strategy.ID,
					TimeframeName: tf.Name,
					Weight:        tf.Weight,
				}).Error; err != nil {
					log.Printf("  Failed to create TF %s: %v", tf.Name, err)
				}
			}

			for _, ind := range s.Indicators {
				if err := db.Create(&models.StrategyIndicator{
					StrategyID:  strategy.ID,
					IndicatorID: ind.ID,
					Weight:      ind.Weight,
				}).Error; err != nil {
					log.Printf("  Failed to create indicator %d: %v", ind.ID, err)
				}
			}

			for _, mm := range s.MoneyMgmt {
				if err := db.Create(&models.StrategyMoneyMgmt{
					StrategyID: strategy.ID,
					Parameter:  mm.Param,
					Value:      mm.Value,
				}).Error; err != nil {
					log.Printf("  Failed to create MM %s: %v", mm.Param, err)
				}
			}

			fmt.Printf("✓ Created strategy: %s (Primary: %s)\n", s.Name, s.PrimaryTF)
		} else if result.Error != nil {
			log.Printf("Failed to check strategy '%s': %v", s.Name, result.Error)
		} else {
			fmt.Printf("⊘ Strategy already exists: %s\n", s.Name)
		}
	}
}

func seedWatchlist(db *gorm.DB) {
	// Default watchlist symbols
	defaultWatchlists := []models.Watchlist{
		{Symbol: "BTCUSDT", IsActive: true},
		{Symbol: "ETHUSDT", IsActive: true},
		{Symbol: "BNBUSDT", IsActive: true},
		{Symbol: "SOLUSDT", IsActive: true},
		{Symbol: "XRPUSDT", IsActive: true},
		{Symbol: "ADAUSDT", IsActive: true},
		{Symbol: "DOGEUSDT", IsActive: true},
		{Symbol: "TRXUSDT", IsActive: true},
		{Symbol: "LINKUSDT", IsActive: true},
		{Symbol: "AVAXUSDT", IsActive: true},
		{Symbol: "SUIUSDT", IsActive: true},
		{Symbol: "LTCUSDT", IsActive: true},
		{Symbol: "NEARUSDT", IsActive: true},
		{Symbol: "UNIUSDT", IsActive: true},
		{Symbol: "FETUSDT", IsActive: true},
	}

	for _, wl := range defaultWatchlists {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Watchlist
		result := db.Where(models.Watchlist{Symbol: wl.Symbol}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&wl).Error; err != nil {
				log.Printf("Failed to create watchlist %s: %v", wl.Symbol, err)
			} else {
				fmt.Printf("✓ Created watchlist: %s\n", wl.Symbol)
			}
		} else if result.Error != nil {
			log.Printf("Failed to check watchlist %s: %v", wl.Symbol, result.Error)
		} else {
			fmt.Printf("⊘ Watchlist already exists: %s\n", wl.Symbol)
		}
	}
}
