package indicators

import (
	"math"
	"sync"
)

// Package-level cache for MA results (thread-safe)
var (
	maResultsCache = make(map[string]MAResult)
	maCacheMutex   = sync.RWMutex{}
)

// GetMACache retrieves MA result from cache
func GetMACache(key string) (MAResult, bool) {
	maCacheMutex.RLock()
	defer maCacheMutex.RUnlock()
	result, ok := maResultsCache[key]
	return result, ok
}

// SetMACache stores MA result in cache
func SetMACache(key string, result MAResult) {
	maCacheMutex.Lock()
	defer maCacheMutex.Unlock()
	maResultsCache[key] = result
}

// MAParameters holds the configuration for MA calculation
type MAParameters struct {
	SMAPeriods []int // e.g., []int{20, 50, 200}
	EMAPeriods []int // e.g., []int{12, 26}
}

// AnalyzeMAWithCache analyzes MA with caching support and config params
func AnalyzeMAWithCache(closes []float64, cacheKey string, cfg MAParameters) MAResult {
	// Check cache first
	if result, ok := GetMACache(cacheKey); ok {
		return result
	}

	// Calculate if not in cache using config params
	var result MAResult
	result = AnalyzeMAWithParams(closes, cfg)

	// Store in cache
	SetMACache(cacheKey, result)
	return result
}

// MAValues holds calculated MA values
type MAValues struct {
	SMA20  float64 `json:"sma20"`
	SMA50  float64 `json:"sma50"`
	SMA200 float64 `json:"sma200"`
	EMA12  float64 `json:"ema12"`
	EMA26  float64 `json:"ema26"`
}

// MAResult holds the Moving Average analysis result
type MAResult struct {
	Signal  int      `json:"signal"` // Score from -100 to 100
	Trend   string   `json:"trend"`  // BULLISH, BEARISH, NEUTRAL
	Values  MAValues `json:"values"`
	Details []string `json:"details"`
}

// AnalyzeMAWithParams analyzes Moving Average signal with custom parameters
func AnalyzeMAWithParams(closes []float64, params MAParameters) MAResult {
	if len(closes) < 200 {
		return MAResult{
			Signal:  0,
			Trend:   "NEUTRAL",
			Values:  MAValues{},
			Details: []string{"Insufficient data"},
		}
	}

	// ⚠️ NOTE: Using current candle (index -1) instead of last closed (index -2)
	//   lastClosedIdx := len(closes) - 2
	//   if lastClosedIdx < 0 { lastClosedIdx = 0 }
	//   curr := closes[lastClosedIdx]

	curr := closes[len(closes)-1] // Use current candle

	// Get periods with fallback to defaults
	smaPeriods := params.SMAPeriods
	if len(smaPeriods) == 0 {
		smaPeriods = []int{20, 50, 200}
	}
	emaPeriods := params.EMAPeriods
	if len(emaPeriods) == 0 {
		emaPeriods = []int{12, 26}
	}

	// Calculate SMA for each period
	smaValues := make(map[int]float64)
	for _, period := range smaPeriods {
		sma := CalculateSMA(closes, period)
		smaValues[period] = sma[len(sma)-1]
	}

	// Calculate EMA for each period
	emaValues := make(map[int]float64)
	for _, period := range emaPeriods {
		ema := CalculateEMA(closes, period)
		emaValues[period] = ema[len(ema)-1]
	}

	// Get standard values for response
	lastSma20 := smaValues[20]
	lastSma50 := smaValues[50]
	lastSma200 := smaValues[200]
	lastEma12 := emaValues[12]
	lastEma26 := emaValues[26]

	signal := 0
	details := make([]string, 0)

	// EMA12 vs EMA26 analysis
	if lastEma12 > lastEma26 {
		signal += 25
		details = append(details, "EMA12 > EMA26")
	} else {
		signal -= 25
		details = append(details, "EMA12 < EMA26")
	}

	// SMA trend alignment (bullish: SMA20 > SMA50 > SMA200)
	if lastSma20 > lastSma50 && lastSma50 > lastSma200 {
		signal += 35
		details = append(details, "SMA bullish")
	} else if lastSma20 < lastSma50 && lastSma50 < lastSma200 {
		signal -= 35
		details = append(details, "SMA bearish")
	}

	// Price vs SMA200
	if curr > lastSma200 {
		signal += 20
		details = append(details, "Above SMA200")
	} else {
		signal -= 20
		details = append(details, "Below SMA200")
	}

	// Price vs SMA20
	if curr > lastSma20 {
		signal += 20
		details = append(details, "Above SMA20")
	} else {
		signal -= 20
		details = append(details, "Below SMA20")
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(math.Max(-100, math.Min(100, float64(signal))))

	// Determine trend
	trend := "NEUTRAL"
	if clampedSignal >= 50 {
		trend = "BULLISH"
	} else if clampedSignal <= -50 {
		trend = "BEARISH"
	}

	return MAResult{
		Signal: clampedSignal,
		Trend:  trend,
		Values: MAValues{
			SMA20:  lastSma20,
			SMA50:  lastSma50,
			SMA200: lastSma200,
			EMA12:  lastEma12,
			EMA26:  lastEma26,
		},
		Details: details,
	}
}

// CalculateSMA calculates Simple Moving Average
func CalculateSMA(data []float64, period int) []float64 {
	if len(data) < period {
		return []float64{}
	}

	result := make([]float64, 0, len(data)-period+1)
	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += data[j]
		}
		result = append(result, sum/float64(period))
	}
	return result
}

// CalculateEMA calculates Exponential Moving Average
func CalculateEMA(data []float64, period int) []float64 {
	if len(data) < period {
		return []float64{}
	}

	result := make([]float64, 0, len(data)-period+1)
	multiplier := 2.0 / float64(period+1)

	// Calculate first SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result = append(result, sum/float64(period))

	// Calculate EMA for rest
	for i := period; i < len(data); i++ {
		ema := (data[i]-result[len(result)-1])*multiplier + result[len(result)-1]
		result = append(result, ema)
	}

	return result
}
