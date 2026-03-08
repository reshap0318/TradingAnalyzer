package indicators

import (
	"testing"
)

// TestCalculateStochastic tests Stochastic calculation
func TestCalculateStochastic(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		highs := []float64{100, 102, 101, 103, 102}
		lows := []float64{98, 99, 97, 100, 99}
		closes := []float64{99, 101, 100, 102, 101}

		result := CalculateStochastic(highs, lows, closes, 14, 3, 3)

		if len(result.K) != 0 || len(result.D) != 0 {
			t.Errorf("expected empty result, got K length %d, D length %d", len(result.K), len(result.D))
		}
	})

	t.Run("sufficient data", func(t *testing.T) {
		// Create 30 data points
		highs := make([]float64, 30)
		lows := make([]float64, 30)
		closes := make([]float64, 30)

		for i := range highs {
			highs[i] = 100 + float64(i)
			lows[i] = 90 + float64(i)
			closes[i] = 95 + float64(i)
		}

		result := CalculateStochastic(highs, lows, closes, 14, 3, 3)

		// Should have valid results
		if len(result.K) == 0 {
			t.Error("expected non-empty K line")
		}
		if len(result.D) == 0 {
			t.Error("expected non-empty D line")
		}

		// K and D should be between 0 and 100
		for i, k := range result.K {
			if k < 0 || k > 100 {
				t.Errorf("K value at index %d is out of range: %f", i, k)
			}
		}
		for i, d := range result.D {
			if d < 0 || d > 100 {
				t.Errorf("D value at index %d is out of range: %f", i, d)
			}
		}
	})

	t.Run("constant highs/lows", func(t *testing.T) {
		// When highs == lows, %K should be 50
		highs := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}
		lows := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}
		closes := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}

		result := CalculateStochastic(highs, lows, closes, 14, 3, 3)

		if len(result.K) > 0 && result.K[len(result.K)-1] != 50 {
			t.Errorf("expected K to be 50 for constant data, got %f", result.K[len(result.K)-1])
		}
	})
}

// TestAnalyzeStochastic tests Stochastic analysis
func TestAnalyzeStochastic(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		ohlcData := make([]OHLCData, 10)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: 1000,
			}
		}

		result := AnalyzeStochastic(ohlcData)

		if result.Signal != 0 {
			t.Errorf("expected signal 0, got %d", result.Signal)
		}
		if result.Zone != "NEUTRAL" {
			t.Errorf("expected zone NEUTRAL, got %s", result.Zone)
		}
		if len(result.Details) != 1 || result.Details[0] != "Insufficient data" {
			t.Errorf("expected 'Insufficient data' detail, got %v", result.Details)
		}
	})

	t.Run("overbought scenario", func(t *testing.T) {
		// Create data that will produce high %K (strong uptrend)
		ohlcData := make([]OHLCData, 50)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i)*2,
				High:  110 + float64(i)*2,
				Low:   95 + float64(i)*2,
				Close: 108 + float64(i)*2, // Close near high
				Volume: 1000,
			}
		}

		result := AnalyzeStochastic(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("oversold scenario", func(t *testing.T) {
		// Create data that will produce low %K (strong downtrend)
		ohlcData := make([]OHLCData, 50)
		for i := range ohlcData {
			ohlcData[i] = OHLCData{
				Open:  200 - float64(i)*2,
				High:  205 - float64(i)*2,
				Low:   190 - float64(i)*2,
				Close: 192 - float64(i)*2, // Close near low
				Volume: 1000,
			}
		}

		result := AnalyzeStochastic(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("signal clamping", func(t *testing.T) {
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

		result := AnalyzeStochastic(ohlcData)

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})
}

// TestStochasticResultStructure tests the structure of StochasticResult
func TestStochasticResultStructure(t *testing.T) {
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

	result := AnalyzeStochastic(ohlcData)

	// Check that values are populated
	if result.Values.K == 0 && result.Values.D == 0 {
		t.Error("expected non-zero K and D values")
	}

	// Check that zone is populated
	if result.Zone == "" {
		t.Error("expected zone to be populated")
	}

	// Check that details are populated
	if len(result.Details) == 0 {
		t.Error("expected details to be populated")
	}
}

// TestCalculateStochasticWithCustomParams tests Stochastic with custom parameters
func TestCalculateStochasticWithCustomParams(t *testing.T) {
	highs := make([]float64, 50)
	lows := make([]float64, 50)
	closes := make([]float64, 50)

	for i := range highs {
		highs[i] = 100 + float64(i)
		lows[i] = 90 + float64(i)
		closes[i] = 95 + float64(i)
	}

	// Custom parameters
	result := CalculateStochastic(highs, lows, closes, 10, 3, 3)

	// Should have valid results
	if len(result.K) == 0 {
		t.Error("expected non-empty K line")
	}
	if len(result.D) == 0 {
		t.Error("expected non-empty D line")
	}
}
