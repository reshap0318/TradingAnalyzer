package indicators

import (
	"testing"
)

// TestCalculateBollingerBands tests Bollinger Bands calculation
func TestCalculateBollingerBands(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		closes := []float64{100, 102, 101, 103, 102}

		result := CalculateBollingerBands(closes, 20, 2.0)

		if len(result.Upper) != 0 || len(result.Middle) != 0 || len(result.Lower) != 0 {
			t.Errorf("expected empty result, got Upper length %d", len(result.Upper))
		}
	})

	t.Run("sufficient data", func(t *testing.T) {
		// Create 30 data points
		closes := make([]float64, 30)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := CalculateBollingerBands(closes, 20, 2.0)

		// Should have 30 - 20 + 1 = 11 values
		expectedLen := 30 - 20 + 1
		if len(result.Upper) != expectedLen {
			t.Errorf("expected Upper length %d, got %d", expectedLen, len(result.Upper))
		}
		if len(result.Middle) != expectedLen {
			t.Errorf("expected Middle length %d, got %d", expectedLen, len(result.Middle))
		}
		if len(result.Lower) != expectedLen {
			t.Errorf("expected Lower length %d, got %d", expectedLen, len(result.Lower))
		}
		if len(result.Bandwidth) != expectedLen {
			t.Errorf("expected Bandwidth length %d, got %d", expectedLen, len(result.Bandwidth))
		}
	})

	t.Run("band order", func(t *testing.T) {
		// Upper > Middle > Lower should always be true
		closes := make([]float64, 30)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := CalculateBollingerBands(closes, 20, 2.0)

		for i := range result.Upper {
			if result.Upper[i] <= result.Middle[i] {
				t.Errorf("Upper band should be > Middle band at index %d", i)
			}
			if result.Middle[i] <= result.Lower[i] {
				t.Errorf("Middle band should be > Lower band at index %d", i)
			}
		}
	})

	t.Run("constant data", func(t *testing.T) {
		// With constant data, bands should converge
		closes := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}

		result := CalculateBollingerBands(closes, 20, 2.0)

		if len(result.Upper) > 0 {
			// With constant data, stdDev is 0, so bands should be at same level
			if result.Upper[len(result.Upper)-1] != result.Middle[len(result.Middle)-1] {
				t.Error("With constant data, Upper should equal Middle")
			}
			if result.Lower[len(result.Lower)-1] != result.Middle[len(result.Middle)-1] {
				t.Error("With constant data, Lower should equal Middle")
			}
		}
	})
}

// TestAnalyzeBollingerBands tests Bollinger Bands analysis
func TestAnalyzeBollingerBands(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		closes := make([]float64, 20)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := AnalyzeBollingerBands(closes)

		if result.Signal != 0 {
			t.Errorf("expected signal 0, got %d", result.Signal)
		}
		if result.Position != "NEUTRAL" {
			t.Errorf("expected position NEUTRAL, got %s", result.Position)
		}
		if len(result.Details) != 1 || result.Details[0] != "Insufficient data" {
			t.Errorf("expected 'Insufficient data' detail, got %v", result.Details)
		}
	})

	t.Run("price at upper band", func(t *testing.T) {
		// Create data with price at upper band
		closes := make([]float64, 50)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := AnalyzeBollingerBands(closes)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("price at lower band", func(t *testing.T) {
		// Create data with strong downtrend (price at lower band)
		closes := make([]float64, 50)
		for i := range closes {
			closes[i] = 200 - float64(i)*2
		}

		result := AnalyzeBollingerBands(closes)

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
		closes := make([]float64, 50)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := AnalyzeBollingerBands(closes)

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("percentB calculation", func(t *testing.T) {
		closes := make([]float64, 50)
		for i := range closes {
			closes[i] = 100 + float64(i)
		}

		result := AnalyzeBollingerBands(closes)

		// %B should be between 0 and 1 for price within bands
		// Can be > 1 or < 0 if price is outside bands
		if result.PercentB < -1 || result.PercentB > 2 {
			t.Errorf("percentB seems unreasonable: %f", result.PercentB)
		}
	})
}

// TestBBResultStructure tests the structure of BBResult
func TestBBResultStructure(t *testing.T) {
	closes := make([]float64, 50)
	for i := range closes {
		closes[i] = 100 + float64(i)
	}

	result := AnalyzeBollingerBands(closes)

	// Check that values are populated
	if result.Values.Upper == 0 || result.Values.Middle == 0 || result.Values.Lower == 0 {
		t.Error("expected band values to be populated")
	}

	// Check that position is populated
	if result.Position == "" {
		t.Error("expected position to be populated")
	}

	// Check that details are populated
	if len(result.Details) == 0 {
		t.Error("expected details to be populated")
	}
}

// TestCalculateBollingerBandsWithCustomParams tests with custom parameters
func TestCalculateBollingerBandsWithCustomParams(t *testing.T) {
	closes := make([]float64, 50)
	for i := range closes {
		closes[i] = 100 + float64(i)
	}

	// Custom parameters: period=14, stdDev=2.5
	result := CalculateBollingerBands(closes, 14, 2.5)

	// Should have valid results
	if len(result.Upper) == 0 {
		t.Error("expected non-empty Upper band")
	}
	if len(result.Middle) == 0 {
		t.Error("expected non-empty Middle band")
	}
	if len(result.Lower) == 0 {
		t.Error("expected non-empty Lower band")
	}
}
