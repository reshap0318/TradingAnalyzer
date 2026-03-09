package service

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/helpers/indicators"
	"github.com/reshap/trading-bot/internal/models"
)

// SignalRawGet retrieves raw OHLCV data for multiple timeframes
func (s *Services) SignalRawGet(ctx *gin.Context, req *dtos.SignalRawRequest) (res *dtos.SignalRawResponse, err error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	if len(req.Timeframes) == 0 {
		return nil, fmt.Errorf("at least one timeframe is required")
	}

	// Fetch klines for each timeframe in parallel
	type klineResult struct {
		timeframe string
		klines    []binance.KlineInfo
		err       error
	}

	resultChan := make(chan klineResult, len(req.Timeframes))

	// Launch goroutines for each timeframe
	for _, tf := range req.Timeframes {
		go func(name string, limit int) {
			klines, err := s.BinanceClient.GetKlines(req.Symbol, name, limit)
			resultChan <- klineResult{
				timeframe: name,
				klines:    klines,
				err:       err,
			}
		}(tf.Name, tf.Limit)
	}

	// Collect results
	timeframesData := make([]dtos.TimeframeRawData, 0, len(req.Timeframes))
	var firstErr error

	for i := 0; i < len(req.Timeframes); i++ {
		result := <-resultChan
		if result.err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to get klines for %s: %v", result.timeframe, result.err)
			}
			continue
		}

		// Convert klines to raw data (pure OHLCV, no calculations)
		raws := make([]dtos.RawData, len(result.klines))
		for j, k := range result.klines {
			raws[j] = dtos.RawData{
				Timestamp: k.OpenTime,
				Open:      k.Open,
				High:      k.High,
				Low:       k.Low,
				Close:     k.Close,
				Volume:    k.Volume,
			}
		}

		timeframesData = append(timeframesData, dtos.TimeframeRawData{
			Timeframe: result.timeframe,
			Raws:      raws,
		})
	}

	if firstErr != nil && len(timeframesData) == 0 {
		return nil, firstErr
	}

	return &dtos.SignalRawResponse{
		Symbol:     req.Symbol,
		Timeframes: timeframesData,
	}, nil
}

// SignalAnalyze analyzes market data and generates trading signal
// Request: symbol + strategy_id (optional) + amount (default 50)
// If strategy_id not provided, uses active strategy
func (s *Services) SignalAnalyze(ctx *gin.Context, req *dtos.SignalAnalyzeRequest) (res *dtos.SignalAnalyzeResponse, err error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Default capital
	tradCapital := req.Capital
	if tradCapital <= 0 {
		tradCapital = 50.0 // Default $50
	}

	// Get strategy
	var strategy *dtos.StrategyData
	if req.StrategyID > 0 {
		strategy, err = s.StrategyGetByID(ctx, req.StrategyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get strategy: %v", err)
		}
	} else {
		strategy, err = s.StrategyGetActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active strategy: %v", err)
		}
	}

	//Fetch thresholds
	thresholds, err := s.repo.Threshold.FindAll(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thresholds: %w", err)
	}

	// Get timeframes from strategy
	timeframeKlines := make([]binance.MultiKlineRequest, 0, len(strategy.Timeframes))
	for _, tf := range strategy.Timeframes {
		timeframeKlines = append(timeframeKlines, binance.MultiKlineRequest{
			Interval: tf.TimeframeName,
			Limit:    150,
		})
	}

	// Fetch klines for all timeframes in parallel
	binanceData, err := s.BinanceClient.GetMultiKlines(req.Symbol, timeframeKlines)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %v", err)
	}

	// Build final response
	return s.signalAnalyzeCalculate(req.Symbol, tradCapital, strategy, binanceData, thresholds)
}

