package indicators

import (
	"math"
)

// TrendlineResult represents the dynamic channel levels at the current candle
type TrendlineResult struct {
	SupportIsValid    bool
	CurrentSupport    float64
	SupportSlope      float64

	ResistanceIsValid bool
	CurrentResistance float64
	ResistanceSlope   float64
}

// AnalyzeDynamicTrendlines calculates the dominant sliding diagonal support and resistance
func AnalyzeDynamicTrendlines(data []OHLCData, lookback int) TrendlineResult {
	result := TrendlineResult{}

	if len(data) < lookback*2 {
		return result
	}

	// Restrict analysis to recent data to avoid ancient irrelevant lines
	analysisWindow := 200
	if len(data) < analysisWindow {
		analysisWindow = len(data)
	}
	recentData := data[len(data)-analysisWindow:]

	// Find swing points using existing helpers
	swingHighs := findSwingHighs(recentData, 5)
	swingLows := findSwingLows(recentData, 5)

	bestResScore := 0
	bestResPrice := 0.0
	bestResSlope := 0.0

	// 1. Find Best Resistance Trendline
	for i := 0; i < len(swingHighs)-1; i++ {
		for j := i + 1; j < len(swingHighs); j++ {
			p1 := swingHighs[i]
			p2 := swingHighs[j]

			// Calculate equation: y = mx + b
			m := (p2.Price - p1.Price) / float64(p2.Index-p1.Index)
			b := p1.Price - (m * float64(p1.Index))

			isValid := true
			touches := 0

			// Validate line against all candles from p1 to end
			for k := p1.Index; k < len(recentData); k++ {
				candleHigh := recentData[k].High
				linePrice := (m * float64(k)) + b

				// If price breaks above resistance line by > 0.2%, line is broken
				if candleHigh > linePrice*1.002 {
					isValid = false
					break
				}

				// Count touches (if high is within 0.5% of the line)
				if math.Abs(candleHigh-linePrice)/linePrice <= 0.005 {
					touches++
				}
			}

			if isValid && touches >= 2 { // Min 2 touches (p1 and p2)
				score := touches
				// Bonus score if it's a recent line (p2 is close to end)
				if len(recentData)-p2.Index < 20 {
					score += 1
				}
				
				if score > bestResScore {
					bestResScore = score
					bestResSlope = m
					// Calculate price AT THE VERY LAST CANDLE
					bestResPrice = (m * float64(len(recentData)-1)) + b
				}
			}
		}
	}

	if bestResScore > 0 {
		result.ResistanceIsValid = true
		result.CurrentResistance = bestResPrice
		result.ResistanceSlope = bestResSlope
	}

	bestSupScore := 0
	bestSupPrice := 0.0
	bestSupSlope := 0.0

	// 2. Find Best Support Trendline
	for i := 0; i < len(swingLows)-1; i++ {
		for j := i + 1; j < len(swingLows); j++ {
			p1 := swingLows[i]
			p2 := swingLows[j]

			m := (p2.Price - p1.Price) / float64(p2.Index-p1.Index)
			b := p1.Price - (m * float64(p1.Index))

			isValid := true
			touches := 0

			for k := p1.Index; k < len(recentData); k++ {
				candleLow := recentData[k].Low
				linePrice := (m * float64(k)) + b

				// If price breaks below support line by > 0.2%, line is broken
				if candleLow < linePrice*0.998 {
					isValid = false
					break
				}

				if math.Abs(linePrice-candleLow)/linePrice <= 0.005 {
					touches++
				}
			}

			if isValid && touches >= 2 {
				score := touches
				if len(recentData)-p2.Index < 20 {
					score += 1
				}

				if score > bestSupScore {
					bestSupScore = score
					bestSupSlope = m
					bestSupPrice = (m * float64(len(recentData)-1)) + b
				}
			}
		}
	}

	if bestSupScore > 0 {
		result.SupportIsValid = true
		result.CurrentSupport = bestSupPrice
		result.SupportSlope = bestSupSlope
	}

	return result
}
