package indicators

import (
	"fmt"
	"time"

	"github.com/reshap/trading-bot/internal/helpers"
)

// CandlePatternWithDate holds a detected pattern with WIB timestamp
type CandlePatternWithDate struct {
	Date     string        `json:"date"`     // WIB formatted string
	Patterns []CandlePattern `json:"patterns"` // List of patterns detected
}

// PatternHistoryResult holds pattern analysis for multiple candles
type PatternHistoryResult struct {
	History      []CandlePatternWithDate `json:"history"`
	TotalBullish int                     `json:"total_bullish"`
	TotalBearish int                     `json:"total_bearish"`
	TotalNeutral int                     `json:"total_neutral"`
	RecentSignal int                     `json:"recent_signal"` // Last candle signal
	TrendScore   int                     `json:"trend_score"`   // 5-candle aggregated score
	Details      []string                `json:"details"`
}

// Candle pattern scoring weights (matching GitHub TradingAnalyzer)
const (
	// Bullish patterns (positive score)
	BullishEngulfingScore = 16
	MorningStarScore      = 20
	HammerScore           = 10
	PiercingLineScore     = 10
	BullishMarubozuScore  = 16

	// Bearish patterns (negative score)
	BearishEngulfingScore = -16
	EveningStarScore      = -20
	ShootingStarScore     = -10
	DarkCloudCoverScore   = -10
	BearishMarubozuScore  = -16

	// Neutral patterns
	DojiScore = 0
)

// CandlePattern represents a detected candlestick pattern
type CandlePattern string

// Candlestick patterns constants
const (
	// Bullish patterns
	BullishEngulfing CandlePattern = "Bullish Engulfing"
	Hammer           CandlePattern = "Hammer"
	MorningStar      CandlePattern = "Morning Star"
	PiercingLine     CandlePattern = "Piercing Line"

	// Bearish patterns
	BearishEngulfing CandlePattern = "Bearish Engulfing"
	ShootingStar     CandlePattern = "Shooting Star"
	EveningStar      CandlePattern = "Evening Star"
	DarkCloudCover   CandlePattern = "Dark Cloud Cover"

	// Neutral/Trend patterns
	Doji              CandlePattern = "Doji"
	BullishMarubozu   CandlePattern = "Bullish Marubozu"
	BearishMarubozu   CandlePattern = "Bearish Marubozu"
)

// Helper functions untuk Candle
func isGreen(c Candle) bool    { return c.Close > c.Open }
func isRed(c Candle) bool      { return c.Close < c.Open }
func bodySize(c Candle) float64 { return abs(c.Close - c.Open) }

func upperWick(c Candle) float64 {
	return c.High - maxF(c.Open, c.Close)
}

func lowerWick(c Candle) float64 {
	return minF(c.Open, c.Close) - c.Low
}

func totalRange(c Candle) float64 {
	return c.High - c.Low
}

// isBullishEngulfing checks for bullish engulfing pattern
// Prev: Red, Curr: Green. Curr Body engulfs Prev Body.
func isBullishEngulfing(prev, curr Candle) bool {
	return isRed(prev) &&
		isGreen(curr) &&
		curr.Open <= prev.Close &&
		curr.Close >= prev.Open
}

// isHammer checks for hammer pattern
// Small body at top, Long lower wick (>= 2x body), Small upper wick
func isHammer(curr Candle) bool {
	body := bodySize(curr)
	uWick := upperWick(curr)
	lWick := lowerWick(curr)

	return lWick >= 2*body && uWick <= body*1.0
}

// isMorningStar checks for morning star pattern (3 candles)
// 1: Long Red. 2: Small body. 3: Strong Green (closes > midpoint of 1)
func isMorningStar(c1, c2, c3 Candle) bool {
	midpoint := (c1.Open + c1.Close) / 2.0

	return isRed(c1) &&
		bodySize(c1) > totalRange(c1)*0.5 &&
		bodySize(c2) < bodySize(c1)*0.3 &&
		isGreen(c3) &&
		c3.Close > midpoint
}

// isPiercingLine checks for piercing line pattern
// Prev: Long Red. Curr: Open below prev low, Close > 50% of prev body.
func isPiercingLine(prev, curr Candle) bool {
	midpoint := (prev.Open + prev.Close) / 2.0

	return isRed(prev) &&
		curr.Open < prev.Low &&
		isGreen(curr) &&
		curr.Close > midpoint &&
		curr.Close < prev.Open
}

// isDoji checks for doji pattern
// Body very small relative to range (<= 10%)
func isDoji(curr Candle) bool {
	return bodySize(curr) <= totalRange(curr)*0.1
}

// isBearishEngulfing checks for bearish engulfing pattern
func isBearishEngulfing(prev, curr Candle) bool {
	return isGreen(prev) &&
		isRed(curr) &&
		curr.Open >= prev.Close &&
		curr.Close <= prev.Open
}

