package indicators

import (
	"fmt"
	"math"

	"github.com/reshap/trading-bot/internal/config"
)

// StochasticParameters holds the configuration for Stochastic calculation
type StochasticParameters struct {
	KPeriod    int
	DPeriod    int
	Smooth     int
	Overbought float64
	Oversold   float64
}

// DefaultStochasticParameters returns default Stochastic parameters
// Deprecated: Use ParseStochasticConfig instead
func DefaultStochasticParameters() StochasticParameters {
	return StochasticParameters{
		KPeriod:    14,
		DPeriod:    3,
		Smooth:     3,
		Overbought: 80,
		Oversold:   20,
	}
}

// ParseStochasticConfig parses config into StochasticParameters
func ParseStochasticConfig(cfg *config.Config) StochasticParameters {
	return StochasticParameters{
		KPeriod:    cfg.INDICATORS.STOCHASTIC.K_PERIOD,
		DPeriod:    cfg.INDICATORS.STOCHASTIC.D_PERIOD,
		Smooth:     cfg.INDICATORS.STOCHASTIC.SMOOTH,
		Overbought: cfg.INDICATORS.STOCHASTIC.OVERBOUGHT,
		Oversold:   cfg.INDICATORS.STOCHASTIC.OVERSOLD,
	}
}

// StochValues holds %K and %D values
type StochValues struct {
	K float64
	D float64
}

// StochLines holds all Stochastic line data
type StochLines struct {
	K []float64
	D []float64
}

// StochasticResult holds the Stochastic analysis result
type StochasticResult struct {
	Signal  int         // Score from -100 to 100
	Zone    string      // OVERBOUGHT, OVERSOLD, NEUTRAL
	Values  StochValues // Current %K and %D
	Details []string
}

// CalculateStochastic calculates Stochastic Oscillator
// Returns %K and %D lines
func CalculateStochastic(highs, lows, closes []float64, kPeriod, dPeriod, smooth int) StochLines {
	if len(closes) < kPeriod+dPeriod {
		return StochLines{
			K: []float64{},
			D: []float64{},
		}
	}

	// Calculate raw %K
	rawK := make([]float64, 0, len(closes)-kPeriod+1)
	for i := kPeriod - 1; i < len(closes); i++ {
		hh := lows[i] // Initialize with first value
		ll := highs[i]
		for j := i - kPeriod + 1; j <= i; j++ {
			if highs[j] > hh {
				hh = highs[j]
			}
			if lows[j] < ll {
				ll = lows[j]
			}
		}

		if hh == ll {
			rawK = append(rawK, 50.0)
		} else {
			rawK = append(rawK, ((closes[i]-ll)/(hh-ll))*100.0)
		}
	}

	// Calculate smoothed %K
	k := make([]float64, 0, len(rawK)-smooth+1)
	for i := smooth - 1; i < len(rawK); i++ {
		sum := 0.0
		for j := i - smooth + 1; j <= i; j++ {
			sum += rawK[j]
		}
		k = append(k, sum/float64(smooth))
	}

	// Calculate %D (SMA of %K)
	d := make([]float64, 0, len(k)-dPeriod+1)
	for i := dPeriod - 1; i < len(k); i++ {
		sum := 0.0
		for j := i - dPeriod + 1; j <= i; j++ {
			sum += k[j]
		}
		d = append(d, sum/float64(dPeriod))
	}

	return StochLines{K: k, D: d}
}

// AnalyzeStochastic analyzes Stochastic signal with scoring system
func AnalyzeStochastic(ohlcData []OHLCData) StochasticResult {
	return AnalyzeStochasticWithParams(ohlcData, DefaultStochasticParameters())
}

// AnalyzeStochasticWithConfig analyzes Stochastic signal using config parameters
func AnalyzeStochasticWithConfig(ohlcData []OHLCData, cfg *config.Config) StochasticResult {
	params := ParseStochasticConfig(cfg)
	return AnalyzeStochasticWithParams(ohlcData, params)
}

