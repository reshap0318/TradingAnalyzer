package binance

import (
	"strconv"
	"time"
)

// parseFloat parses float string
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// parseInt parses int string
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

// parseTime parses timestamp to time.Time
func parseTime(timestamp int64) time.Time {
	return time.Unix(0, timestamp*int64(time.Millisecond))
}

// getCurrentTimestamp returns current timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// formatQuantity formats quantity to string with proper precision
func formatQuantity(qty float64) string {
	return strconv.FormatFloat(qty, 'f', -1, 64)
}

// formatPrice formats price to string with proper precision
func formatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}

// getMaxLeverageBySymbol returns maximum leverage based on symbol category
// Reference: https://www.binance.com/en/futures/leverage-bracket
// Note: Max leverage can vary based on user's account tier and symbol bracket
// Last Updated: March 2026
func getMaxLeverageBySymbol(symbol string) int {
	// Tier 1: 125x max - BTC & ETH only
	tier1 := []string{
		"BTCUSDT", "ETHUSDT",
		"BTCBUSD", "ETHBUSD",
	}
	for _, s := range tier1 {
		if symbol == s {
			return 125
		}
	}

	// Tier 2: 75x max - Major altcoins & established tokens
	tier2 := []string{
		// Major coins
		"BNBUSDT", "SOLUSDT", "XRPUSDT",
		// Layer 1
		"ADAUSDT", "DOTUSDT", "AVAXUSDT", "MATICUSDT",
		"LINKUSDT", "ATOMUSDT", "LTCUSDT", "BCHUSDT",
		"EOSUSDT", "TRXUSDT", "XLMUSDT", "ALGOUSDT",
		"VETUSDT", "FILUSDT", "NEARUSDT", "APTUSDT",
		"SUIUSDT", "ARBUSDT", "OPUSDT", "LDOUSDT",
		// Meme coins (popular)
		"DOGEUSDT", "SHIBUSDT",
		// Others
		"UNIUSDT", "AAVEUSDT", "SUSHIUSDT", "MKRUSDT",
		"ETCUSDT", "XMRUSDT", "FTMUSDT", "SANDUSDT",
		"MANAUSDT", "AXSUSDT", "THETAUSDT", "ICPUSDT",
	}
	for _, s := range tier2 {
		if symbol == s {
			return 75
		}
	}

	// Tier 3: 50x max - DeFi & mid-cap tokens
	tier3 := []string{
		"COMPUSDT", "YFIUSDT", "SNXUSDT", "CRVUSDT", "BALUSDT",
		"RENUSDT", "KNCUSDT", "LRCUSDT", "ZRXUSDT", "BANDUSDT",
		"RUNEUSDT", "KAVAUSDT", "IOTAUSDT", "ZECUSDT", "DASHUSDT",
		"ENJUSDT", "CHZUSDT", "BATUSDT", "QTUMUSDT", "ZILUSDT",
	}
	for _, s := range tier3 {
		if symbol == s {
			return 50
		}
	}

	// Tier 4: 25x max - Small cap & newer tokens
	tier4 := []string{
		"PEPEUSDT", "FLOKIUSDT", "GALAUSDT", "ROSEUSDT",
		"ANKRUSDT", "STORJUSDT", "COTIUSDT", "CHRUSDT",
	}
	for _, s := range tier4 {
		if symbol == s {
			return 25
		}
	}

	// Default - 20x max for unknown/new tokens (safer default)
	return 20
}

// AdjustQuantityPrecision adjusts quantity to match symbol's step size
// Example: If stepSize = 0.001 and qty = 0.12345 → returns 0.123
func AdjustQuantityPrecision(quantity float64, stepSize float64) float64 {
	if stepSize <= 0 {
		stepSize = 0.001 // Default step size
	}

	// Calculate precision from step size
	precision := getPrecisionFromStep(stepSize)

	// Round down to match step size
	adjusted := float64(int(quantity/stepSize)) * stepSize

	// Ensure we don't exceed original quantity and apply precision
	adjusted = roundToPrecision(adjusted, precision)

	return adjusted
}

// AdjustPricePrecision adjusts price to match symbol's tick size
// Example: If tickSize = 0.01 and price = 50000.123 → returns 50000.12
func AdjustPricePrecision(price float64, tickSize float64) float64 {
	if tickSize <= 0 {
		tickSize = 0.01 // Default tick size
	}

	// Calculate precision from tick size
	precision := getPrecisionFromStep(tickSize)

	// Round to match tick size
	adjusted := float64(int(price/tickSize)) * tickSize

	// Apply precision
	adjusted = roundToPrecision(adjusted, precision)

	return adjusted
}

// getPrecisionFromStep calculates decimal precision from step/tick size
// Example: stepSize = 0.001 → precision = 3
func getPrecisionFromStep(stepSize float64) int {
	if stepSize <= 0 {
		return 3
	}

	precision := 0
	for stepSize < 1 {
		stepSize *= 10
		precision++
	}
	return precision
}

// roundToPrecision rounds a float to specified decimal precision
func roundToPrecision(value float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier)) / multiplier
}
