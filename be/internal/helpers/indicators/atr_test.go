package indicators

import (
	"testing"
)

// TestCalculateATR tests ATR calculation
func TestCalculateATR(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		ohlcData := make([]OHLCData, 10)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100,
				High:  105,
				Low:   95,
				Close: 102,
				Volume: 1000,
			}
		}

		result := CalculateATR(ohlcData, 14)

		if len(result) != 0 {
			t.Errorf("expected empty result, got length %d", len(result))
		}
	})

	t.Run("sufficient data", func(t *testing.T) {
		// Create 30 data points
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: 1000,
			}
		}

		result := CalculateATR(ohlcData, 14)

		// Should have 30 - 14 = 16 values (one less than input due to TR calculation)
		expectedLen := 30 - 14
		if len(result) != expectedLen {
			t.Errorf("expected length %d, got %d", expectedLen, len(result))
		}

		// ATR should be positive
		for i, atr := range result {
			if atr <= 0 {
				t.Errorf("ATR value at index %d should be positive, got %f", i, atr)
			}
		}
	})

	t.Run("constant data", func(t *testing.T) {
		// With constant highs/lows, True Range should be 0
		ohlcData := make([]OHLCData, 20)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100,
				High:  100,
				Low:   100,
				Close: 100,
				Volume: 1000,
			}
		}

		result := CalculateATR(ohlcData, 14)

		if len(result) > 0 && result[0] != 0 {
			t.Errorf("expected ATR to be 0 for constant data, got %f", result[0])
		}
	})

	t.Run("volatile data", func(t *testing.T) {
		// Create data with varying volatility
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i)*2,
				High:  110 + float64(i)*2, // Wide range
				Low:   90 + float64(i)*2,  // Wide range
				Close: 105 + float64(i)*2,
				Volume: 1000,
			}
		}

		result := CalculateATR(ohlcData, 14)

		// ATR should be larger for volatile data
		if len(result) > 0 && result[0] < 5 {
			t.Errorf("expected higher ATR for volatile data, got %f", result[0])
		}
	})
}

// TestAnalyzeATR tests ATR analysis
func TestAnalyzeATR(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		ohlcData := make([]OHLCData, 15)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100,
				High:  105,
				Low:   95,
				Close: 102,
				Volume: 1000,
			}
		}

		result := AnalyzeATR(ohlcData)

		if result.ATR != 0 {
			t.Errorf("expected ATR 0, got %f", result.ATR)
		}
		if result.Volatility != "UNKNOWN" {
			t.Errorf("expected volatility UNKNOWN, got %s", result.Volatility)
		}
		if len(result.Details) != 1 || result.Details[0] != "Insufficient data" {
			t.Errorf("expected 'Insufficient data' detail, got %v", result.Details)
		}
	})

	t.Run("high volatility", func(t *testing.T) {
		// Create data with increasing volatility
		ohlcData := make([]OHLCData, 50)
		for i := range ohlcData {
			rangeSize := 20.0 + float64(i)*0.5 // Increasing range
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i)*2,
				High:  100 + float64(i)*2 + rangeSize,
				Low:   100 + float64(i)*2 - rangeSize,
				Close: 100 + float64(i)*2,
				Volume: 1000,
			}
		}

		result := AnalyzeATR(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// ATR should be positive
		if result.ATR <= 0 {
			t.Errorf("expected positive ATR, got %f", result.ATR)
		}

		// ATRPercent should be positive
		if result.ATRPercent <= 0 {
			t.Errorf("expected positive ATRPercent, got %f", result.ATRPercent)
		}
	})

	t.Run("low volatility", func(t *testing.T) {
		// Create data with very small ranges (low volatility)
		ohlcData := make([]OHLCData, 50)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i)*0.1,
				High:  100.5 + float64(i)*0.1, // Very small range
				Low:   99.5 + float64(i)*0.1,  // Very small range
				Close: 100 + float64(i)*0.1,
				Volume: 1000,
			}
		}

		result := AnalyzeATR(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// ATR ratio should be calculated
		if result.ATRRatio <= 0 {
			t.Errorf("expected positive ATRRatio, got %f", result.ATRRatio)
		}
	})

	t.Run("normal volatility", func(t *testing.T) {
		// Create data with moderate volatility
		ohlcData := make([]OHLCData, 50)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  103 + float64(i),
				Low:   97 + float64(i),
				Close: 100 + float64(i),
				Volume: 1000,
			}
		}

		result := AnalyzeATR(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}
	})
}

// TestATRResultStructure tests the structure of ATRResult
func TestATRResultStructure(t *testing.T) {
	ohlcData := make([]OHLCData, 50)
	for i := range ohlcData {
		ohlcData[i] = OHLCData{
			Open:  100 + float64(i),
			High:  105 + float64(i),
			Low:   95 + float64(i),
			Close: 102 + float64(i),
			Volume: 1000,
		}
	}

	result := AnalyzeATR(ohlcData)

	// Check that ATR is populated
	if result.ATR == 0 {
		t.Error("expected ATR to be populated")
	}

	// Check that volatility is populated
	if result.Volatility == "" {
		t.Error("expected volatility to be populated")
	}

	// Check that details are populated
	if len(result.Details) == 0 {
		t.Error("expected details to be populated")
	}
}

// TestCalculateATRWithCustomParams tests ATR with custom parameters
func TestCalculateATRWithCustomParams(t *testing.T) {
	ohlcData := make([]OHLCData, 50)
	for i := range ohlcData {
		ohlcData[i] = OHLCData{
			Open:  100 + float64(i),
			High:  105 + float64(i),
			Low:   95 + float64(i),
			Close: 102 + float64(i),
			Volume: 1000,
		}
	}

	// Custom period
	result := CalculateATR(ohlcData, 10)

	// Should have valid results
	if len(result) == 0 {
		t.Error("expected non-empty ATR values")
	}
}
