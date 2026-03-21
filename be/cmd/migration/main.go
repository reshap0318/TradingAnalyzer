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
	case "refresh":
		runMigration(dsn, "down")
		runMigration(dsn, "up")
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
			MinValue:     25,
			MaxValue:     70,
			Action:       "BUY",
			Color:        "light-green",
			OrderDisplay: 2,
		},
		{
			Category:     "WAIT",
			MinValue:     -25,
			MaxValue:     25,
			Action:       "WAIT",
			Color:        "gray",
			OrderDisplay: 3,
		},
		{
			Category:     "SELL",
			MinValue:     -70,
			MaxValue:     -25,
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

		// ==============================================================================
		// KATEGORI 1: AGRESIF (HIGH FREQUENCY / VOLATILITY)
		// Karakter: TF Kecil, Sinyal Sering, TAPI Money Management super ketat (Pengaman kencang)
		// ==============================================================================

		// ──────────────────────────────────────────────
		// 1. Micro Scalper (Aggressive Signal, Defensive MM)
		// Fokus: Nangkap riak kecil di TF 5m. Indikator difilter murni ke Trend & Vol.
		// Pengaman: Position size sangat kecil (5%), R:R masuk akal (1.5), batas loss harian ketat.
		// Koin: Koin L1/Narasi Hype yang lagi aktif (SUI, SEI, APT, FET, RNDR)
		// ──────────────────────────────────────────────
		{
			Name:      "Micro Scalper Pro",
			PrimaryTF: "5m",
			Timeframes: []tfSeed{
				{"1m", 0.30}, {"5m", 0.70},
			},
			Indicators: []indSeed{
				{1, 0.40}, // MA (Trend)
				{2, 0.40}, // MACD (Momentum)
				{6, 0.20}, // Volume (Validasi)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "45"},                                          // Sinyal gampang trigger
				{"MAX_DAILY_TRADES", "12"},                                        // Boleh trade sering
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "3"}, // KALAH 3X SEHARI LANGSUNG OFF!
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.0"},
				{"RISK_ENTRY_BUFFER", "0.001"},
				{"MAX_POSITION_SIZE", "0.05"}, // PENGAMAN KETAT: Cuma 5% modal per trade
				{"LEVERAGE", "10"},
				{"IS_AGRESSIVE", "true"}, // Hajar market biar gak telat
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},

		// ──────────────────────────────────────────────
		// 2. Volatility Breakout Catcher
		// Fokus: Nunggu koin yang tiba-tiba pumping/dumping (BB Squeeze + Volume).
		// Pengaman: Entry Buffer dilebarin dikit, pakai Limit Order (IsAgressive=false) biar dapet harga wajar.
		// Koin: Meme Coins / High Volatility (DOGE, PEPE, WIF, SHIB, FLOKI)
		// ──────────────────────────────────────────────
		{
			Name:      "Volatility Breakout",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.40}, {"15m", 0.60},
			},
			Indicators: []indSeed{
				{5, 0.50}, // BB (Deteksi harga nembus band)
				{6, 0.30}, // Volume (Wajib ada ledakan volume)
				{1, 0.20}, // MA (Arah tren)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"},
				{"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.06"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.8"}, {"RISK_REWARD_TARGET", "3.0"}, // Reward harus gede buat nutupin resiko
				{"RISK_ENTRY_BUFFER", "0.003"},
				{"MAX_POSITION_SIZE", "0.05"}, // PENGAMAN: Size kecil
				{"LEVERAGE", "5"},
				{"IS_AGRESSIVE", "false"}, // PENGAMAN: Nunggu harga retest (Limit order)
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 3. Knife Catcher (Counter-Trend)
		// Fokus: Nangkap koin yang oversold parah (RSI & Stoch di bawah) dan mantul di BB Bawah.
		// Pengaman: Syarat R:R harus tinggi (2.0 minimal) karena ngelawan arus.
		// Koin: Top 10 Caps yang likuiditasnya tebal, cepat mantul (BTC, ETH, SOL, BNB)
		// ──────────────────────────────────────────────
		{
			Name:      "Knife Catcher",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.30}, {"15m", 0.70},
			},
			Indicators: []indSeed{
				{3, 0.40}, // RSI
				{4, 0.30}, // Stochastic
				{5, 0.30}, // Bollinger Bands
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"},
				{"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"}, // PENGAMAN: Reward harus 2x lipat resiko
				{"RISK_ENTRY_BUFFER", "0.005"},
				{"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "5"},
				{"IS_AGRESSIVE", "false"}, // Wajib limit order di dasar S/R
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ==============================================================================
		// KATEGORI 2: BALANCED (DAY TRADING / MID-TERM)
		// Karakter: Kombinasi indikator seimbang, santai tapi aktif, modal/size menengah.
		// ==============================================================================

		// ──────────────────────────────────────────────
		// 4. Sniper 15m (Unleashed)
		// Fokus: Day trading ideal. Numpang tren 15 menitan tanpa diganggu oscillator.
		// Koin: Universal (Semua Top 50 CMC dengan volume di atas $50M)
		// ──────────────────────────────────────────────
		{
			Name:      "Sniper 15m (Unleashed)",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.40}, {"15m", 0.60},
			},
			Indicators: []indSeed{
				{1, 0.40}, // MA
				{2, 0.40}, // MACD
				{6, 0.20}, // Volume
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"},
				{"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"},
				{"MAX_POSITION_SIZE", "0.10"}, // Size normal (10%)
				{"LEVERAGE", "5"},
				{"IS_AGRESSIVE", "true"}, // Boleh semi-agresif
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ──────────────────────────────────────────────
		// 5. Active Day Trader (Smooth Trend)
		// Fokus: Menggabungkan tren (MA/MACD) dan batasan volatilitas (BB) biar nggak masuk pas pucuk.
		// Koin: Layer 1 & 2 Solid (AVAX, MATIC/POL, ARB, OP, LINK)
		// ──────────────────────────────────────────────
		{
			Name:      "Active Day Trader",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.20}, {"15m", 0.50}, {"1h", 0.30},
			},
			Indicators: []indSeed{
				{1, 0.35}, // MA
				{2, 0.35}, // MACD
				{5, 0.20}, // BB
				{6, 0.10}, // Vol
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"},
				{"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"},
				{"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"},
				{"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ──────────────────────────────────────────────
		// 6. Day Trading Pro (Micro $37)
		// Fokus: Spesial buat akun modal kecil. Wajib dapet Maker Fee dan lolos batas min order.
		// Koin: Bluechips dengan spread rapat (SOL, BNB, ETH, ADA)
		// ──────────────────────────────────────────────
		{
			Name:       "Day Trading Pro (Micro $37)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.10}, {"15m", 0.50}, {"1h", 0.40}},
			Indicators: []indSeed{{1, 0.35}, {2, 0.35}, {6, 0.20}, {8, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"},
				{"MAX_DAILY_TRADES", "3"}, // Hemat fee
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"},
				{"MAX_POSITION_SIZE", "0.20"}, // Size gede (20%) biar lolos min notional $5
				{"LEVERAGE", "10"},            // Lev 10x biar margin ketutup
				{"IS_AGRESSIVE", "false"},     // WAJIB false biar dapet Maker Fee (murah)
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ==============================================================================
		// KATEGORI 3: DEFENSIF (SWING / POSITION)
		// Karakter: TF Besar, Hold berhari-hari. Santai, anti-stop-hunt, leverage kecil.
		// ==============================================================================

		// ──────────────────────────────────────────────
		// 7. Safe Swing Investor
		// Fokus: Mengikuti tren mayor mingguan. Syarat konfirmasi sangat ketat.
		// Koin: HANYA Bluechips & Major L1 (BTC, ETH, BNB, SOL)
		// ──────────────────────────────────────────────
		{
			Name:       "Safe Swing Investor",
			PrimaryTF:  "4h",
			Timeframes: []tfSeed{{"1h", 0.20}, {"4h", 0.50}, {"1d", 0.30}},
			Indicators: []indSeed{
				{1, 0.40}, // MA
				{2, 0.30}, // MACD
				{9, 0.30}, // Trend Bonus (Wajib alignment sempurna)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"},
				{"MAX_DAILY_TRADES", "1"}, // Max 1 posisi
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.005"},
				{"MAX_POSITION_SIZE", "0.25"}, // Berani size gede karena risk kecil
				{"LEVERAGE", "2"},             // Leverage super kecil, anti liquid
				{"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "48"}, // Ngantre bisa 2 hari
			},
		},

		// ──────────────────────────────────────────────
		// 8. Golden Cross Seeker
		// Fokus: Murni mencari persilangan Moving Average (MA) di TF 1 Hari. Setup klasiknya para paus.
		// Koin: Koin dengan kapitalisasi raksasa (BTC, ETH, XRP)
		// ──────────────────────────────────────────────
		{
			Name:       "Golden Cross Seeker",
			PrimaryTF:  "1d",
			Timeframes: []tfSeed{{"4h", 0.30}, {"1d", 0.70}},
			Indicators: []indSeed{
				{1, 0.60}, // MA Dominan
				{6, 0.20}, // Volume konfirmasi
				{9, 0.20}, // Trend Bonus
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"},
				{"MAX_DAILY_TRADES", "1"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.008"},
				{"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "2"},
				{"IS_AGRESSIVE", "true"}, // Golden cross jarang terjadi, hajar market 1, limit 1
				{"ORDER_EXPIRATION_HOURS", "72"},
			},
		},

		// ──────────────────────────────────────────────
		// 9. Mean Reversion Master (Sideways)
		// Fokus: Beli di support harian, jual di resistance harian. Indikator tren (MA) dibuang.
		// Koin: Koin lama yang pergerakannya lambat/ranging (ADA, DOT, TRX, LTC)
		// ──────────────────────────────────────────────
		{
			Name:       "Mean Reversion Master",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.60}, {"4h", 0.40}},
			Indicators: []indSeed{
				{3, 0.40}, // RSI
				{4, 0.30}, // Stochastic
				{5, 0.30}, // Bollinger Bands
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"},
				{"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"},
				{"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "3"},
				{"IS_AGRESSIVE", "false"}, // Mutlak limit order di batas S/R
				{"ORDER_EXPIRATION_HOURS", "24"},
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
				IsActive:  false,
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

	// db.Model(&models.Strategy{}).Where("strategy_name != ?", "Day Trading Pro").Update("is_active", false)
	db.Model(&models.Strategy{}).Where("strategy_name = ?", "Day Trading Pro").Update("is_active", true)

	fmt.Println("✓ Updated strategy IsActive (Day Trading Pro = true, others = false)")
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
