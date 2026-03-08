package indicators

import (
	"math"

	"github.com/reshap/trading-bot/internal/config"
)

// VolumeParameters holds the configuration for Volume analysis
type VolumeParameters struct {
	MAPeriod int
}

// DefaultVolumeParameters returns default Volume analysis parameters (MA period=20)
// Deprecated: Use ParseVolumeConfig instead
func DefaultVolumeParameters() VolumeParameters {
	return VolumeParameters{
		MAPeriod: 20,
	}
}

// ParseVolumeConfig parses config into VolumeParameters
func ParseVolumeConfig(cfg *config.Config) VolumeParameters {
	return VolumeParameters{
		MAPeriod: cfg.INDICATORS.VOLUME.MA_PERIOD,
	}
}

// VolumeResult holds the Volume analysis result
type VolumeResult struct {
	Signal        int     // Score from -100 to 100
	Confirmation  bool    // Whether volume confirms price action
	VolumeRatio   float64 // Current volume / Average volume
	CurrentVolume float64
	AvgVolume     float64
	Details       []string
}

// AnalyzeVolume analyzes volume with scoring system
func AnalyzeVolume(ohlcData []OHLCData) VolumeResult {
	return AnalyzeVolumeWithParams(ohlcData, DefaultVolumeParameters())
}

// AnalyzeVolumeWithConfig analyzes volume using config parameters
func AnalyzeVolumeWithConfig(ohlcData []OHLCData, cfg *config.Config) VolumeResult {
	params := ParseVolumeConfig(cfg)
	return AnalyzeVolumeWithParams(ohlcData, params)
}

// AnalyzeVolumeWithParams analyzes volume with custom parameters
func AnalyzeVolumeWithParams(ohlcData []OHLCData, params VolumeParameters) VolumeResult {
	if len(ohlcData) < params.MAPeriod+5 {
		return VolumeResult{
			Signal:       0,
			Confirmation: false,
			VolumeRatio:  0,
			Details:      []string{"Insufficient data"},
		}
	}

	// Extract volumes and closes
	volumes := make([]float64, len(ohlcData))
	closes := make([]float64, len(ohlcData))
	for i, d := range ohlcData {
		volumes[i] = d.Volume
		closes[i] = d.Close
	}

	currVol := volumes[len(volumes)-1]
	currClose := closes[len(closes)-1]
	prevClose := closes[len(closes)-2]

	// Calculate volume SMA
	avgVol := CalculateSMA(volumes, params.MAPeriod)
	avg := avgVol[len(avgVol)-1]

	// Calculate volume ratio
	ratio := currVol / avg

	signal := 0
	details := make([]string, 0)
	confirmation := false

	// Analyze volume level and direction
	if ratio >= 2.0 {
		// Very high volume (> 200% of average)
		confirmation = true
		if currClose > prevClose {
			signal += 100
			details = append(details, "Very high volume up")
		} else {
			signal -= 100
			details = append(details, "Very high volume down")
		}
	} else if ratio >= 1.5 {
		// High volume (> 150% of average)
		confirmation = true
		if currClose > prevClose {
			signal += 62
			details = append(details, "High volume up")
		} else {
			signal -= 62
			details = append(details, "High volume down")
		}
	} else if ratio >= 1.0 {
		// Normal or above average volume
		if currClose > prevClose {
			signal += 25
		} else {
			signal -= 25
		}
		details = append(details, "Normal volume")
	} else {
		// Low volume
		details = append(details, "Low volume")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return VolumeResult{
		Signal:        clampedSignal,
		Confirmation:  confirmation,
		VolumeRatio:   ratio,
		CurrentVolume: currVol,
		AvgVolume:     avg,
		Details:       details,
	}
}

// VolumeSignal represents the volume signal strength
type VolumeSignal struct {
	Ratio      float64
	Level      string // VERY_HIGH, HIGH, NORMAL, LOW
	Direction  string // UP, DOWN
	Strength   int    // Signal strength 0-100
}

// GetVolumeSignal returns detailed volume signal information
func GetVolumeSignal(ohlcData []OHLCData) VolumeSignal {
	params := DefaultVolumeParameters()

	if len(ohlcData) < params.MAPeriod+5 {
		return VolumeSignal{
			Ratio:    0,
			Level:    "UNKNOWN",
			Direction: "UNKNOWN",
			Strength:  0,
		}
	}

	volumes := make([]float64, len(ohlcData))
	closes := make([]float64, len(ohlcData))
	for i, d := range ohlcData {
		volumes[i] = d.Volume
		closes[i] = d.Close
	}

	currVol := volumes[len(volumes)-1]
	currClose := closes[len(closes)-1]
	prevClose := closes[len(closes)-2]

	avgVol := CalculateSMA(volumes, params.MAPeriod)
	avg := avgVol[len(avgVol)-1]
	ratio := currVol / avg

	var level string
	if ratio >= 2.0 {
		level = "VERY_HIGH"
	} else if ratio >= 1.5 {
		level = "HIGH"
	} else if ratio >= 1.0 {
		level = "NORMAL"
	} else {
		level = "LOW"
	}

	direction := "DOWN"
	if currClose > prevClose {
		direction = "UP"
	}

	return VolumeSignal{
		Ratio:     ratio,
		Level:     level,
		Direction: direction,
		Strength:  int(ratio * 50), // Simple strength calculation
	}
}
