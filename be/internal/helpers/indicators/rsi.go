package indicators

import (
	"fmt"
	"math"

	"github.com/reshap/trading-bot/internal/config"
)

// RSIParameters holds the configuration for RSI calculation
type RSIParameters struct {
	Period     int
	Overbought float64
	Oversold   float64
}

// DefaultRSIParameters returns default RSI parameters (period=14, overbought=70, oversold=30)
// Deprecated: Use ParseRSIConfig instead
func DefaultRSIParameters() RSIParameters {
	return RSIParameters{
		Period:     14,
		Overbought: 70,
		Oversold:   30,
	}
}

// ParseRSIConfig parses config into RSIParameters
func ParseRSIConfig(cfg *config.Config) RSIParameters {
	return RSIParameters{
		Period:     cfg.INDICATORS.RSI.PERIOD,
		Overbought: cfg.INDICATORS.RSI.OVERBOUGHT,
		Oversold:   cfg.INDICATORS.RSI.OVERSOLD,
	}
}

// RSIResult holds the RSI analysis result
type RSIResult struct {
	Signal  int     // Score from -100 to 100
	Value   float64 // Current RSI value
	Zone    string  // OVERBOUGHT, OVERSOLD, BULLISH, BEARISH, NEUTRAL
	Details []string
}

// CalculateRSI calculates Relative Strength Index
// Returns RSI values for the given closes data
func CalculateRSI(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return []float64{}
	}

	gains := make([]float64, 0, len(closes)-1)
	losses := make([]float64, 0, len(closes)-1)

	// Calculate gains and losses
	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, math.Abs(change))
		}
	}

	// Calculate initial average gain and loss
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate first RSI
	rsi := make([]float64, 0, len(gains)-period+1)
	if avgLoss == 0 {
		rsi = append(rsi, 100.0)
	} else {
		rs := avgGain / avgLoss
		rsi = append(rsi, 100-100/(1+rs))
	}

	// Calculate remaining RSI values using smoothed averages
	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi = append(rsi, 100.0)
		} else {
			rs := avgGain / avgLoss
			rsi = append(rsi, 100-100/(1+rs))
		}
	}

	return rsi
}

// AnalyzeRSI analyzes RSI signal with scoring system
// Returns signal score (-100 to 100), current value, zone, and details
func AnalyzeRSI(closes []float64) RSIResult {
	return AnalyzeRSIWithParams(closes, DefaultRSIParameters())
}

// AnalyzeRSIWithConfig analyzes RSI signal using config parameters
func AnalyzeRSIWithConfig(closes []float64, cfg *config.Config) RSIResult {
	params := ParseRSIConfig(cfg)
	return AnalyzeRSIWithParams(closes, params)
}

// AnalyzeRSIWithParams analyzes RSI signal with custom parameters
func AnalyzeRSIWithParams(closes []float64, params RSIParameters) RSIResult {
	rsi := CalculateRSI(closes, params.Period)

	if len(rsi) < 5 {
		return RSIResult{
			Signal:  0,
			Value:   0,
			Zone:    "NEUTRAL",
			Details: []string{"Insufficient data"},
		}
	}

	current := rsi[len(rsi)-1]
	prev := rsi[len(rsi)-2]

	signal := 0
	details := make([]string, 0)
	var zone string

	// Determine zone and base signal
	if current >= params.Overbought {
		zone = "OVERBOUGHT"
		signal -= 60
		details = append(details, fmt.Sprintf("RSI %.1f Overbought", current))
	} else if current <= params.Oversold {
		zone = "OVERSOLD"
		signal += 60
		details = append(details, fmt.Sprintf("RSI %.1f Oversold", current))
	} else if current > 50 {
		zone = "BULLISH"
		signal += 40
		details = append(details, fmt.Sprintf("RSI %.1f Bullish", current))
	} else {
		zone = "BEARISH"
		signal -= 40
		details = append(details, fmt.Sprintf("RSI %.1f Bearish", current))
	}

	// Check for 50-line cross
	if prev < 50 && current >= 50 {
		signal += 40
		details = append(details, "RSI crossed above 50")
	} else if prev > 50 && current <= 50 {
		signal -= 40
		details = append(details, "RSI crossed below 50")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return RSIResult{
		Signal:  clampedSignal,
		Value:   current,
		Zone:    zone,
		Details: details,
	}
}
