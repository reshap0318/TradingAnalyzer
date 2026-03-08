package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// Constants
// ============================================================================

const (
	backtestMaxBinanceLimit = 1500
	backtestBufferPercent   = 20 // 20% extra candles for indicator warmup
	backtestMinCandles      = 150 // Minimum candles needed for indicators

	// Exit reasons
	exitReasonHitTP         = "HIT_TP"
	exitReasonHitSL         = "HIT_SL"
	exitReasonClosedEnd     = "CLOSED_END"
	exitReasonSignalReverse = "SIGNAL_REVERSE"
)

// ============================================================================
// Internal types for simulation
// ============================================================================

// backtestPosition represents an open position during simulation
type backtestPosition struct {
	Side       string
	EntryPrice float64
	Quantity   float64
	TakeProfit float64
	StopLoss   float64
	EntryTime  time.Time
	Capital    float64
}

// ============================================================================
// CRUD Operations
// ============================================================================

// BacktestGetAll lists all backtests (without trades)
func (s *Services) BacktestGetAll(ctx *gin.Context) (res []dtos.BacktestListItem, err error) {
	backtests, err := s.repo.Backtest.FindAllOrderByCreatedAtDESC(nil)
	if err != nil {
		return nil, err
	}

	res = make([]dtos.BacktestListItem, len(backtests))
	for i, bt := range backtests {
		// Load strategy name
		strategyName := ""
		strategy, stratErr := s.StrategyGetByID(ctx, bt.StrategyID)
		if stratErr == nil && strategy != nil {
			strategyName = strategy.StrategyName
		}

		res[i] = dtos.BacktestListItem{
			ID:              bt.ID,
			Name:            bt.Name,
			Symbol:          bt.Symbol,
			StrategyName:    strategyName,
			TotalPnL:        bt.TotalPnL,
			TotalPnLPercent: bt.TotalPnLPercent,
			WinRate:         bt.WinRate,
			TotalTrades:     bt.TotalTrades,
			Status:          bt.Status,
			CreatedAt:       bt.CreatedAt,
		}
	}

	return
}

// BacktestGetByID gets backtest by ID with trades and strategy detail
func (s *Services) BacktestGetByID(ctx *gin.Context, id uint) (res *dtos.BacktestResponse, err error) {
	bt, err := s.repo.Backtest.FindByID(nil, id)
	if err != nil {
		return nil, fmt.Errorf("backtest not found: %w", err)
	}

	// Load trades
	trades, err := s.repo.BacktestTrade.FindByBacktestID(nil, bt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load trades: %w", err)
	}

	// Load strategy detail
	strategy, _ := s.StrategyGetByID(ctx, bt.StrategyID)

	// Convert trades to DTO
	tradeDTOs := make([]dtos.BacktestTradeDTO, len(trades))
	for i, t := range trades {
		tradeDTOs[i] = dtos.BacktestTradeDTO{
			ID:              t.ID,
			EntryTime:       t.EntryTime,
			ExitTime:        t.ExitTime,
			Side:            t.Side,
			EntryPrice:      t.EntryPrice,
			ExitPrice:       t.ExitPrice,
			Quantity:        t.Quantity,
			PnL:             t.PnL,
			PnLPercent:      t.PnLPercent,
			TakeProfit:      t.TakeProfit,
			StopLoss:        t.StopLoss,
			ExitReason:      t.ExitReason,
			Status:          t.Status,
			DurationMinutes: t.DurationMinutes,
		}
	}

	return &dtos.BacktestResponse{
		ID:              bt.ID,
		Name:            bt.Name,
		Symbol:          bt.Symbol,
		StrategyID:      bt.StrategyID,
		StartTime:       bt.StartTime,
		EndTime:         bt.EndTime,
		Capital:         bt.Capital,
		TotalTrades:     bt.TotalTrades,
		WinningTrades:   bt.WinningTrades,
		LosingTrades:    bt.LosingTrades,
		TotalPnL:        bt.TotalPnL,
		TotalPnLPercent: bt.TotalPnLPercent,
		MaxDrawdown:     bt.MaxDrawdown,
		WinRate:         bt.WinRate,
		ProfitFactor:    bt.ProfitFactor,
		Status:          bt.Status,
		ErrorMessage:    bt.ErrorMessage,
		CreatedAt:       bt.CreatedAt,
		CompletedAt:     bt.CompletedAt,
		Strategy:        strategy,
		Trades:          tradeDTOs,
	}, nil
}

