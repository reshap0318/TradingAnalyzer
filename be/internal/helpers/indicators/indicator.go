package indicators

import (
	"github.com/reshap/trading-bot/internal/config"
)

// IndicatorResult holds unified result from any indicator
type IndicatorResult struct {
	Signal  int         `json:"signal"`  // Signal score (-100 to 100)
	Zone    string      `json:"zone"`    // Zone/position (e.g., "OVERBOUGHT", "NEUTRAL")
	Details []string    `json:"details"` // Analysis details
	Values  interface{} `json:"values"`  // Raw indicator values (optional)
}

// AnalyzeIndicatorWithConfig analyzes indicator with config-based parameters and caching support
// cacheKey: unique key for caching (e.g., "15m_1234567890")
func AnalyzeIndicatorWithConfig(
	indicatorKey string,
	ohlcData []OHLCData,
	closes []float64,
	cacheKey string,
	cfg *config.IndicatorsConfig,
) (*IndicatorResult, error) {
	if len(ohlcData) == 0 || len(closes) == 0 {
		return nil, ErrNoData
	}

	rsiParam := RSIParameters{
		Period:     cfg.RSI.PERIOD,
		Overbought: cfg.RSI.OVERBOUGHT,
		Oversold:   cfg.RSI.OVERSOLD,
	}

	macdParam := MACDParameters{
		Fast:   cfg.MACD.FAST,
		Slow:   cfg.MACD.SLOW,
		Signal: cfg.MACD.SIGNAL,
	}

	maParam := MAParameters{
		SMAPeriods: cfg.MOVING_AVERAGE.SMA_PERIODS,
		EMAPeriods: cfg.MOVING_AVERAGE.EMA_PERIODS,
	}

	stochasticParam := StochasticParameters{
		KPeriod:    cfg.STOCHASTIC.K_PERIOD,
		DPeriod:    cfg.STOCHASTIC.D_PERIOD,
		Smooth:     cfg.STOCHASTIC.SMOOTH,
		Overbought: cfg.STOCHASTIC.OVERBOUGHT,
		Oversold:   cfg.STOCHASTIC.OVERSOLD,
	}

	bbParam := BBParameters{
		Period:     cfg.BOLLINGER.PERIOD,
		StdDevMult: cfg.BOLLINGER.STD_DEV,
	}

	atrParam := ATRParameters{
		Period: cfg.ATR.PERIOD,
	}

	volumeParam := VolumeParameters{
		MAPeriod: cfg.VOLUME.MA_PERIOD,
	}

	switch indicatorKey {
	case "rsi":
		result := AnalyzeRSIWithParams(closes, rsiParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    result.Zone,
			Details: result.Details,
			Values:  map[string]float64{"value": result.Value},
		}, nil

	case "macd":
		result := AnalyzeMACDWithParams(closes, macdParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    "",
			Details: result.Details,
			Values:  result.Values,
		}, nil

	case "stochastic":
		// Use cached MA and MACD for trend neutralization (matching GitHub logic)
		maResult := AnalyzeMAWithCache(closes, cacheKey, maParam)
		macdResult := AnalyzeMACDWithCache(closes, cacheKey, macdParam)

		// Use trend-aware Stochastic analysis with config params
		// This matches GitHub's Trend Regime Detection in cryptoDecisionEngine.js
		result := AnalyzeStochasticWithTrendAndParams(ohlcData, maResult.Signal, macdResult.Signal, stochasticParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    result.Zone,
			Details: result.Details,
			Values:  result.Values,
		}, nil

	case "bollinger_bands":
		result := AnalyzeBollingerBandsWithParams(closes, bbParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    result.Position,
			Details: result.Details,
			Values:  result.Values,
		}, nil

	case "atr":
		result := AnalyzeATRWithParams(ohlcData, atrParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    result.Volatility,
			Details: result.Details,
			Values: map[string]float64{
				"atr":        result.ATR,
				"atrPercent": result.ATRPercent,
				"atrRatio":   result.ATRRatio,
			},
		}, nil

	case "volume":
		result := AnalyzeVolumeWithParams(ohlcData, volumeParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    "",
			Details: result.Details,
			Values: map[string]float64{
				"currentVolume": result.CurrentVolume,
				"avgVolume":     result.AvgVolume,
				"volumeRatio":   result.VolumeRatio,
			},
		}, nil

	case "moving_average":
		// Use config-based MA analysis
		result := AnalyzeMAWithCache(closes, cacheKey, maParam)

		return &IndicatorResult{
			Signal:  result.Signal,
			Zone:    result.Trend,
			Details: result.Details,
			Values: map[string]float64{
				"sma20":  result.Values.SMA20,
				"sma50":  result.Values.SMA50,
				"sma200": result.Values.SMA200,
				"ema12":  result.Values.EMA12,
				"ema26":  result.Values.EMA26,
			},
		}, nil

	case "candle_patterns":
		if len(ohlcData) < 6 {
			return nil, ErrInsufficientData
		}

		opens := make([]float64, len(ohlcData))
		highs := make([]float64, len(ohlcData))
		lows := make([]float64, len(ohlcData))
		timestamps := make([]int64, len(ohlcData))

		for i, d := range ohlcData {
			opens[i] = d.Open
			highs[i] = d.High
			lows[i] = d.Low
			timestamps[i] = d.Timestamp
		}

		// Analyze pattern history for last 5 closed candles (matching GitHub)
		historyResult := AnalyzeCandlePatternHistory(opens, highs, lows, closes, timestamps, 5)

		return &IndicatorResult{
			Signal:  historyResult.TrendScore,
			Zone:    "",
			Details: historyResult.Details,
			Values: map[string]int{
				"bullishCount": historyResult.TotalBullish,
				"bearishCount": historyResult.TotalBearish,
				"neutralCount": historyResult.TotalNeutral,
			},
		}, nil

	case "trend_bonus":
		// Trend Bonus uses MA and MACD signals internally with config
		maResult := AnalyzeMAWithCache(closes, cacheKey, maParam)
		macdResult := AnalyzeMACDWithCache(closes, cacheKey, macdParam)

		trendResult := AnalyzeTrendBonus(maResult.Signal, macdResult.Signal)
		return &IndicatorResult{
			Signal:  trendResult.Signal,
			Zone:    trendResult.Trend,
			Details: trendResult.Details,
			Values: map[string]interface{}{
				"maSignal":   maResult.Signal,
				"macdSignal": macdResult.Signal,
			},
		}, nil

	default:
		return nil, ErrUnknownIndicator
	}
}

// Errors for indicator analysis
var (
	ErrNoData           = &indicatorError{"no data provided"}
	ErrInsufficientData = &indicatorError{"insufficient data for analysis"}
	ErrUnknownIndicator = &indicatorError{"unknown indicator key"}
)

type indicatorError struct {
	message string
}

func (e *indicatorError) Error() string {
	return e.message
}
