package indicators

// TrendBonusResult holds the trend bonus analysis result
type TrendBonusResult struct {
	Signal  int      // Score: -20, 0, or +20
	Details []string // Analysis details
	Trend   string   // "UPTREND", "DOWNTREND", or "NEUTRAL"
}

// AnalyzeTrendBonus analyzes trend alignment between MA and MACD
// Returns +20 for strong uptrend (MA+ + MACD+), -20 for strong downtrend (MA- + MACD-)
// Parameters:
//   - maSignal: Pre-calculated MA signal (-100 to 100)
//   - macdSignal: Pre-calculated MACD signal (-100 to 100)
func AnalyzeTrendBonus(maSignal, macdSignal int) TrendBonusResult {
	details := make([]string, 0)
	signal := 0
	trend := "NEUTRAL"

	// Check if both MA and MACD are aligned
	if maSignal > 0 && macdSignal > 0 {
		// Strong uptrend
		signal = 100
		trend = "UPTREND"
		details = append(details, "Strong Uptrend (MA+MACD aligned)")
	} else if maSignal < 0 && macdSignal < 0 {
		// Strong downtrend
		signal = -100
		trend = "DOWNTREND"
		details = append(details, "Strong Downtrend (MA+MACD aligned)")
	} else {
		// Mixed or neutral
		signal = 0
		trend = "NEUTRAL"

		// Provide details about the mix
		if maSignal > 0 && macdSignal < 0 {
			details = append(details, "Mixed signals (MA bullish, MACD bearish)")
		} else if maSignal < 0 && macdSignal > 0 {
			details = append(details, "Mixed signals (MA bearish, MACD bullish)")
		} else {
			details = append(details, "No clear trend alignment")
		}
	}

	return TrendBonusResult{
		Signal:  signal,
		Details: details,
		Trend:   trend,
	}
}