// BacktestDelete deletes a backtest and its trades
func (s *Services) BacktestDelete(ctx *gin.Context, id uint) (res *dtos.BacktestResponse, err error) {
	// Get backtest first
	res, err = s.BacktestGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// Delete trades first (child records)
		if err := s.repo.BacktestTrade.DeleteByBacktestID(tx, id); err != nil {
			return nil, fmt.Errorf("failed to delete trades: %w", err)
		}

		// Delete backtest
		if _, err := s.repo.Backtest.Delete(tx, id); err != nil {
			return nil, fmt.Errorf("failed to delete backtest: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ============================================================================
// Backtest Execution
// ============================================================================

// BacktestCreate creates and runs a new backtest
func (s *Services) BacktestCreate(ctx *gin.Context, req *dtos.BacktestRequest) (res *dtos.BacktestResponse, err error) {
	startExec := time.Now()

	fmt.Println()
	fmt.Println("🚀 [BACKTEST] ═══════════════════════════════════════════")
	fmt.Printf("🚀 [BACKTEST] Starting backtest \"%s\" for %s\n", req.Name, req.Symbol)

	// 1. Load strategy
	strategy, err := s.StrategyGetByID(ctx, req.StrategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load strategy: %w", err)
	}

	// 2. Get money management config (validated with fallback to defaults)
	mmConfig := s.getConfigMM(strategy)

	fmt.Printf("📊 [BACKTEST] Strategy: \"%s\" (ID: %d) | Capital: $%.2f\n", strategy.StrategyName, strategy.ID, req.Capital)
	fmt.Printf("⚙️  [BACKTEST] Leverage: %dx | Mode: %s\n", mmConfig.LEVERAGE, backtestGetMode(mmConfig.IS_AGRESSIVE))

	// 3. Calculate time range
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -req.Days)

	fmt.Printf("⏱️  [BACKTEST] Period: %d days | Start: %s → End: %s\n",
		req.Days,
		startTime.Format("2006-01-02"),
		endTime.Format("2006-01-02"),
	)

	// 4. Fetch thresholds
	thresholds, err := s.repo.Threshold.FindAll(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thresholds: %w", err)
	}

	// 5. Calculate dynamic limits and fetch klines
	fmt.Println()
	fmt.Println("📥 [BACKTEST] Fetching klines...")

	binanceData, err := s.backtestFetchAllKlines(req.Symbol, req.Days, startTime, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}

	// 6. Validate primary timeframe data
	primaryKlines, exists := binanceData[strategy.PrimaryTF]
	if !exists || len(primaryKlines) == 0 {
		return nil, fmt.Errorf("no kline data for primary timeframe %s", strategy.PrimaryTF)
	}

	// Filter klines to only include data within the backtest period
	startTimeMs := startTime.UnixMilli()
	for tf, klines := range binanceData {
		filtered := make([]binance.KlineInfo, 0)
		for _, k := range klines {
			if k.OpenTime >= startTimeMs {
				filtered = append(filtered, k)
			}
		}
		binanceData[tf] = filtered
	}
	primaryKlines = binanceData[strategy.PrimaryTF]

	fmt.Printf("\n🔄 [BACKTEST] Running simulation... (%d primary candles to process)\n\n", len(primaryKlines))

	// 7. Run simulation
	completedTrades := s.backtestRunSimulation(
		req.Capital,
		strategy,
		binanceData,
		primaryKlines,
		thresholds,
		mmConfig,
	)

	// 8. Calculate statistics
	stats := backtestCalculateStats(completedTrades, req.Capital)

	// 9. Log summary
	fmt.Println()
	fmt.Println("📊 [BACKTEST] ═══════════════════════════════════════════")
	fmt.Printf("   ├── Total Trades:  %d\n", stats.totalTrades)
	fmt.Printf("   ├── Win/Loss:      %dW / %dL (%.1f%%)\n", stats.winningTrades, stats.losingTrades, stats.winRate)
	fmt.Printf("   ├── Total PnL:     %+.2f (%.2f%%)\n", stats.totalPnL, stats.totalPnLPercent)
	fmt.Printf("   ├── Profit Factor: %.2f\n", stats.profitFactor)
	fmt.Printf("   └── Max Drawdown:  %.2f\n", stats.maxDrawdown)

	// 10. Save to database
	now := time.Now()
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		backtest := &models.Backtest{
			Name:            req.Name,
			Symbol:          req.Symbol,
			StrategyID:      req.StrategyID,
			StartTime:       startTime,
			EndTime:         endTime,
			Capital:         req.Capital,
			TotalTrades:     stats.totalTrades,
			WinningTrades:   stats.winningTrades,
			LosingTrades:    stats.losingTrades,
			TotalPnL:        stats.totalPnL,
			TotalPnLPercent: stats.totalPnLPercent,
			MaxDrawdown:     stats.maxDrawdown,
			WinRate:         stats.winRate,
			ProfitFactor:    stats.profitFactor,
			Status:          "COMPLETED",
			CompletedAt:     &now,
		}

		backtest, err = s.repo.Backtest.Create(tx, backtest)
		if err != nil {
			return nil, fmt.Errorf("failed to save backtest: %w", err)
		}

		// Save trades
		for _, trade := range completedTrades {
			trade.BacktestID = backtest.ID
			_, err = s.repo.BacktestTrade.Create(tx, &trade)
			if err != nil {
				return nil, fmt.Errorf("failed to save trade: %w", err)
			}
		}

		return backtest, nil
	})

	if err != nil {
		return nil, err
	}

	bt := result.(*models.Backtest)

	elapsed := time.Since(startExec)
	fmt.Printf("💾 [BACKTEST] Saved to database (ID: %d)\n", bt.ID)
	fmt.Printf("✅ [BACKTEST] Completed in %.1fs\n", elapsed.Seconds())
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	return s.BacktestGetByID(ctx, bt.ID)
}

