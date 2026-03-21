package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedIndicators inserts default indicator data
func SeedIndicators(db *gorm.DB) {
	fmt.Println("Seeding indicators...")

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

	count := 0
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
				count++
				fmt.Printf("  ✓ Created indicator: %s (%s) - Weight: %.2f\n", indicatorData.Name, indicatorData.Indicator, indicatorData.Weight)
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
				fmt.Printf("  ✓ Updated indicator: %s (%s) - Weight: %.2f\n", indicatorData.Name, indicatorData.Indicator, indicatorData.Weight)
			}
		}
	}

	fmt.Printf("✓ Seeded/Updated %d indicators\n", count)
}
