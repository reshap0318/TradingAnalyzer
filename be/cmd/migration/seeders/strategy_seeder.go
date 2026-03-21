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

		{
			Name:      "Day Trading Pro",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"15m", 0.50},
				{"1h", 0.30},
				{"4h", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.30},
				{2, 0.22},
				{3, 0.13},
				{4, 0.10},
				{5, 0.10},
				{6, 0.05},
				{7, 0.05},
				{8, 0.05},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ──────────────────────────────────────────────
		// 2.a Day Trading Pro (MICRO ACCOUNT $37 EDITION)
		// Fokus: Limit order di S/R (hemat fee maker), leverage agak tinggi buat lolos min. notional order
		// ──────────────────────────────────────────────
		{
			Name:       "Day Trading Pro (Micro $37)",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.10}, {"15m", 0.40}, {"1h", 0.30}, {"4h", 0.20}},
			Indicators: []indSeed{{1, 0.25}, {2, 0.20}, {3, 0.10}, {4, 0.10}, {5, 0.10}, {6, 0.15}, {7, 0.05}, {8, 0.05}},
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
		// 3. Swing Trader - Medium term, ride the wave
		// Recomendation: BTC, ETH, SOL, BNB, LINK, AVAX
		// ──────────────────────────────────────────────
		{
			Name:       "Swing Trader",
			PrimaryTF:  "4h",
			Timeframes: []tfSeed{{"1h", 0.20}, {"4h", 0.50}, {"1d", 0.30}},
			Indicators: []indSeed{{1, 0.30}, {2, 0.20}, {3, 0.10}, {5, 0.10}, {6, 0.10}, {8, 0.10}, {9, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.05"}, // Modal normal, size 5% udah cukup
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"}, // Nunggu pullback di TF besar
				{"ORDER_EXPIRATION_HOURS", "24"},
			},
		},

		// ──────────────────────────────────────────────
		// 4. Breakout Hunter - High volatility, volume focused
		// Recomendation: DOGE, SHIB, PEPE, WIF, BONK, FLOKI
		// ──────────────────────────────────────────────
		{
			Name:       "Breakout Hunter",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.50}, {"1h", 0.30}},
			Indicators: []indSeed{{1, 0.10}, {2, 0.10}, {5, 0.30}, {6, 0.30}, {7, 0.10}, {8, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"}, // Hajar market karena breakout biasanya kencang
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 5. Trend Follower - Mid to long term momentum
		// Recomendation: INJ, FET, RNDR, TIA, TAO, STX
		// ──────────────────────────────────────────────
		{
			Name:       "Trend Follower",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{{1, 0.35}, {2, 0.25}, {6, 0.15}, {8, 0.10}, {9, 0.15}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "60"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Sabar nunggu harga retrace ke MA
				{"ORDER_EXPIRATION_HOURS", "8"},
			},
		},

		// ──────────────────────────────────────────────
		// 6. Mean Reversion - Ranging market specialist (Sideways)
		// Recomendation: ADA, XRP, DOT, MATIC (POL), LTC, TRX
		// ──────────────────────────────────────────────
		{
			Name:       "Mean Reversion",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.60}, {"1h", 0.20}},
			Indicators: []indSeed{{3, 0.35}, {4, 0.25}, {5, 0.30}, {7, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "12"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.2"}, {"RISK_REWARD_TARGET", "2.0"},
				{"RISK_ENTRY_BUFFER", "0.001"}, {"MAX_POSITION_SIZE", "0.03"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Mutlak false! Beli di support, jual di resis
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 7. Momentum Rider - Catching the fast moves
		// Recomendation: SEI, SUI, APT, JUP
		// ──────────────────────────────────────────────
		{
			Name:       "Momentum Rider",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.20}, {"5m", 0.50}, {"15m", 0.30}},
			Indicators: []indSeed{{2, 0.30}, {3, 0.20}, {6, 0.30}, {7, 0.10}, {9, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "15"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.2"}, {"RISK_REWARD_TARGET", "2.0"},
				{"RISK_ENTRY_BUFFER", "0.001"}, {"MAX_POSITION_SIZE", "0.02"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"}, // Ketinggalan momentum kalau nunggu limit
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},

		// ──────────────────────────────────────────────
		// 8. Volatility Surfer - ATR & Volume based
		// Recomendation: GALA, SAND, MANA, AXS, NEAR, IMX
		// ──────────────────────────────────────────────
		{
			Name:       "Volatility Surfer",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.30}, {"1h", 0.50}, {"4h", 0.20}},
			Indicators: []indSeed{{5, 0.20}, {6, 0.25}, {8, 0.40}, {9, 0.15}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "6"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.05"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "4"},
			},
		},

		// ──────────────────────────────────────────────
		// 9. Deep Pullback Catcher - Buying the dip
		// Recomendation: ETH, SOL, BNB, AVAX, LINK, RNDR
		// ──────────────────────────────────────────────
		{
			Name:       "Deep Pullback",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.40}, {"4h", 0.40}},
			Indicators: []indSeed{{1, 0.20}, {3, 0.30}, {4, 0.20}, {5, 0.20}, {7, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "70"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"}, // Nunggu di support paling bawah
				{"ORDER_EXPIRATION_HOURS", "12"},
			},
		},

		// ──────────────────────────────────────────────
		// 10. Conservative Hodler - Very safe, spot-like
		// Recomendation: BTC, ETH, BNB, SOL, XRP
		// ──────────────────────────────────────────────
		{
			Name:       "Conservative Hodler",
			PrimaryTF:  "1d",
			Timeframes: []tfSeed{{"4h", 0.20}, {"1d", 0.60}, {"1w", 0.20}},
			Indicators: []indSeed{{1, 0.40}, {2, 0.20}, {3, 0.10}, {6, 0.10}, {9, 0.20}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "75"}, {"MAX_DAILY_TRADES", "2"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "5.0"},
				{"RISK_ENTRY_BUFFER", "0.01"}, {"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "1"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "48"},
			},
		},

		// ──────────────────────────────────────────────
		// 11. Altcoin Degen - High risk, high reward (Meme coins)
		// Recomendation: BOME, MEW, MYRO, POPCAT, SLERF, DEGEN
		// ──────────────────────────────────────────────
		{
			Name:       "Altcoin Degen",
			PrimaryTF:  "5m",
			Timeframes: []tfSeed{{"1m", 0.30}, {"5m", 0.50}, {"15m", 0.20}},
			Indicators: []indSeed{{2, 0.20}, {3, 0.20}, {6, 0.30}, {8, 0.20}, {9, 0.10}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "45"}, {"MAX_DAILY_TRADES", "15"},
				{"MAX_DAILY_LOSS_PERCENT", "0.08"}, {"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.02"}, // Size super kecil karena high risk
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},

		// ──────────────────────────────────────────────
		// 12. Candle Sniper - Price action pure focus
		// Recomendation: LDO, SNX, CRV, MKR, AAVE, UNI
		// ──────────────────────────────────────────────
		{
			Name:       "Candle Sniper",
			PrimaryTF:  "4h",
			Timeframes: []tfSeed{{"1h", 0.30}, {"4h", 0.60}, {"1d", 0.10}},
			Indicators: []indSeed{{1, 0.20}, {5, 0.20}, {6, 0.15}, {7, 0.45}}, // Pola candle dominan
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "70"}, {"MAX_DAILY_TRADES", "4"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.06"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Limit order pas candle konfirmasi
				{"ORDER_EXPIRATION_HOURS", "8"},
			},
		},

		// ──────────────────────────────────────────────
		// 13. Micro Trend - Catching intra-day waves
		// Recomendation: SUI, SEI, APT, FTM, MNT, TIA
		// ──────────────────────────────────────────────
		{
			Name:       "Micro Trend",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.25}, {"15m", 0.50}, {"1h", 0.25}},
			Indicators: []indSeed{{1, 0.30}, {2, 0.20}, {3, 0.15}, {6, 0.20}, {9, 0.15}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "55"}, {"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.04"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"}, // Agresif tangkap wave kecil
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 14. Weekend Warrior - Low volume / algorithmic trading
		// Recomendation: LTC, BCH, ETC, XMR, ZEC, DOGE
		// ──────────────────────────────────────────────
		{
			Name:       "Weekend Warrior",
			PrimaryTF:  "1h",
			Timeframes: []tfSeed{{"15m", 0.20}, {"1h", 0.50}, {"4h", 0.30}},
			Indicators: []indSeed{{1, 0.20}, {3, 0.30}, {4, 0.20}, {5, 0.30}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"}, {"MAX_DAILY_TRADES", "6"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"}, {"MAX_POSITION_SIZE", "0.04"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Market sepi mending nunggu limit
				{"ORDER_EXPIRATION_HOURS", "8"},
			},
		},

		// ──────────────────────────────────────────────
		// 15. Golden Cross Seeker - Pure Moving Average logic
		// Recomendation: BTC, ETH, BNB, SOL, LINK, AVAX
		// ──────────────────────────────────────────────
		{
			Name:       "Golden Cross Seeker",
			PrimaryTF:  "1d",
			Timeframes: []tfSeed{{"4h", 0.30}, {"1d", 0.70}},
			Indicators: []indSeed{{1, 0.60}, {6, 0.20}, {9, 0.20}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "70"}, {"MAX_DAILY_TRADES", "2"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "true"}, // Begitu cross, hajar 1 market, 1 antri
				{"ORDER_EXPIRATION_HOURS", "48"},
			},
		},

		// ──────────────────────────────────────────────
		// 16. Safe Swing - Low lev, high confirmation
		// Recomendation: BTC, ETH, XRP, ADA, DOT, AVAX, LINK
		// ──────────────────────────────────────────────
		{
			Name:       "Safe Swing",
			PrimaryTF:  "4h",
			Timeframes: []tfSeed{{"1h", 0.20}, {"4h", 0.50}, {"1d", 0.30}},
			Indicators: []indSeed{{1, 0.20}, {2, 0.20}, {3, 0.15}, {6, 0.15}, {7, 0.15}, {9, 0.15}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "70"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.03"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.004"}, {"MAX_POSITION_SIZE", "0.08"},
				{"LEVERAGE", "3"}, {"IS_AGRESSIVE", "false"}, // Wajib limit order
				{"ORDER_EXPIRATION_HOURS", "24"},
			},
		},

		// ──────────────────────────────────────────────
		// 17. The Sniper - One shot one kill
		// Recomendation: BTC, ETH, SOL, BNB, INJ, AVAX
		// ──────────────────────────────────────────────
		{
			Name:       "The Sniper",
			PrimaryTF:  "15m",
			Timeframes: []tfSeed{{"5m", 0.10}, {"15m", 0.60}, {"1h", 0.30}},
			Indicators: []indSeed{{2, 0.20}, {3, 0.20}, {5, 0.20}, {6, 0.20}, {7, 0.20}},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "80"}, {"MAX_DAILY_TRADES", "3"},
				{"MAX_DAILY_LOSS_PERCENT", "0.02"}, {"MAX_DAILY_LOSS_COUNT", "1"}, // Kalah 1x stop trade hari itu
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "4.0"},
				{"RISK_ENTRY_BUFFER", "0.001"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Cuma open limit order di pucuk / lembah
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		{
			Name:      "Micro Scalper (Aggressive)",
			PrimaryTF: "5m",
			Timeframes: []tfSeed{
				{"1m", 0.20}, {"5m", 0.60}, {"15m", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.35}, // MA (Trend)
				{2, 0.30}, // MACD (Momentum)
				{6, 0.20}, // Volume (Penting buat TF kecil)
				{8, 0.15}, // ATR (Buat ukur volatilitas)
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "45"},   // LONGGARIN! Biar skor 45 aja udah berani nembak.
				{"MAX_DAILY_TRADES", "10"}, // Naikin biar target harian tercapai
				{"MAX_DAILY_LOSS_PERCENT", "0.08"},
				{"MAX_DAILY_LOSS_COUNT", "4"},
				{"RISK_REWARD_RATIO", "1.2"}, // Turunin batas R:R karena di TF kecil jaraknya sempit
				{"RISK_REWARD_TARGET", "2.0"},
				{"RISK_ENTRY_BUFFER", "0.001"},
				{"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "10"},
				{"IS_AGRESSIVE", "true"}, // Biar langsung market execution, TF kecil antri limit sering ditinggal
				{"ORDER_EXPIRATION_HOURS", "1"},
			},
		},

		{
			Name:      "Active Day Trader",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.30}, {"15m", 0.50}, {"1h", 0.20},
			},
			Indicators: []indSeed{
				{1, 0.30},
				{2, 0.30},
				{5, 0.20},
				{6, 0.10},
				{8, 0.10},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"},
				{"MAX_DAILY_TRADES", "8"},
				{"MAX_DAILY_LOSS_PERCENT", "0.06"},
				{"MAX_DAILY_LOSS_COUNT", "3"},
				{"RISK_REWARD_RATIO", "1.5"},
				{"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.002"},
				{"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"},
				{"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		{
			Name:      "Sniper 15m (Unleashed)",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.40}, {"15m", 0.60},
			},
			Indicators: []indSeed{
				{1, 0.40},
				{2, 0.40},
				{6, 0.20},
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "50"},
				{"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.20"},
				{"LEVERAGE", "10"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 18. Pure Momentum Breakout (Solusi 1)
		// PURE TREND: 100% Mengikuti tren. Tanpa oscillator agar tidak terkena cancel-out efek overbought.
		// Sangat brutal dalam mencetak Strong Signal di kondisi market hijau/merah tajam.
		// ──────────────────────────────────────────────
		{
			Name:      "Pure Momentum Breakout",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"15m", 0.70}, {"1h", 0.30},
			},
			Indicators: []indSeed{
				{1, 0.50}, // MA (Penentu Arah Utama)
				{2, 0.30}, // MACD (Kekuatan Pendorong)
				{6, 0.10}, // Volume (Konfirmasi Breakout)
				{9, 0.10}, // Trend Bonus
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "3.0"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.15"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "true"},
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 19. Pure Knife Catcher (Solusi 1)
		// PURE REVERSION: 100% mencari pantulan atas/bawah. Tanpa MA agar tidak diblokir oleh bobot trend.
		// Akurasi tinggi menangkap harga oversold yang parah di timeframe kecil.
		// ──────────────────────────────────────────────
		{
			Name:      "Pure Knife Catcher",
			PrimaryTF: "15m",
			Timeframes: []tfSeed{
				{"5m", 0.40}, {"15m", 0.60},
			},
			Indicators: []indSeed{
				{3, 0.40}, // RSI
				{4, 0.40}, // Stochastic
				{5, 0.20}, // Bollinger Bands
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "65"}, {"MAX_DAILY_TRADES", "5"},
				{"MAX_DAILY_LOSS_PERCENT", "0.05"}, {"MAX_DAILY_LOSS_COUNT", "2"},
				{"RISK_REWARD_RATIO", "1.5"}, {"RISK_REWARD_TARGET", "2.5"},
				{"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"}, // Mutlak jaring bawah
				{"ORDER_EXPIRATION_HOURS", "2"},
			},
		},

		// ──────────────────────────────────────────────
		// 20. Steady Swing Trend (Solusi 1)
		// PURE TREND: Mirip Momentum Breakout tapi khusus untuk long-term holding.
		// ──────────────────────────────────────────────
		{
			Name:       "Steady Swing Trend",
			PrimaryTF:  "4h",
			Timeframes: []tfSeed{{"4h", 0.60}, {"1d", 0.40}},
			Indicators: []indSeed{
				{1, 0.60}, // MA Dominan
				{2, 0.20}, // MACD konfirmator
				{8, 0.10}, // ATR Volatilitas
				{9, 0.10}, // Trend Bonus
			},
			MoneyMgmt: []mmSeed{
				{"MIN_CONFIDENCE", "70"}, {"MAX_DAILY_TRADES", "2"},
				{"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "1"},
				{"RISK_REWARD_RATIO", "2.5"}, {"RISK_REWARD_TARGET", "5.0"},
				{"RISK_ENTRY_BUFFER", "0.005"}, {"MAX_POSITION_SIZE", "0.10"},
				{"LEVERAGE", "2"}, {"IS_AGRESSIVE", "false"},
				{"ORDER_EXPIRATION_HOURS", "48"},
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