// isShootingStar checks for shooting star pattern (bearish hammer)
func isShootingStar(curr Candle) bool {
	body := bodySize(curr)
	uWick := upperWick(curr)
	lWick := lowerWick(curr)

	return uWick >= 2*body && lWick <= body*0.5
}

// isMarubozu checks for marubozu pattern
// Big body, very small wicks (> 85% of total range)
func isMarubozu(curr Candle) bool {
	body := bodySize(curr)
	cRange := totalRange(curr)

	return cRange > 0 && body/cRange > 0.85
}

// isEveningStar checks for evening star pattern (bearish morning star)
func isEveningStar(c1, c2, c3 Candle) bool {
	midpoint := (c1.Open + c1.Close) / 2.0

	return isGreen(c1) &&
		bodySize(c1) > totalRange(c1)*0.5 &&
		bodySize(c2) < bodySize(c1)*0.3 &&
		isRed(c3) &&
		c3.Close < midpoint
}

// isDarkCloudCover checks for dark cloud cover pattern (bearish piercing)
func isDarkCloudCover(prev, curr Candle) bool {
	midpoint := (prev.Open + prev.Close) / 2.0

	return isGreen(prev) &&
		curr.Open > prev.High &&
		isRed(curr) &&
		curr.Close < midpoint &&
		curr.Close > prev.Open
}

// DetectCandlePatterns detects candlestick patterns in the recent data
// Returns array of detected pattern names
func DetectCandlePatterns(opens, highs, lows, closes []float64) []CandlePattern {
	if len(closes) < 6 {  // Need at least 6 candles (5 closed + 1 current)
		return []CandlePattern{}
	}

	// IMPORTANT: Use last CLOSED candle, not current candle
	// - Index -1 (len-1): Current candle - MAY NOT BE CLOSED YET (still forming)
	// - Index -2 (len-2): Last closed candle - FULLY CLOSED (stable pattern)
	// 
	// Example: Current time 00:17, 15m candle
	//   Candle 00:15-00:29: NOT CLOSED (index -1) ❌ Unreliable, pattern may change
	//   Candle 00:00-00:14: CLOSED (index -2) ✅ Stable pattern, won't change
	//
	// We use -2 to ensure we're analyzing completed candles only.
	// This matches GitHub's behavior and prevents false pattern detection.
	lastClosedIdx := len(closes) - 2  // Previous candle (last closed)

	// Helper to get candle at index
	getCandle := func(i int) Candle {
		return Candle{
			Open:  opens[i],
			High:  highs[i],
			Low:   lows[i],
			Close: closes[i],
		}
	}

	curr := getCandle(lastClosedIdx)
	prev1 := getCandle(lastClosedIdx - 1)
	prev2 := getCandle(lastClosedIdx - 2)

	// Ignore flat candles (no trading activity or error)
	if totalRange(curr) == 0 {
		return []CandlePattern{}
	}

	patterns := make([]CandlePattern, 0)

	// Bullish patterns
	if isBullishEngulfing(prev1, curr) {
		patterns = append(patterns, BullishEngulfing)
	}
	if isHammer(curr) {
		patterns = append(patterns, Hammer)
	}
	if isMorningStar(prev2, prev1, curr) {
		patterns = append(patterns, MorningStar)
	}
	if isPiercingLine(prev1, curr) {
		patterns = append(patterns, PiercingLine)
	}

	// Bearish patterns
	if isBearishEngulfing(prev1, curr) {
		patterns = append(patterns, BearishEngulfing)
	}
	if isShootingStar(curr) {
		patterns = append(patterns, ShootingStar)
	}
	if isEveningStar(prev2, prev1, curr) {
		patterns = append(patterns, EveningStar)
	}
	if isDarkCloudCover(prev1, curr) {
		patterns = append(patterns, DarkCloudCover)
	}

	// Neutral/Trend patterns
	if isDoji(curr) {
		patterns = append(patterns, Doji)
	}
	if isMarubozu(curr) {
		if isGreen(curr) {
			patterns = append(patterns, BullishMarubozu)
		} else {
			patterns = append(patterns, BearishMarubozu)
		}
	}

	return patterns
}

// CandlePatternResult holds the candle pattern analysis result
type CandlePatternResult struct {
	Patterns      []CandlePattern
	BullishCount  int
	BearishCount  int
	NeutralCount  int
	Signal        int // Score from -100 to 100
	Details       []string
}