func (s *Services) signalAnalyzeCalculate(
	symbol string,
	tradCapital float64,
	strategy *dtos.StrategyData,
	binanceData map[string][]binance.KlineInfo,
	thresholds []models.Threshold,
) (*dtos.SignalAnalyzeResponse, error) {
	config := s.getConfigMM(strategy)
	finalSignal := 0.0
	breakdown := make([]dtos.TimeframeSignalData, 0)

	for _, tf := range strategy.Timeframes {
		klines, exists := binanceData[tf.TimeframeName]
		if !exists || len(klines) == 0 {
			continue
		}

		cacheKey := fmt.Sprintf("%s_%d", tf.TimeframeName, time.Now().UnixMilli())

		// Convert to OHLCV and closes
		ohlcData, closes := convertKlinesToOHLCV(klines)

		var tfSignal float64
		indicatorBreakdown := make([]dtos.IndicatorBreakdown, 0)
		for _, iw := range strategy.IndicatorWeights {
			result, err := indicators.AnalyzeIndicatorWithConfig(
				iw.IndicatorDetail.Indicator,
				ohlcData,
				closes,
				cacheKey,
				&s.cfg.INDICATORS,
			)

			if err != nil {
				continue
			}

			// Calculate contribution
			contribution := float64(result.Signal) * iw.Weight
			tfSignal += contribution

			// Build indicator detail
			indicatorDetail := dtos.IndicatorBreakdown{
				Name:         iw.IndicatorDetail.Name,
				RawSignal:    result.Signal,
				Weight:       iw.Weight,
				Contribution: helpers.RoundFloat(contribution, 3),
				Details:      result.Details,
			}

			// Add extra fields based on indicator type
			if result.Values != nil {
				if values, ok := result.Values.(map[string]interface{}); ok {
					for k, v := range values {
						if k == "value" {
							indicatorDetail.Value = v
						}
					}
				}
			}

			// Add zone if present
			if result.Zone != "" {
				indicatorDetail.Zone = result.Zone
			}

			indicatorBreakdown = append(indicatorBreakdown, indicatorDetail)
		}

		// ✅ Weight already available from interval.Timeframes
		tfWeight := tf.Weight

		// Calculate contribution to final signal
		tfContribution := tfSignal * tfWeight
		finalSignal += tfContribution

		// Classify timeframe sentiment using threshold (with decimal precision)
		signal, _ := getCategoryFromThreshold(tfSignal, thresholds)

		// Store timeframe breakdown with indicators
		breakdown = append(breakdown, dtos.TimeframeSignalData{
			Timeframe:    tf.TimeframeName,
			Trend:        signal,
			RawSignal:    helpers.RoundFloat(tfSignal, 3),
			Weight:       tfWeight,
			Contribution: helpers.RoundFloat(tfContribution, 3),
			Indicator:    indicatorBreakdown,
		})
	}

	signal, action := getCategoryFromThreshold(finalSignal, thresholds)
	confidence := math.Abs(finalSignal)

	// Check signal validity based on MIN_CONFIDENCE
	minConfidence := float64(config.MIN_CONFIDENCE)
	signalValid := confidence >= minConfidence

	var tradingPlan *dtos.TradingPlan
	var currentPrice float64

	primaryKlines := binanceData[strategy.PrimaryTF]
	if len(primaryKlines) > 0 {
		currentPrice = primaryKlines[len(primaryKlines)-1].Close
	} else {
		return nil, fmt.Errorf("failed to fetch price data for symbol %s on timeframe %s", symbol, strategy.PrimaryTF)
	}

	if signal == "BUY" || signal == "SELL" {
		tradCapital = tradCapital * 0.8
	}

	tradingPlan = s.buildTradingPlan(currentPrice, tradCapital, action, strategy, primaryKlines, config)

	return &dtos.SignalAnalyzeResponse{
		Symbol:           symbol,
		PrimaryTimeframe: strategy.PrimaryTF,
		Timestamp:        time.Now(),
		Signal: dtos.SignalInfo{
			Valid:       signalValid,
			Signal:      signal,
			CurentPrice: currentPrice,
			TradingPlan: tradingPlan,
		},
		Scoring: dtos.ScoringBreakdown{
			TotalScore: helpers.RoundFloat(finalSignal, 3),
			Confidence: helpers.RoundFloat(confidence, 3),
			Breakdown:  breakdown,
		},
	}, nil
}

