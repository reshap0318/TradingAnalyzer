package indicators

import (
	"math"
	"sort"

	"github.com/reshap/trading-bot/internal/config"
)

// SRLevel represents a support or resistance level
type SRLevel struct {
	Price       float64 // Average price level
	Type        string  // "SUPPORT" or "RESISTANCE"
	Touches     int     // Number of times price touched this level
	Strength    float64 // Level strength score (0-100)
	LastTouch   int64   // Timestamp of last touch
	ClusterSize int     // Number of swing points in this cluster
}

// SRResult holds Support & Resistance analysis result
type SRResult struct {
	Levels       []SRLevel // All valid S/R levels
	NearestSup   float64   // Nearest support level
	NearestRes   float64   // Nearest resistance level
	CurrentPrice float64   // Current price for reference
}

// AnalyzeSRWithParams analyzes Support & Resistance with custom parameters (LuxAlgo Logic)
func AnalyzeSRWithParams(ohlcData []OHLCData, closes []float64, params config.SRConfig) SRResult {
	// Validate input data
	if len(ohlcData) == 0 || len(closes) == 0 {
		return SRResult{
			Levels:       []SRLevel{},
			NearestSup:   0,
			NearestRes:   0,
			CurrentPrice: 0,
		}
	}

	// Get current price from last candle
	currentPrice := ohlcData[len(ohlcData)-1].Close

	// Check if we have enough data for lookback
	if len(ohlcData) < params.LOOKBACK_PERIODS {
		return SRResult{
			Levels:       []SRLevel{},
			NearestSup:   0,
			NearestRes:   0,
			CurrentPrice: currentPrice,
		}
	}

	// Use only LOOKBACK_PERIODS candles
	data := ohlcData[len(ohlcData)-params.LOOKBACK_PERIODS:]

	// STEP 1: Identify Swing Highs and Lows
	swingHighs := findSwingHighs(data, 5) // semakin tinggi lookback, semakin sedikit swing high/low yang terdeteksi
	swingLows := findSwingLows(data, 5)

	// STEP 2: Cluster similar levels
	resistanceClusters := clusterLevels(swingHighs, params.TOLERANCE)
	supportClusters := clusterLevels(swingLows, params.TOLERANCE)

	// STEP 3: Count touches and filter by MIN_TOUCHES
	validLevels := make([]SRLevel, 0)

	// Process resistance levels
	for _, cluster := range resistanceClusters {
		if cluster.Touches >= params.MIN_TOUCHES {
			level := SRLevel{
				Price:       cluster.AvgPrice,
				Type:        "RESISTANCE",
				Touches:     cluster.Touches,
				Strength:    calculateStrength(cluster.Touches, cluster.Recency),
				LastTouch:   cluster.LastTouch,
				ClusterSize: len(cluster.Points),
			}
			validLevels = append(validLevels, level)
		}
	}

	// Process support levels
	for _, cluster := range supportClusters {
		if cluster.Touches >= params.MIN_TOUCHES {
			level := SRLevel{
				Price:       cluster.AvgPrice,
				Type:        "SUPPORT",
				Touches:     cluster.Touches,
				Strength:    calculateStrength(cluster.Touches, cluster.Recency),
				LastTouch:   cluster.LastTouch,
				ClusterSize: len(cluster.Points),
			}
			validLevels = append(validLevels, level)
		}
	}

	// STEP 4: Sort by strength (strongest first)
	sort.Slice(validLevels, func(i, j int) bool {
		return validLevels[i].Strength > validLevels[j].Strength
	})

	// STEP 5: Find nearest support and resistance
	nearestSup := 0.0
	nearestRes := 0.0

	for _, level := range validLevels {
		if level.Type == "SUPPORT" && level.Price < currentPrice {
			if nearestSup == 0 || level.Price > nearestSup {
				nearestSup = level.Price
			}
		}
		if level.Type == "RESISTANCE" && level.Price > currentPrice {
			if nearestRes == 0 || level.Price < nearestRes {
				nearestRes = level.Price
			}
		}
	}

	return SRResult{
		Levels:       validLevels,
		NearestSup:   nearestSup,
		NearestRes:   nearestRes,
		CurrentPrice: currentPrice,
	}
}