// AnalyzeCandlePatterns analyzes detected patterns and generates signal
func AnalyzeCandlePatterns(opens, highs, lows, closes []float64) CandlePatternResult {
	patterns := DetectCandlePatterns(opens, highs, lows, closes)

	if len(patterns) == 0 {
		return CandlePatternResult{
			Patterns:     []CandlePattern{},
			BullishCount: 0,
			BearishCount: 0,
			NeutralCount: 0,
			Signal:       0,
			Details:      []string{"No patterns detected"},
		}
	}

	bullishCount := 0
	bearishCount := 0
	neutralCount := 0
	signal := 0
	details := make([]string, 0)

	for _, pattern := range patterns {
		score := getPatternScore(pattern)
		
		if score > 0 {
			bullishCount++
		} else if score < 0 {
			bearishCount++
		} else {
			neutralCount++
		}
		
		signal += score
		details = append(details, string(pattern))
	}

	// Clamp signal between -100 and 100
	clampedSignal := int(maxF(-100, minF(100, float64(signal))))

	return CandlePatternResult{
		Patterns:     patterns,
		BullishCount: bullishCount,
		BearishCount: bearishCount,
		NeutralCount: neutralCount,
		Signal:       clampedSignal,
		Details:      details,
	}
}

// getPatternScore returns the score for a specific pattern
func getPatternScore(pattern CandlePattern) int {
	switch pattern {
	// Bullish patterns
	case BullishEngulfing:
		return BullishEngulfingScore
	case MorningStar:
		return MorningStarScore
	case Hammer:
		return HammerScore
	case PiercingLine:
		return PiercingLineScore
	case BullishMarubozu:
		return BullishMarubozuScore

	// Bearish patterns
	case BearishEngulfing:
		return BearishEngulfingScore
	case EveningStar:
		return EveningStarScore
	case ShootingStar:
		return ShootingStarScore
	case DarkCloudCover:
		return DarkCloudCoverScore
	case BearishMarubozu:
		return BearishMarubozuScore

	// Neutral patterns
	case Doji:
		return DojiScore

	default:
		return 0
	}
}

// AnalyzeCandlePatternHistory analyzes patterns in the last N candles
// Returns pattern history with WIB timestamps and aggregated signals
func AnalyzeCandlePatternHistory(
	opens, highs, lows, closes []float64,
	timestamps []int64, // Unix milliseconds
	historyCount int,
) PatternHistoryResult {
	if len(closes) < historyCount+4 { // Need at least 5 candles for patterns
		return PatternHistoryResult{
			Details: []string{"Insufficient data for history"},
		}
	}

	history := make([]CandlePatternWithDate, 0, historyCount)
	totalBullish := 0
	totalBearish := 0
	totalNeutral := 0
	trendScore := 0

	// Analyze last N candles
	for i := 0; i < historyCount; i++ {
		idx := len(closes) - 1 - i

		// Need at least 5 candles ending at idx for pattern detection
		if idx < 4 {
			break
		}

		// Slice data up to current index
		sliceEnd := idx + 1
		sliceOpens := opens[:sliceEnd]
		sliceHighs := highs[:sliceEnd]
		sliceLows := lows[:sliceEnd]
		sliceCloses := closes[:sliceEnd]

		// Detect patterns for this candle
		patterns := DetectCandlePatterns(sliceOpens, sliceHighs, sliceLows, sliceCloses)

		// Get timestamp and convert to WIB string
		var dateStr string
		if timestamps != nil && len(timestamps) > idx {
			// Convert Unix milliseconds to time.Time
			ts := time.Unix(0, timestamps[idx]*int64(time.Millisecond))
			// Format as WIB string
			dateStr = helpers.FormatWIBDefault(ts)
		} else {
			dateStr = ""
		}

		// Add to history
		history = append(history, CandlePatternWithDate{
			Date:     dateStr,
			Patterns: patterns,
		})

		// Count patterns
		for _, pattern := range patterns {
			score := getPatternScore(pattern)
			if score > 0 {
				totalBullish++
			} else if score < 0 {
				totalBearish++
			} else {
				totalNeutral++
			}
			trendScore += score
		}
	}

	// Calculate recent signal (last candle only)
	singleResult := AnalyzeCandlePatterns(opens, highs, lows, closes)

	return PatternHistoryResult{
		History:      history,
		TotalBullish: totalBullish,
		TotalBearish: totalBearish,
		TotalNeutral: totalNeutral,
		RecentSignal: singleResult.Signal,
		TrendScore:   trendScore,
		Details:      buildHistoryDetails(history),
	}
}

// buildHistoryDetails creates human-readable details from history
func buildHistoryDetails(history []CandlePatternWithDate) []string {
	details := make([]string, 0)

	for _, h := range history {
		if len(h.Patterns) > 0 {
			for _, p := range h.Patterns {
				// Format: "28/02/2026, 22.30.00 - Bullish Marubozu"
				details = append(details, fmt.Sprintf("%s - %s", h.Date, string(p)))
			}
		}
	}

	return details
}
