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
		// ==========================================
		// MICRO & SCALEPING STRATEGIES
		// ==========================================

		{
			Name:       "Micro Breakout Sniper (Agresif)",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.30}, {"5m", 0.50}, {"15m", 0.20}},
			Indicators: []indSeed{
				{1, 0.70}, {2, 0.30}, // Driver: Fokus ke trend MA
				{5, 0.20},            // Filter: Bollinger Squeeze (Hanya eksekusi kalau BB meledak, jika datar diskon 80%)
				{6, 2.00},            // Booster: Volume sangat penting, boost 2x lipat
				{7, 1.30},            // Booster: Pola Candle
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "12"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"}, // RR kecil karena scalping
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"}, // Murni hajar market
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},
		{
			Name:       "RSI Reversal Catcher",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.60}, {"15m", 0.20}},
			Indicators: []indSeed{
				{1, 0.50}, {2, 0.50}, // Driver: Seimbang
				{3, 0.10},            // Filter: RSI Ekstrem (Veto 90% skor jika RSI tidak sejalan)
				{7, 1.50},            // Booster: Butuh Hammer / Engulfing Candle
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "35"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"}, // Antri limit order di ekor candle
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},
		{
			Name:       "Pure Trend Rider",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"5m", 0.70}, {"15m", 0.30}},
			Indicators: []indSeed{
				{1, 0.90}, {2, 0.10}, // Driver: Murni kawal Moving Average
				{8, 0.50},            // Filter: ATR mencegah sideway
				{9, 1.80},            // Booster: Trend Bonus didongkrak 180%
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "10"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.8"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},
		{
			Name:       "Stoch Pullback Sniper",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.50}, {"15m", 0.30}},
			Indicators: []indSeed{
				{1, 0.60}, {2, 0.40}, // Driver
				{4, 0.20},            // Filter: Stochastic ketat (Veto 80%)
				{6, 1.30},            // Booster: Volume konfirmasi masuk
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "10"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},

		// ==========================================
		// DAY TRADING STRATEGIES
		// ==========================================

		{
			Name:       "Day Trading Pro",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.50}, {"1h", 0.30}},
			Indicators: []indSeed{
				{1, 0.60}, {2, 0.40}, // Driver
				{3, 0.50}, {4, 0.50}, {5, 0.50}, {8, 0.80}, // Filter: Semua dijaga seimbang 50%
				{6, 1.20}, {7, 1.20}, // Booster: Konfirmasi standar
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},
		{
			Name:       "Perfect Storm (Anti-Loss)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.50}, {"1h", 0.30}, {"4h", 0.20}},
			Indicators: []indSeed{
				{1, 0.50}, {2, 0.50}, // Driver
				{3, 0.10}, {4, 0.10}, {5, 0.10}, // Filter: Ultra ketat, sedikit saja melenceng langsung veto 90%
				{6, 1.50}, {7, 1.50}, {9, 1.50}, // Booster: Wajib ada ledakan momentum
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "35"}, {"MAX_DAILY_TRADES", "3"}, // Jarang muncul, tapi pasti akurat
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.15"}, // Karena jarang, sekali pasang lumayan besar
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "6"},
			},
		},
		{
			Name:       "Volatility Hunter",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.60}, {"1h", 0.20}},
			Indicators: []indSeed{
				{1, 0.70}, {2, 0.30}, // Driver
				{6, 1.80},            // Booster: Volume Extreme
				{8, 0.20},            // Filter: ATR ketat (Tidak trading saat sepi)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "6"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.8"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"}, // Hajar saat market ramai
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},
		{
			Name:       "BB Squeeze Master",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.60}, {"1h", 0.40}},
			Indicators: []indSeed{
				{1, 0.80}, {2, 0.20}, // Driver
				{5, 0.10},            // Filter: Veto 90% bila Bollinger tidak pecah
				{6, 1.50},            // Booster: Volume ledakan
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "4"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},
		{
			Name:       "Price Action Purity",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"15m", 0.80}, {"1h", 0.20}},
			Indicators: []indSeed{
				{1, 0.90}, {2, 0.10}, // Driver
				{7, 2.00},            // Booster: Hanya percaya Candle (Bonus 2x lipat)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "45"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ==========================================
		// SWING & TREND CONTINUATION
		// ==========================================

		{
			Name:       "Steady Swing Trend",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.70}, {2, 0.30}, // Driver
				{8, 0.60},            // Filter: Minimal pergerakan
				{9, 1.50},            // Booster: Trend mutlak
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.5"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "12"},
			},
		},
		{
			Name:       "Deep Valley Catcher",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.50}, {2, 0.50}, // Driver
				{3, 0.20},            // Filter: Beli hanya saat oversold ekstrem (veto 80%)
				{7, 1.50},            // Booster: Konfirmasi reversal pinbar
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "2"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "3.0"}, {"RISK_REWARD_TARGET", "6.0"}, // Reward wajib huge!
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "24"},
			},
		},
		{
			Name:       "Pure Institutional MA",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.60}, {"4h", 0.40}},
			Indicators: []indSeed{
				{1, 1.00},            // Driver tunggal
				{6, 1.50},            // Booster: Spike volume whale
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "8"},
			},
		},
		{
			Name:       "Safe Hodl Entry",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.40}, {"4h", 0.40}, {"1d", 0.20}},
			Indicators: []indSeed{
				{1, 0.60}, {2, 0.40}, // Driver
				{3, 0.60}, {4, 0.60}, // Filter: Lembut
				{9, 1.20},            // Booster: Lembut
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "1"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "5.0"},
				{"RISK_ENTRY_BUFFER", "0.008"}, {"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "48"},
			},
		},
		{
			Name:       "MACD Cross Confirm",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"1h", 0.70}, {"4h", 0.30}},
			Indicators: []indSeed{
				{1, 0.20}, {2, 0.80}, // Driver: Fokus murni momentum MACD
				{8, 0.60},            // Filter ATR
				{6, 1.30},            // Booster Vol
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "4"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "true"}, // Momen silang ambil market execution
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},
		{
			Name:       "Extreme Scalper (10x Leverage)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"1m", 0.10}, {"5m", 0.40}, {"15m", 0.50}},
			Indicators: []indSeed{
				{1, 0.50}, {2, 0.50}, // Driver
				{5, 0.20},            // Filter BB
				{6, 1.80},            // Volume Wajib
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "40"}, {"MAX_DAILY_TRADES", "15"},
				{"MAX_DAILY_LOSS_PERCENT", "0.08"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "1"},
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

			// NOTE: indBenchmark in Deep Valley Catcher typo handling
			// Wait, the struct definition used earlier was indSeed, let's make sure it's valid Go code
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

	// Active one by default
	db.Model(&models.Strategy{}).Where("strategy_name = ?", "Day Trading Pro").Update("is_active", true)

	fmt.Println("✓ Updated strategy IsActive (Day Trading Pro = true)")
}
