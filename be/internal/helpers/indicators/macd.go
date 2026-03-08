package indicators

import (
	"math"
	"sync"
)

// Package-level cache for MACD results (thread-safe)
var (
	macdResultsCache = make(map[string]MACDResult)
	macdCacheMutex   = sync.RWMutex{}
)

// GetMACDCache retrieves MACD result from cache
func GetMACDCache(key string) (MACDResult, bool) {
	macdCacheMutex.RLock()
	defer macdCacheMutex.RUnlock()
	result, ok := macdResultsCache[key]
	return result, ok
}

// SetMACDCache stores MACD result in cache
func SetMACDCache(key string, result MACDResult) {
	macdCacheMutex.Lock()
	defer macdCacheMutex.Unlock()
	macdResultsCache[key] = result
}

// AnalyzeMACDWithCache analyzes MACD with caching support and config params
func AnalyzeMACDWithCache(closes []float64, cacheKey string, cfg MACDParameters) MACDResult {
	// Check cache first
	if result, ok := GetMACDCache(cacheKey); ok {
		return result
	}

	// Calculate if not in cache using config params
	var result MACDResult
	result = AnalyzeMACDWithParams(closes, cfg)

	// Store in cache
	SetMACDCache(cacheKey, result)
	return result
}

// MACDParameters holds the configuration for MACD calculation
type MACDParameters struct {
	Fast   int
	Slow   int
	Signal int
}

// MACDValues holds the current MACD values
type MACDValues struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// MACDResult holds the complete MACD analysis result
type MACDResult struct {
	Signal  int // Score from -100 to 100
	Values  MACDValues
	Details []string
}

// MACDLines holds all MACD line data
type MACDLines struct {
	MACDLine      []float64
	SignalLine    []float64
	HistogramLine []float64
}

// CalculateMACD calculates MACD Line, Signal Line, and Histogram
// Returns all line values for the given closes data
func CalculateMACD(closes []float64, params MACDParameters) MACDLines {
	fast, slow, signal := params.Fast, params.Slow, params.Signal

	// Check if we have enough data
	if len(closes) < slow+signal {
		return MACDLines{
			MACDLine:      []float64{},
			SignalLine:    []float64{},
			HistogramLine: []float64{},
		}
	}

	// Calculate EMA fast and slow
	emaFast := CalculateEMA(closes, fast)
	emaSlow := CalculateEMA(closes, slow)

	// Calculate MACD Line (EMA Fast - EMA Slow)
	// Align the arrays: emaFast is longer than emaSlow by (slow - fast) elements
	macdLine := make([]float64, 0, len(emaSlow))
	for i := 0; i < len(emaSlow); i++ {
		macdLine = append(macdLine, emaFast[i+slow-fast]-emaSlow[i])
	}

	// Calculate Signal Line (EMA of MACD Line)
	signalLine := CalculateEMA(macdLine, signal)

	// Calculate Histogram (MACD Line - Signal Line)
	histogram := make([]float64, 0, len(signalLine))
	for i := 0; i < len(signalLine); i++ {
		histogram = append(histogram, macdLine[i+signal-1]-signalLine[i])
	}

	return MACDLines{
		MACDLine:      macdLine,
		SignalLine:    signalLine,
		HistogramLine: histogram,
	}
}

// AnalyzeMACDWithParams analyzes MACD signal with custom parameters
func AnalyzeMACDWithParams(closes []float64, params MACDParameters) MACDResult {
	// If signal instability occurs, revert to skipping current candle:
	//   if len(closes) > 1 { closes = closes[:len(closes)-1] }

	lines := CalculateMACD(closes, params)

	// Check if we have enough histogram data
	if len(lines.HistogramLine) < 3 {
		return MACDResult{
			Signal:  0,
			Values:  MACDValues{},
			Details: []string{"Insufficient data"},
		}
	}

	// Get current and previous values
	currMACD := lines.MACDLine[len(lines.MACDLine)-1]
	currSig := lines.SignalLine[len(lines.SignalLine)-1]
	prevMACD := lines.MACDLine[len(lines.MACDLine)-2]
	prevSig := lines.SignalLine[len(lines.SignalLine)-2]
	currHist := lines.HistogramLine[len(lines.HistogramLine)-1]
	prevHist := lines.HistogramLine[len(lines.HistogramLine)-2]

	signal := 0
	details := make([]string, 0)

	// MACD/Signal cross analysis
	if prevMACD <= prevSig && currMACD > currSig {
		// Bullish cross
		signal += 44
		details = append(details, "MACD bullish cross")
	} else if prevMACD >= prevSig && currMACD < currSig {
		// Bearish cross
		signal -= 44
		details = append(details, "MACD bearish cross")
	} else if currMACD > currSig {
		// MACD above signal
		signal += 22
		details = append(details, "MACD above signal")
	} else {
		// MACD below signal
		signal -= 22
		details = append(details, "MACD below signal")
	}

	// MACD above/below zero
	if currMACD > 0 {
		signal += 17
		details = append(details, "MACD above zero")
	} else {
		signal -= 17
		details = append(details, "MACD below zero")
	}

	// Histogram rising/falling
	if currHist > prevHist {
		signal += 17
		details = append(details, "Histogram rising")
	} else {
		signal -= 17
		details = append(details, "Histogram falling")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	return MACDResult{
		Signal: clampedSignal,
		Values: MACDValues{
			MACD:      currMACD,
			Signal:    currSig,
			Histogram: currHist,
		},
		Details: details,
	}
}