func (s *Services) getConfigMM(strategy *dtos.StrategyData) (mmConfig *config.MMConfig) {
	mmConfig = &config.MMConfig{
		MIN_CONFIDENCE:         strategy.MoneyManagement.MIN_CONFIDENCE,
		MAX_DAILY_TRADES:       strategy.MoneyManagement.MAX_DAILY_TRADES,
		MAX_DAILY_LOSS_PERCENT: strategy.MoneyManagement.MAX_DAILY_LOSS_PERCENT,
		MAX_DAILY_LOSS_COUNT:   strategy.MoneyManagement.MAX_DAILY_LOSS_COUNT,
		RISK_REWARD_RATIO:      strategy.MoneyManagement.RISK_REWARD_RATIO,
		RISK_REWARD_TARGET:     strategy.MoneyManagement.RISK_REWARD_TARGET,
		MAX_POSITION_SIZE:      strategy.MoneyManagement.MAX_POSITION_SIZE,
		MAX_RISK_PER_TRADE:     strategy.MoneyManagement.MAX_RISK_PER_TRADE,
		LEVERAGE:               strategy.MoneyManagement.LEVERAGE,
		IS_AGRESSIVE:           strategy.MoneyManagement.IS_AGRESSIVE,
		ORDER_EXPIRATION_HOURS: strategy.MoneyManagement.ORDER_EXPIRATION_HOURS,
	}

	// Override with global config if not set in strategy
	if mmConfig.MIN_CONFIDENCE == 0 {
		mmConfig.MIN_CONFIDENCE = s.cfg.MM.MIN_CONFIDENCE
	}
	if mmConfig.MAX_DAILY_TRADES == 0 {
		mmConfig.MAX_DAILY_TRADES = s.cfg.MM.MAX_DAILY_TRADES
	}
	if mmConfig.MAX_DAILY_LOSS_PERCENT == 0 {
		mmConfig.MAX_DAILY_LOSS_PERCENT = s.cfg.MM.MAX_DAILY_LOSS_PERCENT
	}
	if mmConfig.MAX_DAILY_LOSS_COUNT == 0 {
		mmConfig.MAX_DAILY_LOSS_COUNT = s.cfg.MM.MAX_DAILY_LOSS_COUNT
	}
	if mmConfig.RISK_REWARD_RATIO == 0 {
		mmConfig.RISK_REWARD_RATIO = s.cfg.MM.RISK_REWARD_RATIO
	}
	if mmConfig.RISK_REWARD_TARGET == 0 {
		mmConfig.RISK_REWARD_TARGET = s.cfg.MM.RISK_REWARD_TARGET
	}
	if mmConfig.MAX_POSITION_SIZE == 0 {
		mmConfig.MAX_POSITION_SIZE = s.cfg.MM.MAX_POSITION_SIZE
	}
	if mmConfig.MAX_RISK_PER_TRADE == 0 {
		mmConfig.MAX_RISK_PER_TRADE = s.cfg.MM.MAX_RISK_PER_TRADE
	}
	if mmConfig.LEVERAGE == 0 {
		mmConfig.LEVERAGE = s.cfg.MM.LEVERAGE
	}
	if mmConfig.ORDER_EXPIRATION_HOURS == 0 {
		mmConfig.ORDER_EXPIRATION_HOURS = s.cfg.MM.ORDER_EXPIRATION_HOURS
	}
	return
}

