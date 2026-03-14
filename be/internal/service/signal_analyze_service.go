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
			Limit:    300,
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

	signal, _ := getCategoryFromThreshold(finalSignal, thresholds)
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

	tradingPlan = s.buildTradingPlan(currentPrice, tradCapital, signal, primaryKlines, config)

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
		RISK_ENTRY_BUFFER:      strategy.MoneyManagement.RISK_ENTRY_BUFFER,
		MAX_POSITION_SIZE:      strategy.MoneyManagement.MAX_POSITION_SIZE,
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
	if mmConfig.RISK_ENTRY_BUFFER == 0 {
		mmConfig.RISK_ENTRY_BUFFER = s.cfg.MM.RISK_ENTRY_BUFFER
	}
	if mmConfig.MAX_POSITION_SIZE == 0 {
		mmConfig.MAX_POSITION_SIZE = s.cfg.MM.MAX_POSITION_SIZE
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
	signal string,
	primaryKlines []binance.KlineInfo,
	config *config.MMConfig,
) *dtos.TradingPlan {
	// Use RISK_ENTRY_BUFFER from config (convert to decimal, e.g., 0.5% = 0.005)
	bufferPercent := float64(config.RISK_ENTRY_BUFFER)
	fallbackBufferPercent := bufferPercent * 2 // Fallback buffer is 2x entry buffer
	leverage := config.LEVERAGE
	isAggressive := config.IS_AGRESSIVE

	signalStrength := 1.0 // Default 100%
	if signal == "BUY" || signal == "SELL" || signal == "WAIT" {
		signalStrength = 0.8 // Reduce to 80% for regular signals
	}

	tradingCapital = tradingCapital * float64(config.MAX_POSITION_SIZE) * signalStrength

	if currentPrice <= 0 || len(primaryKlines) == 0 {
		return &dtos.TradingPlan{
			Mode:          "WAIT",
			Entries:       make([]dtos.TradingPlanEntry, 0),
			BufferPercent: helpers.RoundFloat(bufferPercent*100, 3),
		}
	}

	ohlcData, closes := convertKlinesToOHLCV(primaryKlines)

	srResult := indicators.AnalyzeSRWithParams(ohlcData, closes, s.cfg.INDICATORS.SUPPORT_RESIST)
	atrResult := indicators.AnalyzeATRWithConfig(ohlcData, s.cfg)
	atrValue := atrResult.ATR
	if atrValue == 0 {
		atrValue = currentPrice * 0.01 // Fallback to 1% if ATR calculation fails
	}

	// Convert config bufferPercent to an ATR multiplier (e.g., 0.0075 -> 0.75x ATR)
	atrMultiplier := bufferPercent * 100.0
	if atrMultiplier < 0.2 {
		atrMultiplier = 0.2 // Minimum 0.2x ATR
	}

	var tp, sl, support, resistance float64
	var entries []dtos.TradingPlanEntry

	switch signal {
	case "BUY", "STRONG_BUY", "WAIT":
		// For BUY: Entry near support, TP near resistance, SL below support
		resistance = srResult.NearestRes
		support = srResult.NearestSup

		if resistance == 0 {
			resistance = currentPrice * (1 + fallbackBufferPercent)
		}
		if support == 0 {
			support = currentPrice * (1 - fallbackBufferPercent)
		}

		// Ensure support < currentPrice < resistance
		if support >= currentPrice {
			support = currentPrice * (1 - fallbackBufferPercent)
		}
		if resistance <= currentPrice {
			resistance = currentPrice * (1 + fallbackBufferPercent)
		}

		// Calculate dynamic buffer using ATR
		buf := atrValue * atrMultiplier
		slBuf := atrValue * (atrMultiplier * 2.0)

		// TP just below resistance, SL just below support with ATR volatility buffer
		tp = resistance - buf
		sl = support - slBuf
		entryBase := support + buf
		if entryBase >= currentPrice {
			entryBase = currentPrice - buf
		}

		if isAggressive {
			// AGGRESSIVE MODE: 50% at current price + 50% at pullback near support
			entries = make([]dtos.TradingPlanEntry, 0, 2)

			entry1Price := currentPrice
			entry1Value := tradingCapital * 0.5
			entry1Qty := (entry1Value * float64(leverage)) / entry1Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entry1Price,
				PositionSize:  "50%",
				PositionValue: entry1Value,
				PositionQty:   entry1Qty,
			})

			entry2Price := entryBase
			entry2Value := tradingCapital * 0.5
			entry2Qty := (entry2Value * float64(leverage)) / entry2Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   2,
				EntryPrice:    entry2Price,
				PositionSize:  "50%",
				PositionValue: entry2Value,
				PositionQty:   entry2Qty,
			})

		} else {
			// CONSERVATIVE MODE: Single entry near support
			entries = make([]dtos.TradingPlanEntry, 0, 1)

			entryPrice := entryBase
			entryValue := tradingCapital
			entryQty := (entryValue * float64(leverage)) / entryPrice

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entryPrice,
				PositionSize:  "100%",
				PositionValue: entryValue,
				PositionQty:   entryQty,
			})
		}
	case "SELL", "STRONG_SELL":
		// For SELL: Entry near resistance, TP near support, SL above resistance
		resistance = srResult.NearestRes
		support = srResult.NearestSup

		if resistance == 0 {
			resistance = currentPrice * (1 + fallbackBufferPercent)
		}
		if support == 0 {
			support = currentPrice * (1 - fallbackBufferPercent)
		}

		// Ensure support < currentPrice < resistance
		if support >= currentPrice {
			support = currentPrice * (1 - fallbackBufferPercent)
		}
		if resistance <= currentPrice {
			resistance = currentPrice * (1 + fallbackBufferPercent)
		}

		// Calculate dynamic buffer using ATR
		buf := atrValue * atrMultiplier
		slBuf := atrValue * (atrMultiplier * 2.0)

		// TP just above support, SL just above resistance with ATR volatility buffer
		tp = support + buf
		sl = resistance + slBuf
		entryBase := resistance - buf
		if entryBase <= currentPrice {
			entryBase = currentPrice + buf
		}

		if isAggressive {
			// AGGRESSIVE MODE: 50% at current price + 50% at pullback near resistance
			entries = make([]dtos.TradingPlanEntry, 0, 2)

			entry1Price := currentPrice
			entry1Value := tradingCapital * 0.5
			entry1Qty := (entry1Value * float64(leverage)) / entry1Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entry1Price,
				PositionSize:  "50%",
				PositionValue: entry1Value,
				PositionQty:   entry1Qty,
			})

			entry2Price := entryBase
			entry2Value := tradingCapital * 0.5
			entry2Qty := (entry2Value * float64(leverage)) / entry2Price

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   2,
				EntryPrice:    entry2Price,
				PositionSize:  "50%",
				PositionValue: entry2Value,
				PositionQty:   entry2Qty,
			})

		} else {
			// CONSERVATIVE MODE: Single entry near resistance
			entries = make([]dtos.TradingPlanEntry, 0, 1)

			entryPrice := entryBase
			entryValue := tradingCapital
			entryQty := (entryValue * float64(leverage)) / entryPrice

			entries = append(entries, dtos.TradingPlanEntry{
				EntryNumber:   1,
				EntryPrice:    entryPrice,
				PositionSize:  "100%",
				PositionValue: entryValue,
				PositionQty:   entryQty,
			})
		}
	default:
		// Unknown action - return WAIT plan
		return &dtos.TradingPlan{
			Mode:          "WAIT",
			Entries:       make([]dtos.TradingPlanEntry, 0),
			BufferPercent: helpers.RoundFloat(bufferPercent*100, 3),
			Summary: &dtos.TradingPlanSummary{
				TotalEntries:        0,
				TotalPositionValue:  0,
				TotalPositionQty:    0,
				AvgEntryPrice:       0,
				MaxRiskUSDT:         0,
				MaxRiskPercent:      0,
				TargetProfitUSDT:    0,
				TargetProfitPercent: 0,
			},
		}
	}

	// Calculate summary data (pre-calculated for easy access)
	var totalValue float64
	var totalQty float64

	for _, entry := range entries {
		totalValue += entry.PositionValue
		totalQty += entry.PositionQty
	}

	var avgEntryPrice float64
	if totalQty > 0 {
		avgEntryPrice = (totalValue * float64(leverage)) / totalQty
	}

	// Calculate risk and profit
	var maxRiskUSDT, targetProfitUSDT float64
	switch signal {
	case "BUY", "STRONG_BUY", "WAIT":
		maxRiskUSDT = (avgEntryPrice - sl) * totalQty
		targetProfitUSDT = (tp - avgEntryPrice) * totalQty
	case "SELL", "STRONG_SELL":
		maxRiskUSDT = (sl - avgEntryPrice) * totalQty
		targetProfitUSDT = (avgEntryPrice - tp) * totalQty
	}

	// Calculate percentages from position value
	var maxRiskPercent, targetProfitPercent float64
	if totalValue > 0 {
		maxRiskPercent = (maxRiskUSDT / totalValue) * 100
		targetProfitPercent = (targetProfitUSDT / totalValue) * 100
	}

	// Calculate percentages from trading capital (more meaningful for risk management)
	var riskFromCapital, profitFromCapital, effectiveLeverage float64
	if tradingCapital > 0 {
		riskFromCapital = (maxRiskUSDT / tradingCapital) * 100
		profitFromCapital = (targetProfitUSDT / tradingCapital) * 100
		effectiveLeverage = totalValue / tradingCapital
	}

	// Calculate Risk/Reward ratio
	var calcRiskRewardRatio float64
	if len(entries) > 0 && maxRiskUSDT > 0 {
		calcRiskRewardRatio = helpers.RoundFloat(targetProfitUSDT/maxRiskUSDT, 2)
	}

	return &dtos.TradingPlan{
		Mode:            getMode(isAggressive),
		Entries:         entries,
		TakeProfit:      tp,
		StopLoss:        sl,
		Resistance:      resistance,
		Support:         support,
		RiskRewardRatio: calcRiskRewardRatio,
		BufferPercent:   bufferPercent * 100,
		Summary: &dtos.TradingPlanSummary{
			TotalEntries:        len(entries),
			TotalPositionValue:  helpers.RoundFloat(totalValue, 2),
			TotalPositionQty:    helpers.RoundFloat(totalQty, 8),
			AvgEntryPrice:       helpers.RoundFloat(avgEntryPrice, 8),
			MaxRiskUSDT:         helpers.RoundFloat(maxRiskUSDT, 2),
			MaxRiskPercent:      helpers.RoundFloat(maxRiskPercent, 2),
			RiskFromCapital:     helpers.RoundFloat(riskFromCapital, 2),
			TargetProfitUSDT:    helpers.RoundFloat(targetProfitUSDT, 2),
			TargetProfitPercent: helpers.RoundFloat(targetProfitPercent, 2),
			ProfitFromCapital:   helpers.RoundFloat(profitFromCapital, 2),
			EffectiveLeverage:   helpers.RoundFloat(effectiveLeverage, 2),
		},
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
