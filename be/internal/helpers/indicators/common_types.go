package indicators

// OHLCData represents OHLCV candle data
type OHLCData struct {
	Timestamp int64   // Unix milliseconds (UTC)
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// Candle represents a single candlestick
type Candle struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