func (s *Services) buildTradingPlan(
	currentPrice float64,
	tradingCapital float64,
	action string,
	strategy *dtos.StrategyData,
	primaryKlines []binance.KlineInfo,
	config *config.MMConfig,
) *dtos.TradingPlan {
	bufferPercent := 0.015        // 1.5% buffer for S/R levels
	fallbackBufferPercent := 0.03 // 3% fallback buffer if no S/R data
	leverage := config.LEVERAGE
	isAggressive := config.IS_AGRESSIVE

	if currentPrice <= 0 || action == "WAIT" || len(primaryKlines) == 0 {
		return &dtos.TradingPlan{
			Mode:          "WAIT",
			Entries:       make([]dtos.TradingPlanEntry, 0),
			BufferPercent: helpers.RoundFloat(bufferPercent*100, 3),
		}
	}

	ohlcData, closes := convertKlinesToOHLCV(primaryKlines)

	srResult := indicators.AnalyzeSRWithParams(ohlcData, closes, s.cfg.INDICATORS.SUPPORT_RESIST)

	var tp, sl float64
	var entries []dtos.TradingPlanEntry

	if action == "BUY" || action == "STRONG_BUY" {
		// For BUY: TP = Resistance, SL = Support (with buffer)
		resistance := srResult.NearestRes
		support := srResult.NearestSup

		if resistance == 0 {
			resistance = currentPrice * (1 + fallbackBufferPercent) // Fallback: 2% above
		}
		if support == 0 {
			support = currentPrice * (1 - fallbackBufferPercent) // Fallback: 2% below
		}

		tp = resistance
		sl = support * (1 - bufferPercent)

		if isAggressive {
			// AGGRESSIVE MODE: Multiple entries (50% now + 50% pullback)
			entries = make([]dtos.TradingPlanEntry, 0, 2)

			// Entry 1: 50% at current price
			entry1Price := currentPrice
			entry1Value := tradingCapital * 0.5
			leveragedValue1 := entry1Value * float64(leverage)
			entry1Qty := leveragedValue1 / entry1Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entry1Price,
				PositionSize:  "50%",
				PositionValue: entry1Value,
				PositionQty:   entry1Qty,
			})

			// Entry 2: 50% at pullback to support
			entry2Price := support * (1 + bufferPercent)
			entry2Value := tradingCapital * 0.5
			leveragedValue2 := entry2Value * float64(leverage)
			entry2Qty := leveragedValue2 / entry2Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   2,
				EntryPrice:    entry2Price,
				PositionSize:  "50%",
				PositionValue: entry2Value,
				PositionQty:   entry2Qty,
			})

		} else {
			// CONSERVATIVE MODE: Single entry near current price
			// Uses small buffer below current price for slight discount
			entries = make([]dtos.TradingPlanEntry, 0, 1)

			// [OLD] Entry at support + buffer (rollback jika dibutuhkan)
			entryPrice := support * (1 + bufferPercent)
			// entryPrice := currentPrice * (1 - bufferPercent/3) // ~0.5% below current price
			entryValue := tradingCapital
			leveragedValue := entryValue * float64(leverage)
			entryQty := leveragedValue / entryPrice

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entryPrice,
				PositionSize:  "100%",
				PositionValue: entryValue,
				PositionQty:   entryQty,
			})
		}

	} else if action == "SELL" || action == "STRONG_SELL" {
		// For SELL: TP = Support, SL = Resistance (with buffer)
		resistance := srResult.NearestRes
		support := srResult.NearestSup

		if resistance == 0 {
			resistance = currentPrice * (1 + fallbackBufferPercent) // Fallback: 2% above
		}
		if support == 0 {
			support = currentPrice * (1 - fallbackBufferPercent) // Fallback: 2% below
		}

		tp = support
		sl = resistance * (1 + bufferPercent)

		if isAggressive {
			// AGGRESSIVE MODE: Multiple entries (50% now + 50% pullback)
			entries = make([]dtos.TradingPlanEntry, 0, 2)

			// Entry 1: 50% at current price
			entry1Price := currentPrice
			entry1Value := tradingCapital * 0.5
			leveragedValue1 := entry1Value * float64(leverage)
			entry1Qty := leveragedValue1 / entry1Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entry1Price,
				PositionSize:  "50%",
				PositionValue: entry1Value,
				PositionQty:   entry1Qty,
			})

			// Entry 2: 50% at pullback to resistance
			entry2Price := resistance * (1 - bufferPercent)
			entry2Value := tradingCapital * 0.5
			leveragedValue2 := entry2Value * float64(leverage)
			entry2Qty := leveragedValue2 / entry2Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   2,
				EntryPrice:    entry2Price,
				PositionSize:  "50%",
				PositionValue: entry2Value,
				PositionQty:   entry2Qty,
			})

		} else {
			// CONSERVATIVE MODE: Single entry near current price
			// Uses small buffer above current price for slight discount
			entries = make([]dtos.TradingPlanEntry, 0, 1)

			// [OLD] Entry at resistance - buffer (rollback jika dibutuhkan)
			entryPrice := resistance * (1 - bufferPercent)
			// entryPrice := currentPrice * (1 + bufferPercent/3) // ~0.5% above current price
			entryValue := tradingCapital
			leveragedValue := entryValue * float64(leverage)
			entryQty := leveragedValue / entryPrice

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entryPrice,
				PositionSize:  "100%",
				PositionValue: entryValue,
				PositionQty:   entryQty,
			})
		}
	} else {
		// Unknown action - return WAIT plan
		return &dtos.TradingPlan{
			Mode:          "WAIT",
			Entries:       make([]dtos.TradingPlanEntry, 0),
			BufferPercent: helpers.RoundFloat(bufferPercent*100, 3),
		}
	}

	// Calculate Risk/Reward ratio
	var riskRewardRatio float64
	if len(entries) > 0 {
		// Calculate average entry price (weighted by position value)
		var totalValue float64
		var weightedSum float64
		for _, entry := range entries {
			totalValue += entry.PositionValue
			weightedSum += entry.PositionValue * entry.EntryPrice
		}

		var avgEntryPrice float64
		if totalValue > 0 {
			avgEntryPrice = weightedSum / totalValue
		}

		if action == "BUY" || action == "STRONG_BUY" {
			reward := tp - avgEntryPrice
			risk := avgEntryPrice - sl
			if risk > 0 {
				riskRewardRatio = helpers.RoundFloat(reward/risk, 2)
			}
		} else {
			reward := avgEntryPrice - tp
			risk := sl - avgEntryPrice
			if risk > 0 {
				riskRewardRatio = helpers.RoundFloat(reward/risk, 2)
			}
		}
	}

	return &dtos.TradingPlan{
		Mode:            getMode(isAggressive),
		Entries:         entries,
		TakeProfit:      tp,
		StopLoss:        sl,
		RiskRewardRatio: riskRewardRatio,
		BufferPercent:   bufferPercent * 100,
	}
}

func getMode(isAggressive bool) string {
	if isAggressive {
		return "AGGRESSIVE"
	}
	return "CONSERVATIVE"
}

// convertKlinesToOHLCV converts binance KlineInfo to indicators.OHLCData
func convertKlinesToOHLCV(klines []binance.KlineInfo) ([]indicators.OHLCData, []float64) {
	ohlcData := make([]indicators.OHLCData, len(klines))
	closes := make([]float64, len(klines))

	for i, k := range klines {
		ohlcData[i] = indicators.OHLCData{
			Timestamp: k.OpenTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
		closes[i] = k.Close
	}

	return ohlcData, closes
}

func getCategoryFromThreshold(signal float64, thresholds []models.Threshold) (string, string) {
	for _, threshold := range thresholds {
		// Use float comparison for decimal precision
		if signal >= float64(threshold.MinValue) && signal < float64(threshold.MaxValue) {
			return threshold.Category, threshold.Action
		}
	}
	// Default to WAIT if no threshold matches
	return "WAIT", "WAIT"
}