// ============================================================================
// Helper: Fetch Klines with Batching
// ============================================================================

// backtestFetchAllKlines fetches klines for all strategy timeframes with dynamic limits
func (s *Services) backtestFetchAllKlines(
	symbol string,
	days int,
	startTime time.Time,
	strategy *dtos.StrategyData,
) (map[string][]binance.KlineInfo, error) {
	// Get timeframe in_minutes from strategy timeframes via database
	timeframeMap := make(map[string]int) // tf_name -> in_minutes
	for _, tf := range strategy.Timeframes {
		tfModel, err := s.repo.Timeframe.FindByField(nil, &models.Timeframe{Name: tf.TimeframeName})
		if err != nil || len(tfModel) == 0 {
			return nil, fmt.Errorf("timeframe %s not found in database", tf.TimeframeName)
		}
		timeframeMap[tf.TimeframeName] = tfModel[0].InMinutes
	}

	// Fetch klines for each timeframe
	results := make(map[string][]binance.KlineInfo)

	for _, tf := range strategy.Timeframes {
		tfMinutes := timeframeMap[tf.TimeframeName]
		candlesPerDay := (24 * 60) / tfMinutes
		totalNeeded := (days * candlesPerDay) + (days * candlesPerDay * backtestBufferPercent / 100)

		// Ensure minimum candles for indicator calculation
		if totalNeeded < backtestMinCandles {
			totalNeeded = backtestMinCandles
		}

		klines, err := s.backtestFetchKlinesBatched(symbol, tf.TimeframeName, totalNeeded, startTime)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch klines for %s: %w", tf.TimeframeName, err)
		}

		statusIcon := "✅"
		batchInfo := ""
		if totalNeeded > backtestMaxBinanceLimit {
			batches := int(math.Ceil(float64(totalNeeded) / float64(backtestMaxBinanceLimit)))
			batchInfo = fmt.Sprintf(" (%d batches)", batches)
			statusIcon = "🔄"
		}

		fmt.Printf("   ├── %s: need=%d, got=%d%s %s\n",
			tf.TimeframeName, totalNeeded, len(klines), batchInfo, statusIcon)

		results[tf.TimeframeName] = klines
	}

	return results, nil
}

