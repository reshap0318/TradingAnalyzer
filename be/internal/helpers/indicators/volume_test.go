package indicators

import (
	"testing"
)

// TestAnalyzeVolume tests Volume analysis
func TestAnalyzeVolume(t *testing.T) {
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

		result := AnalyzeVolume(ohlcData)

		if result.Signal != 0 {
			t.Errorf("expected signal 0, got %d", result.Signal)
		}
		if result.VolumeRatio != 0 {
			t.Errorf("expected volume ratio 0, got %f", result.VolumeRatio)
		}
		if len(result.Details) != 1 || result.Details[0] != "Insufficient data" {
			t.Errorf("expected 'Insufficient data' detail, got %v", result.Details)
		}
	})

	t.Run("high volume up", func(t *testing.T) {
		// Create data with high volume and price up
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			volume := 1000.0
			if i >= 25 {
				volume = 3000.0 // Very high volume at the end
			}
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: volume,
			}
		}
		// Make sure last candle is up
		ohlcData[29].Close = 135
		ohlcData[28].Close = 132

		result := AnalyzeVolume(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Volume ratio should be > 1 for high volume
		if result.VolumeRatio < 1.0 {
			t.Errorf("expected volume ratio > 1, got %f", result.VolumeRatio)
		}

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("high volume down", func(t *testing.T) {
		// Create data with high volume and price down
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			volume := 1000.0
			if i >= 25 {
				volume = 3000.0 // Very high volume at the end
			}
			ohlcData[i] = OHLCData{
				Open:  200 - float64(i),
				High:  205 - float64(i),
				Low:   195 - float64(i),
				Close: 198 - float64(i),
				Volume: volume,
			}
		}
		// Make sure last candle is down
		ohlcData[29].Close = 165
		ohlcData[28].Close = 168

		result := AnalyzeVolume(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})

	t.Run("low volume", func(t *testing.T) {
		// Create data with decreasing volume
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			volume := 1000.0 - float64(i)*20 // Decreasing volume
			if volume < 100 {
				volume = 100
			}
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: volume,
			}
		}

		result := AnalyzeVolume(ohlcData)

		// Should have details populated
		if len(result.Details) == 0 {
			t.Error("expected details to be populated")
		}

		// Details should mention low volume
		hasLowVolume := false
		for _, detail := range result.Details {
			if detail == "Low volume" {
				hasLowVolume = true
				break
			}
		}
		if !hasLowVolume {
			t.Error("expected 'Low volume' detail")
		}
	})

	t.Run("signal clamping", func(t *testing.T) {
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

		result := AnalyzeVolume(ohlcData)

		// Signal should be clamped between -100 and 100
		if result.Signal > 100 || result.Signal < -100 {
			t.Errorf("signal not clamped properly, got %d", result.Signal)
		}
	})
}

// TestVolumeResultStructure tests the structure of VolumeResult
func TestVolumeResultStructure(t *testing.T) {
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

	result := AnalyzeVolume(ohlcData)

	// Check that volume values are populated
	if result.CurrentVolume == 0 {
		t.Error("expected CurrentVolume to be populated")
	}
	if result.AvgVolume == 0 {
		t.Error("expected AvgVolume to be populated")
	}

	// Check that details are populated
	if len(result.Details) == 0 {
		t.Error("expected details to be populated")
	}
}

// TestAnalyzeVolumeWithCustomParams tests Volume with custom parameters
func TestAnalyzeVolumeWithCustomParams(t *testing.T) {
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

	params := VolumeParameters{MAPeriod: 15}
	result := AnalyzeVolumeWithParams(ohlcData, params)

	// Should have valid results
	if result.CurrentVolume == 0 {
		t.Error("expected non-zero CurrentVolume")
	}
	if result.AvgVolume == 0 {
		t.Error("expected non-zero AvgVolume")
	}
}

// TestGetVolumeSignal tests GetVolumeSignal function
func TestGetVolumeSignal(t *testing.T) {
	t.Run("very high volume", func(t *testing.T) {
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			volume := 1000.0
			if i == 29 {
				volume = 5000.0 // Very high volume
			}
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: volume,
			}
		}

		signal := GetVolumeSignal(ohlcData)

		if signal.Level != "VERY_HIGH" {
			t.Errorf("expected VERY_HIGH volume level, got %s", signal.Level)
		}
		if signal.Ratio < 2.0 {
			t.Errorf("expected ratio >= 2.0, got %f", signal.Ratio)
		}
	})

	t.Run("low volume", func(t *testing.T) {
		ohlcData := make([]OHLCData, 30)
		for i := range ohlcData {
			volume := 1000.0 - float64(i)*30
			if volume < 100 {
				volume = 100
			}
			ohlcData[i] = OHLCData{
				Open:  100 + float64(i),
				High:  105 + float64(i),
				Low:   95 + float64(i),
				Close: 102 + float64(i),
				Volume: volume,
			}
		}

		signal := GetVolumeSignal(ohlcData)

		if signal.Level != "LOW" {
			t.Errorf("expected LOW volume level, got %s", signal.Level)
		}
	})

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

		signal := GetVolumeSignal(ohlcData)

		if signal.Level != "UNKNOWN" {
			t.Errorf("expected UNKNOWN volume level for insufficient data, got %s", signal.Level)
		}
	})
}