// swingPoint represents a swing high or low
type swingPoint struct {
	Price     float64
	Timestamp int64
	Index     int
}

// cluster represents a group of similar swing points
type cluster struct {
	AvgPrice  float64
	Touches   int
	Recency   float64 // 0-1 score (1 = most recent)
	LastTouch int64
	Points    []swingPoint
}

// findSwingHighs finds swing highs in the data
func findSwingHighs(data []OHLCData, lookback int) []swingPoint {
	swingHighs := make([]swingPoint, 0)

	for i := lookback; i < len(data)-lookback; i++ {
		isSwingHigh := true
		currentHigh := data[i].High

		// Check if currentHigh is highest in the lookback window
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && data[j].High >= currentHigh {
				isSwingHigh = false
				break
			}
		}

		if isSwingHigh {
			swingHighs = append(swingHighs, swingPoint{
				Price:     currentHigh,
				Timestamp: data[i].Timestamp,
				Index:     i,
			})
		}
	}

	return swingHighs
}

// findSwingLows finds swing lows in the data
func findSwingLows(data []OHLCData, lookback int) []swingPoint {
	swingLows := make([]swingPoint, 0)

	for i := lookback; i < len(data)-lookback; i++ {
		isSwingLow := true
		currentLow := data[i].Low

		// Check if currentLow is lowest in the lookback window
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && data[j].Low <= currentLow {
				isSwingLow = false
				break
			}
		}

		if isSwingLow {
			swingLows = append(swingLows, swingPoint{
				Price:     currentLow,
				Timestamp: data[i].Timestamp,
				Index:     i,
			})
		}
	}

	return swingLows
}

// clusterLevels groups similar price levels together
func clusterLevels(points []swingPoint, tolerance float64) []cluster {
	if len(points) == 0 {
		return []cluster{}
	}

	// Sort points by price
	sorted := make([]swingPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Price < sorted[j].Price
	})

	clusters := make([]cluster, 0)
	currentCluster := cluster{
		Points: []swingPoint{sorted[0]},
	}

	for i := 1; i < len(sorted); i++ {
		// Check if this point is within tolerance of current cluster
		avgPrice := currentCluster.getAvgPrice()
		toleranceRange := avgPrice * tolerance

		if math.Abs(sorted[i].Price-avgPrice) <= toleranceRange {
			// Add to current cluster
			currentCluster.Points = append(currentCluster.Points, sorted[i])
		} else {
			// Finalize current cluster and start new one
			if len(currentCluster.Points) > 0 {
				currentCluster.finalize()
				clusters = append(clusters, currentCluster)
			}
			currentCluster = cluster{
				Points: []swingPoint{sorted[i]},
			}
		}
	}

	// Finalize last cluster
	if len(currentCluster.Points) > 0 {
		currentCluster.finalize()
		clusters = append(clusters, currentCluster)
	}

	return clusters
}

// getAvgPrice calculates average price of cluster points
func (c *cluster) getAvgPrice() float64 {
	if len(c.Points) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range c.Points {
		sum += p.Price
	}
	return sum / float64(len(c.Points))
}

// finalize calculates cluster statistics
func (c *cluster) finalize() {
	c.AvgPrice = c.getAvgPrice()
	c.Touches = len(c.Points)

	// Find last touch timestamp
	maxTime := int64(0)
	for _, p := range c.Points {
		if p.Timestamp > maxTime {
			maxTime = p.Timestamp
		}
	}
	c.LastTouch = maxTime

	// Calculate recency (0-1 score, 1 = most recent)
	// Assuming latest timestamp is around now
	c.Recency = 1.0 // Will be calculated relative to data
}

// calculateStrength calculates level strength score (0-100)
func calculateStrength(touches int, recency float64) float64 {
	// Weight: 70% touches, 30% recency
	touchScore := math.Min(float64(touches)*15, 100) // Max 100 at ~7 touches
	recencyScore := recency * 100

	strength := (touchScore * 0.7) + (recencyScore * 0.3)
	return math.Min(strength, 100) // Cap at 100
}
