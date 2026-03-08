package indicators

import (
	"math"

	"github.com/reshap/trading-bot/internal/config"
)

// BBParameters holds the configuration for Bollinger Bands calculation
type BBParameters struct {
	Period     int
	StdDevMult float64
}

// DefaultBBParameters returns default Bollinger Bands parameters (period=20, stdDev=2)
// Deprecated: Use ParseBollingerConfig instead
func DefaultBBParameters() BBParameters {
	return BBParameters{
		Period:     20,
		StdDevMult: 2.0,
	}
}

// ParseBollingerConfig parses config into BBParameters
func ParseBollingerConfig(cfg *config.Config) BBParameters {
	return BBParameters{
		Period:     cfg.INDICATORS.BOLLINGER.PERIOD,
		StdDevMult: cfg.INDICATORS.BOLLINGER.STD_DEV,
	}
}

// BBValues holds upper, middle, and lower band values
type BBValues struct {
	Upper  float64
	Middle float64
	Lower  float64
}

// BBLines holds all Bollinger Bands line data
type BBLines struct {
	Upper     []float64
	Middle    []float64
	Lower     []float64
	Bandwidth []float64
}

// BBResult holds the Bollinger Bands analysis result
type BBResult struct {
	Signal   int      // Score from -100 to 100
	Position string   // ABOVE_UPPER, BELOW_LOWER, UPPER_HALF, LOWER_HALF, NEUTRAL
	PercentB float64  // Position within bands (0-1)
	Values   BBValues // Current band values
	Details  []string
}

// CalculateBollingerBands calculates Bollinger Bands
// Returns upper, middle, lower bands and bandwidth
func CalculateBollingerBands(closes []float64, period int, stdDevMult float64) BBLines {
	if len(closes) < period {
		return BBLines{
			Upper:     []float64{},
			Middle:    []float64{},
			Lower:     []float64{},
			Bandwidth: []float64{},
		}
	}

	// Calculate middle band (SMA)
	middle := CalculateSMA(closes, period)

	// Calculate upper and lower bands
	upper := make([]float64, 0, len(closes)-period+1)
	lower := make([]float64, 0, len(closes)-period+1)
	bandwidth := make([]float64, 0, len(closes)-period+1)

	for i := period - 1; i < len(closes); i++ {
		// Calculate standard deviation
		avg := middle[i-period+1]
		sumSquares := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := closes[j] - avg
			sumSquares += diff * diff
		}
		stdDev := math.Sqrt(sumSquares / float64(period))

		// Calculate bands
		u := avg + stdDev*stdDevMult
		l := avg - stdDev*stdDevMult
		upper = append(upper, u)
		lower = append(lower, l)

		// Calculate bandwidth percentage
		bw := ((u - l) / avg) * 100.0
		bandwidth = append(bandwidth, bw)
	}

	return BBLines{
		Upper:     upper,
		Middle:    middle,
		Lower:     lower,
		Bandwidth: bandwidth,
	}
}

// AnalyzeBollingerBands analyzes Bollinger Bands signal with scoring system
func AnalyzeBollingerBands(closes []float64) BBResult {
	return AnalyzeBollingerBandsWithParams(closes, DefaultBBParameters())
}

// AnalyzeBollingerBandsWithConfig analyzes Bollinger Bands signal using config parameters
func AnalyzeBollingerBandsWithConfig(closes []float64, cfg *config.Config) BBResult {
	params := ParseBollingerConfig(cfg)
	return AnalyzeBollingerBandsWithParams(closes, params)
}

// AnalyzeBollingerBandsWithParams analyzes Bollinger Bands signal with custom parameters
func AnalyzeBollingerBandsWithParams(closes []float64, params BBParameters) BBResult {
	bb := CalculateBollingerBands(closes, params.Period, params.StdDevMult)

	if len(bb.Upper) < 5 {
		return BBResult{
			Signal:   0,
			Position: "NEUTRAL",
			PercentB: 0,
			Values:   BBValues{},
			Details:  []string{"Insufficient data"},
		}
	}

	price := closes[len(closes)-1]
	prevPrice := closes[len(closes)-2]
	upper := bb.Upper[len(bb.Upper)-1]
	middle := bb.Middle[len(bb.Middle)-1]
	lower := bb.Lower[len(bb.Lower)-1]
	prevLower := bb.Lower[len(bb.Lower)-2]
	prevUpper := bb.Upper[len(bb.Upper)-2]

	// Calculate %B (position within bands)
	percentB := (price - lower) / (upper - lower)

	signal := 0
	details := make([]string, 0)
	var position string

	// Determine position
	if price >= upper {
		position = "ABOVE_UPPER"
		signal -= 45
		details = append(details, "At upper band")
	} else if price <= lower {
		position = "BELOW_LOWER"
		signal += 45
		details = append(details, "At lower band")
	} else if price > middle {
		position = "UPPER_HALF"
		signal += 27
		details = append(details, "Upper half")
	} else {
		position = "LOWER_HALF"
		signal -= 27
		details = append(details, "Lower half")
	}

	// Check for band bounces
	if prevPrice <= prevLower && price > lower {
		signal += 54
		details = append(details, "Bounced from lower")
	} else if prevPrice >= prevUpper && price < upper {
		signal -= 54
		details = append(details, "Rejected from upper")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return BBResult{
		Signal:   clampedSignal,
		Position: position,
		PercentB: percentB,
		Values:   BBValues{Upper: upper, Middle: middle, Lower: lower},
		Details:  details,
	}
}
