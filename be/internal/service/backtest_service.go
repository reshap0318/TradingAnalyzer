package service

import (
	"fmt"
	"math"
	"sort"
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
	backtestBufferPercent   = 20  // 20% extra candles for indicator warmup
	backtestMinCandles      = 150 // Minimum candles needed for indicators

	// Exit reasons
	exitReasonHitTP         = "HIT_TP"
	exitReasonHitSL         = "HIT_SL"
	exitReasonClosedEnd     = "CLOSED_END"
	exitReasonSignalReverse = "SIGNAL_REVERSE"
	exitReasonExpired       = "EXPIRED"
)

// ============================================================================
// Internal types for simulation
// ============================================================================

// backtestOrder represents a pending or filled order during simulation
type backtestOrder struct {
	TradeNum   int
	Side       string
	EntryPrice float64    // Target price from TradingPlan
	Quantity   float64
	TakeProfit float64
	StopLoss   float64
	EntryTime  time.Time  // When the order was created (OPEN)
	FilledTime *time.Time // When the order was filled (nil = still pending)
	Capital    float64
	IsFilled   bool    // false = pending, true = filled
	Signal     string  // Original signal (STRONG_BUY, BUY, etc.)
	Confidence float64
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

	trades, err := s.repo.BacktestTrade.FindByBacktestID(nil, bt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load trades: %w", err)
	}

	strategy, _ := s.StrategyGetByID(ctx, bt.StrategyID)

	tradeDTOs := make([]dtos.BacktestTradeDTO, len(trades))
	for i, t := range trades {
		tradeDTOs[i] = dtos.BacktestTradeDTO{
			ID:              t.ID,
			EntryTime:       t.EntryTime,
			FilledTime:      t.FilledTime,
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
	res, err = s.BacktestGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		if err := s.repo.BacktestTrade.DeleteByBacktestID(tx, id); err != nil {
			return nil, fmt.Errorf("failed to delete trades: %w", err)
		}
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
	fmt.Printf("⚙️  [BACKTEST] Leverage: %dx | Mode: %s | Expiration: %dh\n", mmConfig.LEVERAGE, backtestGetMode(mmConfig.IS_AGRESSIVE), mmConfig.ORDER_EXPIRATION_HOURS)

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
	fmt.Printf("   ├── Expired:       %d\n", stats.expiredTrades)
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
	timeframeMap := make(map[string]int)
	for _, tf := range strategy.Timeframes {
		tfModel, err := s.repo.Timeframe.FindByField(nil, &models.Timeframe{Name: tf.TimeframeName})
		if err != nil || len(tfModel) == 0 {
			return nil, fmt.Errorf("timeframe %s not found in database", tf.TimeframeName)
		}
		timeframeMap[tf.TimeframeName] = tfModel[0].InMinutes
	}

	results := make(map[string][]binance.KlineInfo)

	for _, tf := range strategy.Timeframes {
		tfMinutes := timeframeMap[tf.TimeframeName]
		candlesPerDay := (24 * 60) / tfMinutes
		totalNeeded := (days * candlesPerDay) + (days * candlesPerDay * backtestBufferPercent / 100)

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
		return s.BinanceClient.GetKlinesWithStartTime(symbol, interval, totalNeeded, startTime.UnixMilli())
	}

	var allKlines []binance.KlineInfo
	currentStartTime := startTime.UnixMilli()
	remaining := totalNeeded

	for remaining > 0 {
		limit := remaining
		if limit > backtestMaxBinanceLimit {
			limit = backtestMaxBinanceLimit
		}

		klines, err := s.BinanceClient.GetKlinesWithStartTime(symbol, interval, limit, currentStartTime)
		if err != nil {
			return nil, err
		}

		if len(klines) == 0 {
			break
		}

		allKlines = append(allKlines, klines...)
		remaining -= len(klines)

		lastCandle := klines[len(klines)-1]
		currentStartTime = lastCandle.CloseTime + 1
	}

	sort.Slice(allKlines, func(i, j int) bool {
		return allKlines[i].OpenTime < allKlines[j].OpenTime
	})

	return allKlines, nil
}

// ============================================================================
// Helper: Simulation Engine
// ============================================================================

// backtestRunSimulation runs the candle-by-candle simulation with pending order lifecycle
// Flow: Signal → OPEN (pending) → FILLED (price hit entry) → HIT_TP/HIT_SL/CLOSED_END
//
//	or → EXPIRED (not filled within ORDER_EXPIRATION_HOURS)
//
// Aggressive mode: entry 1 at current price (instant fill), entry 2 at support (pending)
func (s *Services) backtestRunSimulation(
	capital float64,
	strategy *dtos.StrategyData,
	allKlines map[string][]binance.KlineInfo,
	primaryKlines []binance.KlineInfo,
	thresholds []models.Threshold,
	mmConfig *config.MMConfig,
) []models.BacktestTrade {
	var completedTrades []models.BacktestTrade
	var pendingOrders []*backtestOrder // Orders waiting to be filled
	var filledOrders []*backtestOrder  // Active filled positions
	tradeCounter := 0
	tradCapital := capital
	expirationHours := int(mmConfig.ORDER_EXPIRATION_HOURS)
	if expirationHours <= 0 {
		expirationHours = 4 // Default: 4 hours
	}

	// Need minimum candles to calculate indicators
	startIdx := backtestMinCandles
	if startIdx >= len(primaryKlines) {
		startIdx = len(primaryKlines) / 2
	}

	for i := startIdx; i < len(primaryKlines); i++ {
		candle := primaryKlines[i]
		candleTime := time.UnixMilli(candle.OpenTime)

		// ──────────────────────────────────────────────────────────
		// STEP 1: Check EXPIRED on pending orders
		// ──────────────────────────────────────────────────────────
		var remainingPending []*backtestOrder
		for _, order := range pendingOrders {
			elapsed := candleTime.Sub(order.EntryTime)
			if elapsed.Hours() >= float64(expirationHours) {
				trade := backtestCreateExpiredTrade(order, candleTime)
				backtestLogExpired(order.TradeNum, order, candleTime, expirationHours)
				completedTrades = append(completedTrades, trade)
			} else {
				remainingPending = append(remainingPending, order)
			}
		}
		pendingOrders = remainingPending

		// ──────────────────────────────────────────────────────────
		// STEP 2: Check FILL on pending orders
		// ──────────────────────────────────────────────────────────
		var stillPending []*backtestOrder
		for _, order := range pendingOrders {
			filled := false
			if order.Side == "BUY" {
				// BUY: filled when candle low dips to entry price (reaches support)
				filled = candle.Low <= order.EntryPrice
			} else {
				// SELL: filled when candle high rises to entry price (reaches resistance)
				filled = candle.High >= order.EntryPrice
			}

			if filled {
				filledTime := candleTime
				order.FilledTime = &filledTime
				order.IsFilled = true
				filledOrders = append(filledOrders, order)
				backtestLogFilled(order.TradeNum, order, candleTime)
			} else {
				stillPending = append(stillPending, order)
			}
		}
		pendingOrders = stillPending

		// ──────────────────────────────────────────────────────────
		// STEP 3: Check TP/SL on FILLED positions
		// ──────────────────────────────────────────────────────────
		var activePositions []*backtestOrder
		for _, order := range filledOrders {
			exitReason := backtestCheckTPSL(order, candle)
			if exitReason != "" {
				trade := backtestCloseTrade(order, candle, exitReason, candleTime)
				backtestLogTradeResult(order.TradeNum, &trade, exitReason)
				completedTrades = append(completedTrades, trade)
				tradCapital += trade.PnL
			} else {
				activePositions = append(activePositions, order)
			}
		}
		filledOrders = activePositions

		// ──────────────────────────────────────────────────────────
		// STEP 4: Run signal analysis → create new pending orders
		// ──────────────────────────────────────────────────────────
		// Skip if we already have pending or filled orders
		if len(pendingOrders) > 0 || len(filledOrders) > 0 {
			continue
		}

		// Build subset klines up to this candle
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

		analyzeResult, err := s.signalAnalyzeCalculate(
			"",
			tradCapital,
			strategy,
			subsetKlines,
			thresholds,
		)
		if err != nil || analyzeResult == nil {
			continue
		}

		signal := analyzeResult.Signal.Signal
		if !analyzeResult.Signal.Valid {
			continue
		}

		action := ""
		if signal == "STRONG_BUY" || signal == "BUY" {
			action = "BUY"
		} else if signal == "STRONG_SELL" || signal == "SELL" {
			action = "SELL"
		}
		if action == "" {
			continue
		}

		// Create new orders from trading plan
		if analyzeResult.Signal.TradingPlan != nil {
			plan := analyzeResult.Signal.TradingPlan
			if plan.Mode == "WAIT" || len(plan.Entries) == 0 {
				continue
			}

			for _, entry := range plan.Entries {
				tradeCounter++
				entryPrice := entry.EntryPrice
				isFilled := false
				var filledTimePtr *time.Time

				// Aggressive mode: entry 1 is at current price (instant fill at candle close)
				if mmConfig.IS_AGRESSIVE && entry.EntryNumber == 1 {
					entryPrice = candle.Close
					isFilled = true
					ft := candleTime
					filledTimePtr = &ft
				}

				order := &backtestOrder{
					TradeNum:   tradeCounter,
					Side:       action,
					EntryPrice: entryPrice,
					Quantity:   entry.PositionQty,
					TakeProfit: plan.TakeProfit,
					StopLoss:   plan.StopLoss,
					EntryTime:  candleTime,
					FilledTime: filledTimePtr,
					Capital:    entry.PositionValue,
					IsFilled:   isFilled,
					Signal:     signal,
					Confidence: analyzeResult.Scoring.Confidence,
				}

				// Log entry creation
				backtestLogEntry(order)

				if isFilled {
					// Aggressive entry 1: directly filled
					backtestLogFilled(order.TradeNum, order, candleTime)
					filledOrders = append(filledOrders, order)
				} else {
					// Pending order waiting for price to reach entry
					pendingOrders = append(pendingOrders, order)
				}
			}
		}
	}

	// ──────────────────────────────────────────────────────────
	// END OF BACKTEST: Close/expire remaining orders
	// ──────────────────────────────────────────────────────────
	lastCandle := primaryKlines[len(primaryKlines)-1]
	lastTime := time.UnixMilli(lastCandle.CloseTime)

	// Close filled positions at last price
	for _, order := range filledOrders {
		trade := backtestCloseTrade(order, lastCandle, exitReasonClosedEnd, lastTime)
		backtestLogTradeResult(order.TradeNum, &trade, exitReasonClosedEnd)
		completedTrades = append(completedTrades, trade)
	}

	// Expire remaining pending orders
	for _, order := range pendingOrders {
		trade := backtestCreateExpiredTrade(order, lastTime)
		backtestLogExpired(order.TradeNum, order, lastTime, expirationHours)
		completedTrades = append(completedTrades, trade)
	}

	return completedTrades
}

// ============================================================================
// Helper: Trade Operations
// ============================================================================

// backtestCheckTPSL checks if TP or SL is hit on a candle (only for filled positions)
func backtestCheckTPSL(order *backtestOrder, candle binance.KlineInfo) string {
	if !order.IsFilled {
		return ""
	}

	if order.Side == "BUY" {
		if order.StopLoss > 0 && candle.Low <= order.StopLoss {
			return exitReasonHitSL
		}
		if order.TakeProfit > 0 && candle.High >= order.TakeProfit {
			return exitReasonHitTP
		}
	} else {
		if order.StopLoss > 0 && candle.High >= order.StopLoss {
			return exitReasonHitSL
		}
		if order.TakeProfit > 0 && candle.Low <= order.TakeProfit {
			return exitReasonHitTP
		}
	}
	return ""
}

// backtestCloseTrade creates a closed trade from a filled order
func backtestCloseTrade(order *backtestOrder, candle binance.KlineInfo, reason string, exitTime time.Time) models.BacktestTrade {
	var exitPrice float64
	switch reason {
	case exitReasonHitTP:
		exitPrice = order.TakeProfit
	case exitReasonHitSL:
		exitPrice = order.StopLoss
	default:
		exitPrice = candle.Close
	}

	var pnl float64
	if order.Side == "BUY" {
		pnl = (exitPrice - order.EntryPrice) * order.Quantity
	} else {
		pnl = (order.EntryPrice - exitPrice) * order.Quantity
	}

	pnlPercent := 0.0
	if order.Capital > 0 {
		pnlPercent = (pnl / order.Capital) * 100
	}

	// Duration from FILLED time (not entry/order creation time)
	var durationMinutes int64
	if order.FilledTime != nil {
		durationMinutes = int64(exitTime.Sub(*order.FilledTime).Minutes())
	}

	return models.BacktestTrade{
		EntryTime:       order.EntryTime,
		FilledTime:      order.FilledTime,
		ExitTime:        &exitTime,
		Side:            order.Side,
		EntryPrice:      order.EntryPrice,
		ExitPrice:       exitPrice,
		Quantity:        order.Quantity,
		PnL:             helpers.RoundFloat(pnl, 2),
		PnLPercent:      helpers.RoundFloat(pnlPercent, 2),
		TakeProfit:      order.TakeProfit,
		StopLoss:        order.StopLoss,
		ExitReason:      reason,
		Status:          "CLOSED",
		DurationMinutes: durationMinutes,
	}
}

// backtestCreateExpiredTrade creates an expired trade record (PnL = 0)
func backtestCreateExpiredTrade(order *backtestOrder, expiredTime time.Time) models.BacktestTrade {
	durationMinutes := int64(expiredTime.Sub(order.EntryTime).Minutes())

	return models.BacktestTrade{
		EntryTime:       order.EntryTime,
		ExitTime:        &expiredTime,
		Side:            order.Side,
		EntryPrice:      order.EntryPrice,
		ExitPrice:       0,
		Quantity:        order.Quantity,
		PnL:             0,
		PnLPercent:      0,
		TakeProfit:      order.TakeProfit,
		StopLoss:        order.StopLoss,
		ExitReason:      exitReasonExpired,
		Status:          "EXPIRED",
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
	expiredTrades   int
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
	filledTrades := 0

	for _, trade := range trades {
		if trade.ExitReason == exitReasonExpired {
			stats.expiredTrades++
			continue
		}

		filledTrades++
		stats.totalPnL += trade.PnL

		if trade.PnL > 0 {
			stats.winningTrades++
			totalProfit += trade.PnL
		} else if trade.PnL < 0 {
			stats.losingTrades++
			totalLoss += math.Abs(trade.PnL)
		}

		runningPnL += trade.PnL
		if runningPnL > peakPnL {
			peakPnL = runningPnL
		}
		drawdown := peakPnL - runningPnL
		if drawdown > stats.maxDrawdown {
			stats.maxDrawdown = drawdown
		}
	}

	stats.totalPnL = helpers.RoundFloat(stats.totalPnL, 2)
	if initialCapital > 0 {
		stats.totalPnLPercent = helpers.RoundFloat((stats.totalPnL/initialCapital)*100, 2)
	}

	if filledTrades > 0 {
		stats.winRate = helpers.RoundFloat(float64(stats.winningTrades)/float64(filledTrades)*100, 2)
	}

	if totalLoss > 0 {
		stats.profitFactor = helpers.RoundFloat(totalProfit/totalLoss, 2)
	} else if totalProfit > 0 {
		stats.profitFactor = 999.99
	}

	stats.maxDrawdown = helpers.RoundFloat(stats.maxDrawdown, 2)

	return stats
}

// ============================================================================
// Helper: Console Logging
// ============================================================================

// backtestLogEntry logs a new order creation (OPEN)
func backtestLogEntry(order *backtestOrder) {
	mode := "PENDING"
	if order.IsFilled {
		mode = "INSTANT"
	}
	fmt.Printf("📌 [TRADE #%d] ENTRY %s (%s) @ %s\n", order.TradeNum, order.Side, mode, order.EntryTime.Format("2006-01-02 15:04"))
	fmt.Printf("   ├── Entry Price: $%.4f | Qty: %.4f | Value: $%.2f\n", order.EntryPrice, order.Quantity, order.Capital)
	fmt.Printf("   ├── TP: $%.4f | SL: $%.4f\n", order.TakeProfit, order.StopLoss)
	fmt.Printf("   └── Confidence: %.1f | Signal: %s\n", order.Confidence, order.Signal)
	fmt.Println()
}

// backtestLogFilled logs when a pending order is filled
func backtestLogFilled(tradeNum int, order *backtestOrder, filledTime time.Time) {
	waited := filledTime.Sub(order.EntryTime)
	fmt.Printf("✅ [TRADE #%d] FILLED @ %s\n", tradeNum, filledTime.Format("2006-01-02 15:04"))
	fmt.Printf("   ├── Fill Price: $%.4f\n", order.EntryPrice)
	fmt.Printf("   └── Waited: %s\n", backtestFormatDuration(int64(waited.Minutes())))
	fmt.Println()
}

// backtestLogExpired logs when a pending order expires
func backtestLogExpired(tradeNum int, order *backtestOrder, expiredTime time.Time, expirationHours int) {
	fmt.Printf("⏰ [TRADE #%d] EXPIRED @ %s\n", tradeNum, expiredTime.Format("2006-01-02 15:04"))
	fmt.Printf("   └── Not filled after %dh (ORDER_EXPIRATION_HOURS=%d)\n", expirationHours, expirationHours)
	fmt.Println()
}

// backtestLogTradeResult logs a trade result (HIT_TP, HIT_SL, etc.)
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
	fmt.Printf("   └── Duration: %s (from fill)\n", backtestFormatDuration(trade.DurationMinutes))
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