// AnalyzeStochasticWithParams analyzes Stochastic signal with custom parameters
func AnalyzeStochasticWithParams(ohlcData []OHLCData, params StochasticParameters) StochasticResult {
	highs := make([]float64, len(ohlcData))
	lows := make([]float64, len(ohlcData))
	closes := make([]float64, len(ohlcData))
	for i, d := range ohlcData {
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
	}

	stoch := CalculateStochastic(highs, lows, closes, params.KPeriod, params.DPeriod, params.Smooth)

	if len(stoch.K) < 3 || len(stoch.D) < 3 {
		return StochasticResult{
			Signal:  0,
			Zone:    "NEUTRAL",
			Values:  StochValues{},
			Details: []string{"Insufficient data"},
		}
	}

	currK := stoch.K[len(stoch.K)-1]
	currD := stoch.D[len(stoch.D)-1]
	prevK := stoch.K[len(stoch.K)-2]
	prevD := stoch.D[len(stoch.D)-2]

	signal := 0
	details := make([]string, 0)
	var zone string

	// Determine zone
	if currK >= params.Overbought {
		zone = "OVERBOUGHT"
		signal -= 36
		details = append(details, fmt.Sprintf("%%K %.1f Overbought", currK))
	} else if currK <= params.Oversold {
		zone = "OVERSOLD"
		signal += 36
		details = append(details, fmt.Sprintf("%%K %.1f Oversold", currK))
	} else {
		zone = "NEUTRAL"
		details = append(details, fmt.Sprintf("%%K %.1f", currK))
	}

	// Check for K/D cross
	if prevK <= prevD && currK > currD {
		signal += 64
		details = append(details, "%K crossed above %D")
	} else if prevK >= prevD && currK < currD {
		signal -= 64
		details = append(details, "%K crossed below %D")
	} else if currK > currD {
		signal += 27
		details = append(details, "%K above %D")
	} else {
		signal -= 27
		details = append(details, "%K below %D")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return StochasticResult{
		Signal:  clampedSignal,
		Zone:    zone,
		Values:  StochValues{K: currK, D: currD},
		Details: details,
	}
}

// AnalyzeStochasticWithTrendAndParams analyzes Stochastic with trend neutralization and custom parameters
// This matches GitHub's Trend Regime Detection in cryptoDecisionEngine.js
// Parameters:
//   - ohlcData: OHLCV data for Stochastic calculation
//   - maSignal: Pre-calculated MA signal (-100 to 100)
//   - macdSignal: Pre-calculated MACD signal (-100 to 100)
//   - params: Custom Stochastic parameters (K_PERIOD, D_PERIOD, SMOOTH, OVERBOUGHT, OVERSOLD)
//
// Logic:
//   - If strong uptrend (MA+ + MACD+) and Stoch overbought (signal < 0) → neutralize
//   - If strong downtrend (MA- + MACD-) and Stoch oversold (signal > 0) → neutralize
func AnalyzeStochasticWithTrendAndParams(ohlcData []OHLCData, maSignal, macdSignal int, params StochasticParameters) StochasticResult {
	// Get base Stochastic result with custom params
	result := AnalyzeStochasticWithParams(ohlcData, params)

	// Check for strong trend (both MA and MACD aligned) - matching GitHub logic
	isUptrend := maSignal > 0 && macdSignal > 0
	isDowntrend := maSignal < 0 && macdSignal < 0

	// Neutralize overbought in uptrend (matching GitHub: set contribution to 0)
	if isUptrend && result.Signal < 0 {
		result.Signal = 0
		result.Details = append(result.Details, "Overbought Ignored (Strong Uptrend)")
	}

	// Neutralize oversold in downtrend (matching GitHub: set contribution to 0)
	if isDowntrend && result.Signal > 0 {
		result.Signal = 0
		result.Details = append(result.Details, "Oversold Ignored (Strong Downtrend)")
	}

	return result
}