// backtestFetchKlinesBatched fetches klines in batches of max 1500
func (s *Services) backtestFetchKlinesBatched(
	symbol string,
	interval string,
	totalNeeded int,
	startTime time.Time,
) ([]binance.KlineInfo, error) {
	if totalNeeded <= backtestMaxBinanceLimit {
		// Single request is enough
		return s.backtestFetchKlinesWithStartTime(symbol, interval, totalNeeded, startTime.UnixMilli())
	}

	// Need multiple batches
	var allKlines []binance.KlineInfo
	currentStartTime := startTime.UnixMilli()
	remaining := totalNeeded

	for remaining > 0 {
		limit := remaining
		if limit > backtestMaxBinanceLimit {
			limit = backtestMaxBinanceLimit
		}

		klines, err := s.backtestFetchKlinesWithStartTime(symbol, interval, limit, currentStartTime)
		if err != nil {
			return nil, err
		}

		if len(klines) == 0 {
			break
		}

		allKlines = append(allKlines, klines...)
		remaining -= len(klines)

		// Move startTime to after the last received candle
		lastCandle := klines[len(klines)-1]
		currentStartTime = lastCandle.CloseTime + 1
	}

	// Sort by OpenTime to ensure correct order
	sort.Slice(allKlines, func(i, j int) bool {
		return allKlines[i].OpenTime < allKlines[j].OpenTime
	})

	return allKlines, nil
}

// backtestFetchKlinesWithStartTime fetches klines with startTime parameter
// This is a helper function, NOT modifying binance/service.go
func (s *Services) backtestFetchKlinesWithStartTime(
	symbol string,
	interval string,
	limit int,
	startTimeMs int64,
) ([]binance.KlineInfo, error) {
	ctx := context.Background()

	resp, err := s.BinanceClient.GetAPIClient().NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		StartTime(startTimeMs).
		Limit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get klines for %s %s: %w", symbol, interval, err)
	}

	klines := make([]binance.KlineInfo, len(resp))
	for i, k := range resp {
		klines[i] = binance.KlineInfo{
			OpenTime:  k.OpenTime,
			Open:      backtestParseFloat(k.Open),
			High:      backtestParseFloat(k.High),
			Low:       backtestParseFloat(k.Low),
			Close:     backtestParseFloat(k.Close),
			Volume:    backtestParseFloat(k.Volume),
			CloseTime: k.CloseTime,
		}
	}

	return klines, nil
}

// ============================================================================
// Helper: Simulation Engine
// ============================================================================

// backtestRunSimulation runs the candle-by-candle simulation
func (s *Services) backtestRunSimulation(
	capital float64,
	strategy *dtos.StrategyData,
	allKlines map[string][]binance.KlineInfo,
	primaryKlines []binance.KlineInfo,
	thresholds []models.Threshold,
	mmConfig *config.MMConfig,
) []models.BacktestTrade {
	var completedTrades []models.BacktestTrade
	var currentPosition *backtestPosition
	tradeCounter := 0
	tradCapital := capital

	// We need at least backtestMinCandles candles to calculate indicators
	startIdx := backtestMinCandles
	if startIdx >= len(primaryKlines) {
		startIdx = len(primaryKlines) / 2 // Use at least half the data for simulation
	}

	for i := startIdx; i < len(primaryKlines); i++ {
		candle := primaryKlines[i]
		candleTime := time.UnixMilli(candle.OpenTime)

		// Build subset klines up to this candle for each timeframe
		subsetKlines := make(map[string][]binance.KlineInfo)
		for tf, klines := range allKlines {
			subset := make([]binance.KlineInfo, 0)
			for _, k := range klines {
				if k.OpenTime <= candle.OpenTime {
					subset = append(subset, k)
				}
			}
			if len(subset) > 0 {
				subsetKlines[tf] = subset
			}
		}

		// Check TP/SL for current position FIRST
		if currentPosition != nil {
			exitReason := backtestCheckTPSL(currentPosition, candle)
			if exitReason != "" {
				tradeCounter++
				trade := backtestCloseTrade(currentPosition, candle, exitReason, candleTime)

				// Log trade result
				backtestLogTradeResult(tradeCounter, &trade, exitReason)

				completedTrades = append(completedTrades, trade)
				tradCapital += trade.PnL
				currentPosition = nil
			}
		}

		// Run signal analysis on subset
		analyzeResult, err := s.signalAnalyzeCalculate(
			"", // symbol not needed for calculation
			tradCapital,
			strategy,
			subsetKlines,
			thresholds,
		)
		if err != nil || analyzeResult == nil {
			continue
		}

		signal := analyzeResult.Signal.Signal
		isValid := analyzeResult.Signal.Valid

		if !isValid {
			continue
		}

		// Determine action
		action := ""
		if signal == "STRONG_BUY" || signal == "BUY" {
			action = "BUY"
		} else if signal == "STRONG_SELL" || signal == "SELL" {
			action = "SELL"
		}

		if action == "" {
			continue
		}

		// Check if we need to reverse position
		if currentPosition != nil {
			if (action == "BUY" && currentPosition.Side == "SELL") ||
				(action == "SELL" && currentPosition.Side == "BUY") {
				// Reverse: close current position
				tradeCounter++
				trade := backtestCloseTrade(currentPosition, candle, exitReasonSignalReverse, candleTime)

				backtestLogTradeResult(tradeCounter, &trade, exitReasonSignalReverse)

				completedTrades = append(completedTrades, trade)
				tradCapital += trade.PnL
				currentPosition = nil
			} else {
				// Same direction, skip
				continue
			}
		}

		// Open new position
		if currentPosition == nil && analyzeResult.Signal.TradingPlan != nil {
			plan := analyzeResult.Signal.TradingPlan
			if plan.Mode == "WAIT" || len(plan.Entries) == 0 {
				continue
			}

			// Use first entry for backtest simulation
			entry := plan.Entries[0]
			entryPrice := candle.Close // Use candle close as entry price for simulation

			tradeCounter++
			currentPosition = &backtestPosition{
				Side:       action,
				EntryPrice: entryPrice,
				Quantity:   entry.PositionQty,
				TakeProfit: plan.TakeProfit,
				StopLoss:   plan.StopLoss,
				EntryTime:  candleTime,
				Capital:    entry.PositionValue,
			}

			// Log entry
			backtestLogEntry(tradeCounter, currentPosition, signal, analyzeResult.Scoring.Confidence)
		}
	}

	// Close any remaining open position at the end
	if currentPosition != nil {
		lastCandle := primaryKlines[len(primaryKlines)-1]
		lastTime := time.UnixMilli(lastCandle.CloseTime)

		tradeCounter++
		trade := backtestCloseTrade(currentPosition, lastCandle, exitReasonClosedEnd, lastTime)

		backtestLogTradeResult(tradeCounter, &trade, exitReasonClosedEnd)

		completedTrades = append(completedTrades, trade)
	}

	return completedTrades
}

