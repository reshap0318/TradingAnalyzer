package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedStrategies inserts multiple strategy configurations
func SeedStrategies(db *gorm.DB) {
	fmt.Println("Seeding strategies...")

	type tfSeed struct {
		Name   string
		Weight float64
	}
	type indSeed struct {
		ID     uint
		Weight float64
		TF     *string // nil = all TFs, "15m" = only 15m
	}
	type mmSeed struct {
		Param string
		Value string
	}
	type symSeed struct {
		Symbol string
	}
	type strategySeed struct {
		Name       string
		PrimaryTF  string
		Timeframes []tfSeed
		Indicators []indSeed
		MoneyMgmt  []mmSeed
		Symbols    []symSeed
	}

	// Helper to create *string for TF targeting
	tf := func(s string) *string { return &s }

	strategies := []strategySeed{
		// ==========================================
		// MICRO & SCALPING STRATEGIES (5m)
		// ==========================================

		{
			Name:       "Micro Breakout Sniper",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.50}, {"15m", 0.30}},
			Indicators: []indSeed{
				// 15m: MA confirms macro direction
				{1, 0.80, tf("15m")}, // MA Driver on 15m
				{2, 0.20, tf("15m")}, // MACD Driver on 15m
				// 5m: Volume + BB entry trigger
				{5, 0.20, tf("5m")},  // BB Filter on 5m (veto 80% if no squeeze)
				{6, 2.00, tf("5m")},  // Volume Booster on 5m
				// 1m: Candle precision
				{7, 1.30, tf("1m")},  // Candle Booster on 1m
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "12"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
			Symbols: []symSeed{
				{"SOLUSDT"}, {"SUIUSDT"}, {"SEIUSDT"}, {"PEPEUSDT"}, {"WIFUSDT"},
			},
		},
		{
			Name:       "RSI Reversal Catcher",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.60}, {"15m", 0.20}},
			Indicators: []indSeed{
				{1, 0.50, tf("15m")}, {2, 0.50, tf("15m")}, // Drivers on 15m macro
				{3, 0.10, tf("5m")},  // RSI Filter on 5m (veto 90%)
				{7, 1.50, tf("5m")},  // Candle Booster on 5m
				{7, 1.30, tf("1m")},  // Candle Booster on 1m (entry precision)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "35"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
			Symbols: []symSeed{
				{"DOGEUSDT"}, {"SHIBUSDT"}, {"FLOKIUSDT"}, {"BONKUSDT"},
			},
		},
		{
			Name:       "Pure Trend Rider",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"5m", 0.70}, {"15m", 0.30}},
			Indicators: []indSeed{
				{1, 0.90, nil}, {2, 0.10, nil}, // Drivers on ALL TFs
				{8, 0.50, tf("5m")},  // ATR Filter on 5m
				{9, 1.80, tf("15m")}, // Trend Booster on 15m
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "10"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.8"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
			Symbols: []symSeed{
				{"SOLUSDT"}, {"APTUSDT"}, {"INJUSDT"}, {"JUPUSDT"},
			},
		},
		{
			Name:       "Stoch Pullback Sniper",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.50}, {"15m", 0.30}},
			Indicators: []indSeed{
				{1, 0.60, tf("15m")}, {2, 0.40, tf("15m")}, // Drivers on 15m
				{4, 0.20, tf("5m")},  // Stoch Filter on 5m (veto 80%)
				{6, 1.30, tf("5m")},  // Volume Booster on 5m
				{7, 1.30, tf("1m")},  // Candle Booster on 1m (entry precision)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "10"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
			Symbols: []symSeed{
				{"ETHUSDT"}, {"LINKUSDT"}, {"NEARUSDT"}, {"AVAXUSDT"},
			},
		},

		// ==========================================
		// DAY TRADING STRATEGIES (15m)
		// ==========================================

		{
			Name:       "Day Trading Pro",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.50}, {"1h", 0.30}},
			Indicators: []indSeed{
				{1, 0.60, nil}, {2, 0.40, nil}, // Drivers on ALL TFs
				{3, 0.50, tf("15m")}, {4, 0.50, tf("15m")}, // Filters on 15m
				{5, 0.50, tf("15m")}, {8, 0.80, tf("1h")},  // BB on 15m, ATR on 1h
				{6, 1.20, tf("15m")}, {7, 1.20, tf("5m")},  // Boosters
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
			Symbols: []symSeed{
				{"ETHUSDT"}, {"SOLUSDT"}, {"LINKUSDT"}, {"AVAXUSDT"}, {"NEARUSDT"},
			},
		},
		{
			Name:       "Perfect Storm (Anti-Loss)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.50}, {"1h", 0.30}, {"4h", 0.20}},
			Indicators: []indSeed{
				{1, 0.50, tf("4h")}, {2, 0.50, tf("4h")},               // Drivers only on 4h (macro)
				{3, 0.10, tf("15m")}, {4, 0.10, tf("15m")}, {5, 0.10, tf("15m")}, // Ultra tight filters on 15m
				{6, 1.50, tf("15m")}, {7, 1.50, tf("15m")}, {9, 1.50, tf("1h")},  // Boosters
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "35"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "6"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"BNBUSDT"},
			},
		},
		{
			Name:       "Volatility Hunter",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.60}, {"1h", 0.20}},
			Indicators: []indSeed{
				{1, 0.70, tf("1h")}, {2, 0.30, tf("1h")}, // Drivers on 1h
				{6, 1.80, tf("15m")},  // Volume Booster on 15m
				{8, 0.20, tf("15m")},  // ATR Filter on 15m (no flat market)
				{7, 1.50, tf("5m")},   // Candle Booster on 5m (entry precision)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "6"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.8"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
			Symbols: []symSeed{
				{"SUIUSDT"}, {"APTUSDT"}, {"SEIUSDT"}, {"JUPUSDT"},
			},
		},
		{
			Name:       "BB Squeeze Master",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.60}, {"1h", 0.40}},
			Indicators: []indSeed{
				{1, 0.80, tf("1h")}, {2, 0.20, tf("1h")}, // Drivers on 1h
				{5, 0.10, tf("15m")}, // BB Filter on 15m (veto 90%)
				{6, 1.50, tf("15m")}, // Volume Booster on 15m
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "4"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
			Symbols: []symSeed{
				{"ETHUSDT"}, {"SOLUSDT"}, {"INJUSDT"}, {"RNDRUSDT"},
			},
		},
		{
			Name:       "Price Action Purity",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.80}, {"1h", 0.20}},
			Indicators: []indSeed{
				{1, 0.90, tf("1h")}, {2, 0.10, tf("1h")}, // Drivers on 1h
				{7, 2.00, tf("15m")}, // Candle Booster on 15m (2x)
				{8, 0.60, tf("1h")},  // ATR Filter on 1h (avoid flat market)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "45"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"LINKUSDT"}, {"AAVEUSDT"},
			},
		},
		{
			Name:       "Extreme Scalper (10x Leverage)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"1m", 0.10}, {"5m", 0.40}, {"15m", 0.50}},
			Indicators: []indSeed{
				{1, 0.50, nil}, {2, 0.50, nil}, // Drivers on ALL TFs
				{5, 0.20, tf("5m")},  // BB Filter on 5m
				{6, 1.80, tf("5m")},  // Volume Booster on 5m
				{7, 1.50, tf("1m")},  // Candle Booster on 1m (scalp precision)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "15"},
				{"MAX_DAILY_LOSS_PERCENT", "0.08"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
			Symbols: []symSeed{
				{"PEPEUSDT"}, {"WIFUSDT"}, {"DOGEUSDT"}, {"BONKUSDT"},
			},
		},

		// ==========================================
		// SWING & TREND CONTINUATION (1h)
		// ==========================================

		{
			Name:       "Steady Swing Trend",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.70, tf("4h")}, {2, 0.30, tf("4h")}, // Drivers on 4h
				{8, 0.60, tf("1h")},  // ATR Filter on 1h
				{9, 1.50, tf("4h")},  // Trend Booster on 4h
				{6, 1.30, tf("15m")}, // Volume Booster on 15m (entry confirmation)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.5"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "12"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"FETUSDT"}, {"TAOUSDT"},
			},
		},
		{
			Name:       "Deep Valley Catcher",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.50, tf("4h")}, {2, 0.50, tf("4h")}, // Drivers on 4h
				{3, 0.20, tf("1h")},  // RSI Filter on 1h (veto 80%)
				{7, 1.50, tf("15m")}, // Candle Booster on 15m (reversal precision)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "2"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "3.0"}, {"RISK_REWARD_TARGET", "6.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "24"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"SOLUSDT"}, {"AVAXUSDT"},
			},
		},
		{
			Name:       "Pure Institutional MA",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.60}, {"4h", 0.40}},
			Indicators: []indSeed{
				{1, 1.00, tf("4h")},  // Solo MA Driver on 4h
				{6, 1.50, tf("1h")},  // Volume Booster on 1h
				{9, 1.50, tf("4h")},  // Trend Booster on 4h (alignment bonus)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "8"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"BNBUSDT"},
			},
		},
		{
			Name:       "Safe Hodl Entry",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.40}, {"4h", 0.40}, {"1d", 0.20}},
			Indicators: []indSeed{
				{1, 0.60, tf("1d")}, {2, 0.40, tf("1d")}, // Drivers on Daily
				{3, 0.60, tf("4h")}, {4, 0.60, tf("4h")}, // Soft Filters on 4h
				{9, 1.20, tf("1d")},  // Trend Booster on Daily
				{6, 1.20, tf("1h")},  // Volume Booster on 1h (momentum confirmation)
				{5, 0.50, tf("1h")},  // BB Filter on 1h (band protection)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "1"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "5.0"},
				{"RISK_ENTRY_BUFFER", "0.008"}, {"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "48"},
			},
			Symbols: []symSeed{
				{"BTCUSDT"}, {"ETHUSDT"}, {"XRPUSDT"}, {"ADAUSDT"},
			},
		},
		{
			Name:       "MACD Cross Confirm",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.70}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.20, tf("4h")}, {2, 0.80, tf("1h")}, // MA macro on 4h, MACD momentum on 1h
				{8, 0.60, tf("1h")}, // ATR Filter on 1h
				{6, 1.30, tf("1h")}, // Volume Booster on 1h
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "4"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
			Symbols: []symSeed{
				{"LINKUSDT"}, {"INJUSDT"}, {"RNDRUSDT"}, {"FETUSDT"},
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
					StrategyID:    strategy.ID,
					IndicatorID:   ind.ID,
					Weight:        ind.Weight,
					TimeframeName: ind.TF,
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

			for _, sym := range s.Symbols {
				if err := db.Create(&models.StrategySymbol{
					StrategyID: strategy.ID,
					Symbol:     sym.Symbol,
					IsActive:   true,
				}).Error; err != nil {
					log.Printf("  Failed to create Symbol %s: %v", sym.Symbol, err)
				}
			}

			fmt.Printf("✓ Created strategy: %s (Primary: %s)\n", s.Name, s.PrimaryTF)
		} else if result.Error != nil {
			log.Printf("Failed to check strategy '%s': %v", s.Name, result.Error)
		} else {
			fmt.Printf("⊘ Strategy already exists: %s\n", s.Name)
		}
	}

	// Active one by default
	db.Model(&models.Strategy{}).Where("strategy_name = ?", "Day Trading Pro").Update("is_active", true)

	fmt.Println("✓ Updated strategy IsActive (Day Trading Pro = true)")
}
