package indicators

import (
	"fmt"
	"math"

	"github.com/reshap/trading-bot/internal/config"
)

// ATRParameters holds the configuration for ATR calculation
type ATRParameters struct {
	Period int
}

// DefaultATRParameters returns default ATR parameters (period=14)
// Deprecated: Use ParseATRConfig instead
func DefaultATRParameters() ATRParameters {
	return ATRParameters{
		Period: 14,
	}
}

// ParseATRConfig parses config into ATRParameters
func ParseATRConfig(cfg *config.Config) ATRParameters {
	return ATRParameters{
		Period: cfg.INDICATORS.ATR.PERIOD,
	}
}

// ATRResult holds the ATR analysis result
type ATRResult struct {
	ATR        float64 // Current ATR value
	ATRPercent float64 // ATR as percentage of price
	Volatility string  // HIGH, NORMAL, LOW
	ATRRatio   float64 // Current ATR / Average ATR
	Details    []string
	Signal     int     // Score from -100 to 100
}

// CalculateATR calculates Average True Range
// Returns ATR values for the given OHLC data
func CalculateATR(ohlcData []OHLCData, period int) []float64 {
	if len(ohlcData) < period+1 {
		return []float64{}
	}

	// Calculate True Range
	tr := make([]float64, 0, len(ohlcData)-1)
	for i := 1; i < len(ohlcData); i++ {
		highLow := ohlcData[i].High - ohlcData[i].Low
		highClose := math.Abs(ohlcData[i].High - ohlcData[i-1].Close)
		lowClose := math.Abs(ohlcData[i].Low - ohlcData[i-1].Close)

		max := highLow
		if highClose > max {
			max = highClose
		}
		if lowClose > max {
			max = lowClose
		}
		tr = append(tr, max)
	}

	// Calculate initial ATR (SMA of first TR values)
	atr := 0.0
	for i := 0; i < period; i++ {
		atr += tr[i]
	}
	atr /= float64(period)

	// Calculate ATR values using smoothed average
	atrValues := make([]float64, 0, len(tr)-period+1)
	atrValues = append(atrValues, atr)

	for i := period; i < len(tr); i++ {
		atr = (atr*float64(period-1) + tr[i]) / float64(period)
		atrValues = append(atrValues, atr)
	}

	return atrValues
}

// AnalyzeATR analyzes ATR volatility
func AnalyzeATR(ohlcData []OHLCData) ATRResult {
	return AnalyzeATRWithParams(ohlcData, DefaultATRParameters())
}

// AnalyzeATRWithConfig analyzes ATR volatility using config parameters
func AnalyzeATRWithConfig(ohlcData []OHLCData, cfg *config.Config) ATRResult {
	params := ParseATRConfig(cfg)
	return AnalyzeATRWithParams(ohlcData, params)
}

// AnalyzeATRWithParams analyzes ATR volatility with custom parameters
func AnalyzeATRWithParams(ohlcData []OHLCData, params ATRParameters) ATRResult {
	atrValues := CalculateATR(ohlcData, params.Period)

	if len(atrValues) < 20 {
		return ATRResult{
			ATR:        0,
			ATRPercent: 0,
			Volatility: "UNKNOWN",
			ATRRatio:   0,
			Details:    []string{"Insufficient data"},
		}
	}

	current := atrValues[len(atrValues)-1]
	price := ohlcData[len(ohlcData)-1].Close

	// Calculate ATR as percentage of price
	percent := (current / price) * 100.0

	// Calculate average ATR of last 20 periods
	avg := 0.0
	for i := len(atrValues) - 20; i < len(atrValues); i++ {
		avg += atrValues[i]
	}
	avg /= 20.0

	// Calculate ratio of current ATR to average
	ratio := current / avg

	var volatility string
	details := make([]string, 0)

	// Determine volatility level
	if ratio >= 1.5 {
		volatility = "HIGH"
		details = append(details, fmt.Sprintf("ATR %.2f%% High", percent))
	} else if ratio <= 0.7 {
		volatility = "LOW"
		details = append(details, fmt.Sprintf("ATR %.2f%% Low", percent))
	} else {
		volatility = "NORMAL"
		details = append(details, fmt.Sprintf("ATR %.2f%%", percent))
	}

	// Calculate directional signal based on ATR trend
	signal := 0

	// ATR rising = increasing volatility (typically bearish for markets)
	if ratio >= 1.5 {
		signal -= 50 // High volatility = fear = bearish
	} else if ratio <= 0.7 {
		signal += 50 // Low volatility = calm = bullish
	}

	// ATR trend direction (comparing to 5 periods ago)
	if len(atrValues) >= 5 {
		prev5 := atrValues[len(atrValues)-5]
		if current > prev5 {
			signal -= 50 // ATR rising
		} else {
			signal += 50 // ATR falling
		}
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return ATRResult{
		ATR:        current,
		ATRPercent: percent,
		Volatility: volatility,
		ATRRatio:   ratio,
		Details:    details,
		Signal:     clampedSignal,
	}
}