// ============================================================================
// Helper: Trade Operations
// ============================================================================

// backtestCheckTPSL checks if TP or SL is hit on a candle
func backtestCheckTPSL(pos *backtestPosition, candle binance.KlineInfo) string {
	if pos.Side == "BUY" {
		// Long position: TP hit if high >= TP, SL hit if low <= SL
		if pos.StopLoss > 0 && candle.Low <= pos.StopLoss {
			return exitReasonHitSL
		}
		if pos.TakeProfit > 0 && candle.High >= pos.TakeProfit {
			return exitReasonHitTP
		}
	} else {
		// Short position: TP hit if low <= TP, SL hit if high >= SL
		if pos.StopLoss > 0 && candle.High >= pos.StopLoss {
			return exitReasonHitSL
		}
		if pos.TakeProfit > 0 && candle.Low <= pos.TakeProfit {
			return exitReasonHitTP
		}
	}
	return ""
}

// backtestCloseTrade creates a closed trade from position
func backtestCloseTrade(pos *backtestPosition, candle binance.KlineInfo, reason string, exitTime time.Time) models.BacktestTrade {
	var exitPrice float64
	switch reason {
	case exitReasonHitTP:
		exitPrice = pos.TakeProfit
	case exitReasonHitSL:
		exitPrice = pos.StopLoss
	default:
		exitPrice = candle.Close
	}

	// Calculate PnL
	var pnl float64
	if pos.Side == "BUY" {
		pnl = (exitPrice - pos.EntryPrice) * pos.Quantity
	} else {
		pnl = (pos.EntryPrice - exitPrice) * pos.Quantity
	}

	pnlPercent := 0.0
	if pos.Capital > 0 {
		pnlPercent = (pnl / pos.Capital) * 100
	}

	duration := exitTime.Sub(pos.EntryTime)
	durationMinutes := int64(duration.Minutes())

	return models.BacktestTrade{
		EntryTime:       pos.EntryTime,
		ExitTime:        &exitTime,
		Side:            pos.Side,
		EntryPrice:      pos.EntryPrice,
		ExitPrice:       exitPrice,
		Quantity:        pos.Quantity,
		PnL:             helpers.RoundFloat(pnl, 2),
		PnLPercent:      helpers.RoundFloat(pnlPercent, 2),
		TakeProfit:      pos.TakeProfit,
		StopLoss:        pos.StopLoss,
		ExitReason:      reason,
		Status:          "CLOSED",
		DurationMinutes: durationMinutes,
	}
}

