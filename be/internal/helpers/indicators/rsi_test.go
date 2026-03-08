package indicators

import (
	"testing"
)

// TestCalculateRSI tests RSI calculation
func TestCalculateRSI(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		data := make([]float64, 10)
		result := CalculateRSI(data, 14)

		if len(result) != 0 {
			t.Errorf("expected empty result, got length %d", len(result))
		}
	})

	t.Run("sufficient data", func(t *testing.T) {
		// Create 30 data points
		data := make([]float64, 30)
		for i := range data {
			data[i] = float64(i + 1)
		}

		result := CalculateRSI(data, 14)

		// Should have 30 - 14 = 16 values
		expectedLen := 30 - 14
		if len(result) != expectedLen {
			t.Errorf("expected length %d, got %d", expectedLen, len(result))
		}

		// RSI should be between 0 and 100
		for i, rsi := range result {
			if rsi < 0 || rsi > 100 {
				t.Errorf("RSI value at index %d is out of range: %f", i, rsi)
			}
		}
	})
}

// TestAnalyzeRSI tests RSI analysis
func TestAnalyzeRSI(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		data := make([]float64, 15)
		result := AnalyzeRSI(data)

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
		// Create data that will produce high RSI (strong uptrend)
		data := make([]float64, 50)
		for i := range data {
			data[i] = 100 + float64(i)*2 // Strong upward movement
		}

		result := AnalyzeRSI(data)

		// Should have some details
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}
	})

	t.Run("signal clamping", func(t *testing.T) {
		data := make([]float64, 50)
		for i := range data {
			data[i] = float64(i + 1)
		}

		result := AnalyzeRSI(data)

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})
}

// TestRSIResultStructure tests the structure of RSIResult
func TestRSIResultStructure(t *testing.T) {
	data := make([]float64, 50)
	for i := range data {
		data[i] = 50 + float64(i)*0.5
	}

	result := AnalyzeRSI(data)

	// Check that value is populated
	if result.Value == 0 {
		t.Error("expected non-zero RSI value")
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