// ============================================================================
// Helper: Statistics
// ============================================================================

type backtestStats struct {
	totalTrades     int
	winningTrades   int
	losingTrades    int
	totalPnL        float64
	totalPnLPercent float64
	maxDrawdown     float64
	winRate         float64
	profitFactor    float64
}

// backtestCalculateStats calculates backtest performance statistics
func backtestCalculateStats(trades []models.BacktestTrade, initialCapital float64) backtestStats {
	stats := backtestStats{}
	stats.totalTrades = len(trades)

	if stats.totalTrades == 0 {
		return stats
	}

	var totalProfit, totalLoss float64
	var runningPnL float64
	var peakPnL float64

	for _, trade := range trades {
		stats.totalPnL += trade.PnL

		if trade.PnL > 0 {
			stats.winningTrades++
			totalProfit += trade.PnL
		} else if trade.PnL < 0 {
			stats.losingTrades++
			totalLoss += math.Abs(trade.PnL)
		}

		// Track drawdown
		runningPnL += trade.PnL
		if runningPnL > peakPnL {
			peakPnL = runningPnL
		}
		drawdown := peakPnL - runningPnL
		if drawdown > stats.maxDrawdown {
			stats.maxDrawdown = drawdown
		}
	}

	// Calculate percentages
	stats.totalPnL = helpers.RoundFloat(stats.totalPnL, 2)
	if initialCapital > 0 {
		stats.totalPnLPercent = helpers.RoundFloat((stats.totalPnL/initialCapital)*100, 2)
	}

	if stats.totalTrades > 0 {
		stats.winRate = helpers.RoundFloat(float64(stats.winningTrades)/float64(stats.totalTrades)*100, 2)
	}

	if totalLoss > 0 {
		stats.profitFactor = helpers.RoundFloat(totalProfit/totalLoss, 2)
	} else if totalProfit > 0 {
		stats.profitFactor = 999.99 // All wins, no losses
	}

	stats.maxDrawdown = helpers.RoundFloat(stats.maxDrawdown, 2)

	return stats
}

// ============================================================================
// Helper: Console Logging
// ============================================================================

// backtestLogEntry logs a new trade entry
func backtestLogEntry(tradeNum int, pos *backtestPosition, signal string, confidence float64) {
	fmt.Printf("📌 [TRADE #%d] ENTRY %s @ %s\n", tradeNum, pos.Side, pos.EntryTime.Format("2006-01-02 15:04"))
	fmt.Printf("   ├── Price: $%.4f | Qty: %.4f | Value: $%.2f\n", pos.EntryPrice, pos.Quantity, pos.Capital)
	fmt.Printf("   ├── TP: $%.4f | SL: $%.4f\n", pos.TakeProfit, pos.StopLoss)
	fmt.Printf("   └── Confidence: %.1f | Signal: %s\n", confidence, signal)
	fmt.Println()
}

// backtestLogTradeResult logs a trade result
func backtestLogTradeResult(tradeNum int, trade *models.BacktestTrade, reason string) {
	icon := "✅"
	if reason == exitReasonHitSL {
		icon = "❌"
	} else if reason == exitReasonClosedEnd {
		icon = "⏹️ "
	} else if reason == exitReasonSignalReverse {
		icon = "🔄"
	}

	fmt.Printf("%s [TRADE #%d] %s @ %s\n", icon, tradeNum, reason, trade.ExitTime.Format("2006-01-02 15:04"))
	fmt.Printf("   ├── Exit: $%.4f | PnL: %+.2f (%+.2f%%)\n", trade.ExitPrice, trade.PnL, trade.PnLPercent)
	fmt.Printf("   └── Duration: %s\n", backtestFormatDuration(trade.DurationMinutes))
	fmt.Println()
}

// backtestFormatDuration formats duration minutes to human readable
func backtestFormatDuration(minutes int64) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	days := hours / 24
	remainHours := hours % 24
	return fmt.Sprintf("%dd %dh", days, remainHours)
}

// backtestGetMode returns mode string
func backtestGetMode(isAggressive bool) string {
	if isAggressive {
		return "AGGRESSIVE"
	}
	return "CONSERVATIVE"
}

// backtestParseFloat parses float string (helper to avoid importing binance internals)
func backtestParseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

