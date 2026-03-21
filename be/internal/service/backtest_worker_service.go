package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

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
	exitReasonHitTP      = "HIT_TP"
	exitReasonHitSL      = "HIT_SL"
	exitReasonClosedEnd  = "CLOSED_END"
	exitReasonDeadSignal = "DEAD_SIGNAL"
	exitReasonExpired    = "EXPIRED"
)

// ============================================================================
// Internal types for simulation
// ============================================================================

// virtualWallet tracks the simulated balance
type virtualWallet struct {
	Balance        float64
	InitialBalance float64
}

// dailyStats tracks daily trading statistics
type dailyStats struct {
	Date              string  // YYYY-MM-DD
	Count             int     // Number of trades today
	PnL               float64 // Net PnL today
	SLHits            int     // Stop loss hits today
	TPHits            int     // Take profit hits today
	ConsecutiveLosses int     // Consecutive losses (NOT reset daily)
}

// tradeEntry represents a single entry in a multi-entry trade
type tradeEntry struct {
	EntryNum  int
	Type      string // MARKET or LIMIT
	Price     float64
	Qty       float64
	Timestamp time.Time
	Status    string // PENDING, FILLED, CANCELLED, EXPIRED
	CreatedAt time.Time
}

// activeTrade represents a currently active trade position
type activeTrade struct {
	TradeNum         int
	Side             string
	Signal           string
	Confidence       float64
	TradingMode      string // AGGRESSIVE or CONSERVATIVE
	Entries          []*tradeEntry
	TotalQty         float64
	AvgEntryPrice    float64
	TakeProfit       float64
	StopLoss         float64
	RiskRewardRatio  float64
	EntryTime        time.Time
	FilledTime       *time.Time
	TPPrice          float64
	SLPrice          float64
	DailyTradeCount  int
	DailyPnL         float64
	ConsecutiveLoss  int
	CapitalAllocated float64
}

// equityPoint represents a point in the equity curve
type equityPoint struct {
	Timestamp int64
	Balance   float64
	PnL       float64
}

// simulationResult holds the complete simulation output
type simulationResult struct {
	Trades      []models.BacktestTrade
	EquityCurve []equityPoint
	Summary     dtos.BacktestSummary
}

// ============================================================================
// Backtest Worker
// ============================================================================

// backtestRunWorker runs the backtest simulation in the background
func (s *Services) backtestRunWorker(
	backtestID uint,
	days int,
	strategy *dtos.StrategyData,
	mmConfig *config.MMConfig,
	symbol string,
	initialCapital float64,
	startTime time.Time,
	endTime time.Time,
) {
	fmt.Println()
	fmt.Println("👷 [BACKTEST WORKER] ═══════════════════════════════════")
	fmt.Printf("👷 [BACKTEST WORKER] Starting worker for backtest ID: %d\n", backtestID)
	fmt.Printf("👷 [BACKTEST WORKER] Symbol: %s, Capital: $%.2f, Days: %d\n", symbol, initialCapital, days)

	// 1. Update status to RUNNING
	now := time.Now()
	_, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		update := map[string]interface{}{
			"status":     "RUNNING",
			"updated_at": now,
		}
		return nil, s.repo.Backtest.Update(tx, update, backtestID)
	})

	if err != nil {
		fmt.Printf("❌ [BACKTEST WORKER] Failed to update status to RUNNING: %v\n", err)
		return
	}

	fmt.Printf("📊 [BACKTEST WORKER] Status updated to RUNNING\n")

	// 2. Fetch thresholds
	thresholds, err := s.repo.Threshold.FindAll(nil)
	if err != nil {
		s.backtestWorkerFailed(backtestID, fmt.Errorf("failed to fetch thresholds: %w", err))
		return
	}

	// 3. Fetch klines
	fmt.Println()
	fmt.Println("📥 [BACKTEST WORKER] Fetching klines...")

	binanceData, err := s.backtestFetchAllKlines(symbol, days, startTime, endTime, strategy)
	if err != nil {
		s.backtestWorkerFailed(backtestID, fmt.Errorf("failed to fetch klines: %w", err))
		return
	}

	// 4. Validate primary timeframe data
	primaryKlines, exists := binanceData[strategy.PrimaryTF]
	if !exists || len(primaryKlines) == 0 {
		s.backtestWorkerFailed(backtestID, fmt.Errorf("no kline data for primary timeframe %s", strategy.PrimaryTF))
		return
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

	fmt.Printf("\n🔄 [BACKTEST WORKER] Running simulation... (%d primary candles to process)\n\n", len(primaryKlines))

	// 5. Run simulation (per TRADE_BACKTEST_FLOW.md)
	simulationResult := s.backtestRunSimulation(
		symbol,
		initialCapital,
		strategy,
		binanceData,
		primaryKlines,
		thresholds,
		mmConfig,
	)

	// 6. Calculate and save statistics
	summary := simulationResult.Summary
	equityCurve := simulationResult.EquityCurve

	// 7. Log summary
	fmt.Println()
	fmt.Println("📊 [BACKTEST WORKER] ═══════════════════════════════════")
	fmt.Printf("   ├── Total Trades:  %d\n", summary.TotalTrades)
	fmt.Printf("   ├── Win/Loss:      %dW / %dL (%.1f%%)\n",
		summary.WinningTrades,
		summary.LosingTrades,
		summary.WinRate)
	fmt.Printf("   ├── Expired:       %d\n", summary.ExpiredTrades)
	fmt.Printf("   ├── Cancelled:     %d\n", summary.CancelledTrades)
	fmt.Printf("   ├── Total PnL:     %+.2f (%.2f%%)\n", summary.NetProfit, summary.NetProfitPercent)
	fmt.Printf("   ├── Profit Factor: %.2f\n", summary.ProfitFactor)
	fmt.Printf("   └── Max Drawdown:  %.2f%%\n", summary.MaxDrawdown)

	// 8. Save results to database
	completedAt := time.Now()

	// Convert equity curve to JSON
	equityCurveJSON, _ := json.Marshal(equityCurve)

	_, err = s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		update := map[string]interface{}{
			"status":            "COMPLETED",
			"total_trades":      summary.TotalTrades,
			"winning_trades":    summary.WinningTrades,
			"losing_trades":     summary.LosingTrades,
			"expired_trades":    summary.ExpiredTrades,
			"cancelled_trades":  summary.CancelledTrades,
			"total_pnl":         helpers.RoundFloat(summary.NetProfit, 2),
			"total_pnl_percent": helpers.RoundFloat(summary.NetProfitPercent, 2),
			"max_drawdown_pct":  helpers.RoundFloat(summary.MaxDrawdown, 2),
			"win_rate":          helpers.RoundFloat(summary.WinRate, 2),
			"profit_factor":     helpers.RoundFloat(summary.ProfitFactor, 2),
			"avg_win":           helpers.RoundFloat(summary.AvgWin, 2),
			"avg_loss":          helpers.RoundFloat(summary.AvgLoss, 2),
			"largest_win":       helpers.RoundFloat(summary.LargestWin, 2),
			"largest_loss":      helpers.RoundFloat(summary.LargestLoss, 2),
			"equity_curve_json": string(equityCurveJSON),
			"completed_at":      &completedAt,
			"updated_at":        completedAt,
		}
		if err := s.repo.Backtest.Update(tx, update, backtestID); err != nil {
			return nil, fmt.Errorf("failed to update backtest: %w", err)
		}

		// Insert all trades
		for _, trade := range simulationResult.Trades {
			trade.BacktestID = backtestID
			_, err = s.repo.BacktestTrade.Create(tx, &trade)
			if err != nil {
				return nil, fmt.Errorf("failed to save trade: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		fmt.Printf("❌ [BACKTEST WORKER] Failed to save results: %v\n", err)
		s.backtestWorkerFailed(backtestID, fmt.Errorf("failed to save results: %w", err))
		return
	}

	fmt.Printf("💾 [BACKTEST WORKER] Results saved to database\n")
	fmt.Printf("✅ [BACKTEST WORKER] Backtest ID: %d completed successfully\n", backtestID)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
}

// backtestWorkerFailed updates the backtest status to FAILED with error message
func (s *Services) backtestWorkerFailed(backtestID uint, err error) {
	now := time.Now()
	errorMsg := err.Error()

	_, updateErr := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		update := map[string]interface{}{
			"status":        "FAILED",
			"error_message": errorMsg,
			"updated_at":    now,
		}
		return nil, s.repo.Backtest.Update(tx, update, backtestID)
	})

	if updateErr != nil {
		fmt.Printf("❌ [BACKTEST WORKER] Failed to update status to FAILED: %v\n", updateErr)
	}

	fmt.Printf("❌ [BACKTEST WORKER] Backtest ID: %d failed: %v\n", backtestID, err)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
}

// ============================================================================
// Helper: Fetch Klines
// ============================================================================

// backtestFetchAllKlines fetches klines for all strategy timeframes
func (s *Services) backtestFetchAllKlines(
	symbol string,
	days int,
	startTime time.Time,
	endTime time.Time,
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

		// Use endTime from backtest period
		klines, err := s.backtestFetchKlinesBatched(symbol, tf.TimeframeName, startTime, endTime)
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

// backtestFetchKlinesBatched fetches klines in batches with automatic batch calculation
// This function handles pagination automatically when data exceeds Binance limit (1500)
// Parameters:
//   - symbol: Trading pair (e.g., BTCUSDT)
//   - interval: Timeframe (e.g., 1h, 15m)
//   - startTime: Start time for data range
//   - endTime: End time for data range (optional, if zero will fetch all available)
//
// Returns:
//   - []binance.KlineInfo: Sorted klines from startTime to endTime
//   - error: Any error encountered during fetch
func (s *Services) backtestFetchKlinesBatched(
	symbol string,
	interval string,
	startTime time.Time,
	endTime time.Time,
) ([]binance.KlineInfo, error) {
	startTimeMs := startTime.UnixMilli()
	endTimeMs := endTime.UnixMilli()

	// Calculate total batches needed based on time range
	// Binance returns max 1500 candles per request
	totalBatches := s.backtestCalculateBatches(startTimeMs, endTimeMs, interval)

	fmt.Printf("   ├── Fetching %s (%s): %d batches needed\n", symbol, interval, totalBatches)

	var allKlines []binance.KlineInfo
	currentStartTime := startTimeMs

	for batch := 0; batch < totalBatches; batch++ {
		// Fetch with limit 1500 (Binance max)
		klines, err := s.BinanceClient.GetKlinesWithStartTime(symbol, interval, backtestMaxBinanceLimit, currentStartTime)
		if err != nil {
			return nil, fmt.Errorf("batch %d failed: %w", batch+1, err)
		}

		if len(klines) == 0 {
			break
		}

		// Filter klines that exceed endTime
		for _, k := range klines {
			if k.OpenTime <= endTimeMs {
				allKlines = append(allKlines, k)
			}
		}

		// Check if we've reached the end time
		lastCandle := klines[len(klines)-1]
		if lastCandle.CloseTime >= endTimeMs {
			break
		}

		// Set start time for next batch (1ms after last candle's close time)
		currentStartTime = lastCandle.CloseTime + 1
	}

	// Sort by OpenTime (ascending)
	sort.Slice(allKlines, func(i, j int) bool {
		return allKlines[i].OpenTime < allKlines[j].OpenTime
	})

	return allKlines, nil
}

// backtestCalculateBatches calculates number of batches needed based on time range
// Formula: (endTime - startTime) / (interval_in_ms * 1500)
func (s *Services) backtestCalculateBatches(startTimeMs, endTimeMs int64, interval string) int {
	// Parse interval to get milliseconds
	intervalMs := s.backtestIntervalToMs(interval)

	// Total candles in range
	totalCandles := (endTimeMs - startTimeMs) / intervalMs

	// Calculate batches (1500 candles per batch)
	batches := int(math.Ceil(float64(totalCandles) / float64(backtestMaxBinanceLimit)))

	// Minimum 1 batch
	if batches < 1 {
		return 1
	}

	return batches
}

// backtestIntervalToMs converts interval string to milliseconds
func (s *Services) backtestIntervalToMs(interval string) int64 {
	// Remove suffix and parse number
	unit := interval[len(interval)-1:]
	numStr := interval[:len(interval)-1]

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 60000 // Default to 1m
	}

	switch unit {
	case "m":
		return num * 60 * 1000 // minutes to ms
	case "h":
		return num * 60 * 60 * 1000 // hours to ms
	case "d":
		return num * 24 * 60 * 60 * 1000 // days to ms
	case "w":
		return num * 7 * 24 * 60 * 60 * 1000 // weeks to ms
	case "M":
		return num * 30 * 24 * 60 * 60 * 1000 // months to ms (approximate)
	default:
		return 60000 // Default to 1m
	}
}

// ============================================================================
// Simulation Engine
// ============================================================================

// backtestRunSimulation runs the complete backtest simulation
func (s *Services) backtestRunSimulation(
	symbol string,
	initialCapital float64,
	strategy *dtos.StrategyData,
	allKlines map[string][]binance.KlineInfo,
	primaryKlines []binance.KlineInfo,
	thresholds []models.Threshold,
	mmConfig *config.MMConfig,
) simulationResult {
	var (
		completedTrades []models.BacktestTrade
		equityCurve     []equityPoint
		tradeCounter    int
	)

	// ─────────────────────────────────────────────────────────────────────
	// 1. INITIALIZATION
	// ─────────────────────────────────────────────────────────────────────
	wallet := &virtualWallet{
		Balance:        initialCapital,
		InitialBalance: initialCapital,
	}

	currentDate := ""
	dailyStats := &dailyStats{
		Date:              "",
		Count:             0,
		PnL:               0.0,
		SLHits:            0,
		TPHits:            0,
		ConsecutiveLosses: 0,
	}

	var currentActiveTrade *activeTrade = nil

	// Add initial equity point
	equityCurve = append(equityCurve, equityPoint{
		Timestamp: primaryKlines[0].OpenTime,
		Balance:   wallet.Balance,
		PnL:       0.0,
	})

	// Get expiration hours config
	expirationHours := int(mmConfig.ORDER_EXPIRATION_HOURS)
	if expirationHours <= 0 {
		expirationHours = 4
	}

	startIdx := backtestMinCandles
	if startIdx >= len(primaryKlines) {
		startIdx = len(primaryKlines) / 2
	}

	fmt.Printf("📊 [SIMULATION] Starting with Balance: $%.2f\n", wallet.Balance)
	fmt.Printf("📊 [SIMULATION] Expiration Hours: %dh\n", expirationHours)
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────
	// 2. MAIN LOOP: Iterate per candle
	// ─────────────────────────────────────────────────────────────────────
	for i := startIdx; i < len(primaryKlines); i++ {
		candle := primaryKlines[i]
		candleTime := time.UnixMilli(candle.OpenTime)
		candleDate := candleTime.Format("2006-01-02")

		// 2.2 Check day change & reset daily stats
		if candleDate != currentDate {
			if currentDate != "" {
				fmt.Printf("📅 [DAY CHANGE] %s -> %s | Daily PnL: %+.2f, Trades: %d\n",
					currentDate, candleDate, dailyStats.PnL, dailyStats.Count)
			}
			currentDate = candleDate
			dailyStats.Date = candleDate
			dailyStats.Count = 0
			dailyStats.PnL = 0.0
			dailyStats.SLHits = 0
			dailyStats.TPHits = 0
			dailyStats.ConsecutiveLosses = 0
		}

		// ─────────────────────────────────────────────────────────────────
		// 2.3 TRADE MONITOR PHASE
		// ─────────────────────────────────────────────────────────────────
		if currentActiveTrade != nil {
			// FASE 1: Check TP / SL
			exitReason := s.backtestCheckTPSL(currentActiveTrade, candle)
			if exitReason != "" {
				trade := s.backtestCloseTrade(currentActiveTrade, candle, exitReason, candleTime, wallet, dailyStats)
				completedTrades = append(completedTrades, trade)

				equityCurve = append(equityCurve, equityPoint{
					Timestamp: candle.OpenTime,
					Balance:   wallet.Balance,
					PnL:       wallet.Balance - wallet.InitialBalance,
				})

				currentActiveTrade = nil
				continue
			}

			// FASE 1.5: Cek Reverse Signal
			if currentActiveTrade.TotalQty > 0 {
				subsetKlinesRev := s.backtestBuildSubsetKlines(allKlines, candle.OpenTime)
				analyzeResultRev, errRev := s.signalAnalyzeCalculate(
					symbol,
					wallet.Balance,
					strategy,
					subsetKlinesRev,
					thresholds,
				)

				if errRev == nil && analyzeResultRev != nil {
					// if errRev == nil && analyzeResultRev.Signal.Valid {
					isReversed := false
					newSignal := analyzeResultRev.Signal.Signal

					if (currentActiveTrade.Side == "BUY" || currentActiveTrade.Side == "STRONG_BUY") && (newSignal == "SELL" || newSignal == "STRONG_SELL") {
						isReversed = true
					} else if (currentActiveTrade.Side == "SELL" || currentActiveTrade.Side == "STRONG_SELL") && (newSignal == "BUY" || newSignal == "STRONG_BUY") {
						isReversed = true
					}

					if isReversed {
						fmt.Printf("🚨 [REVERSE SIGNAL] Changing from %s to %s @ %s\n", currentActiveTrade.Side, newSignal, candleTime.Format("2006-01-02 15:04"))
						trade := s.backtestCloseTrade(currentActiveTrade, candle, "REVERSE_SIGNAL", candleTime, wallet, dailyStats)
						completedTrades = append(completedTrades, trade)

						equityCurve = append(equityCurve, equityPoint{
							Timestamp: candle.OpenTime,
							Balance:   wallet.Balance,
							PnL:       wallet.Balance - wallet.InitialBalance,
						})

						currentActiveTrade = nil
						continue
					}
				}
			}

			// FASE 2: Sync Pending Entries
			s.backtestSyncPendingEntries(currentActiveTrade, candle, candleTime, expirationHours)

			// FASE 3: Dead Signal Check
			if s.backtestIsDeadSignal(currentActiveTrade) {
				trade := s.backtestCreateDeadSignalTrade(currentActiveTrade, candleTime)
				completedTrades = append(completedTrades, trade)
				currentActiveTrade = nil
				continue
			}

			continue
		}

		// ─────────────────────────────────────────────────────────────────
		// 2.4 TRADE EXECUTE & SIGNAL ANALYZE PHASE
		// ─────────────────────────────────────────────────────────────────
		subsetKlines := s.backtestBuildSubsetKlines(allKlines, candle.OpenTime)

		analyzeResult, err := s.signalAnalyzeCalculate(
			symbol,
			wallet.Balance,
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

		if analyzeResult.Signal.TradingPlan == nil {
			continue
		}

		plan := analyzeResult.Signal.TradingPlan
		if plan.Mode == "WAIT" || len(plan.Entries) == 0 {
			continue
		}

		// 2.4.2 Validasi 5 Gates
		if !s.backtestValidate5Gates(wallet, dailyStats, plan, mmConfig) {
			continue
		}

		// 2.4.3 Execute Virtual Order
		tradeCounter++
		currentActiveTrade = s.backtestCreateActiveTrade(
			tradeCounter,
			action,
			signal,
			analyzeResult.Scoring.Confidence,
			plan,
			candle,
			candleTime,
			wallet,
			dailyStats,
			mmConfig,
		)

		dailyStats.Count++

		fmt.Printf("🎯 [TRADE #%d] %s signal opened @ %s\n", tradeCounter, action, candleTime.Format("2006-01-02 15:04"))
		fmt.Printf("   ├── Entries: %d | Mode: %s\n", len(plan.Entries), plan.Mode)
		fmt.Printf("   ├── TP: $%.4f | SL: $%.4f | R:R: %.2f\n", plan.TakeProfit, plan.StopLoss, plan.RiskRewardRatio)
		fmt.Printf("   └── Confidence: %.1f%%\n", analyzeResult.Scoring.Confidence)
		fmt.Println()
	}

	// ─────────────────────────────────────────────────────────────────────
	// 3. FINALIZATION
	// ─────────────────────────────────────────────────────────────────────
	lastCandle := primaryKlines[len(primaryKlines)-1]
	lastTime := time.UnixMilli(lastCandle.CloseTime)

	if currentActiveTrade != nil {
		trade := s.backtestCloseTrade(currentActiveTrade, lastCandle, exitReasonClosedEnd, lastTime, wallet, dailyStats)
		completedTrades = append(completedTrades, trade)

		equityCurve = append(equityCurve, equityPoint{
			Timestamp: lastCandle.OpenTime,
			Balance:   wallet.Balance,
			PnL:       wallet.Balance - wallet.InitialBalance,
		})

		fmt.Printf("⏹️  [FORCE CLOSE] Closing active trade @ $%.4f\n", lastCandle.Close)
		fmt.Println()
	}

	summary := s.backtestCalculateSummary(completedTrades, initialCapital)

	fmt.Printf("📊 [SIMULATION COMPLETE] Final Balance: $%.2f (PnL: %+.2f)\n",
		wallet.Balance, wallet.Balance-wallet.InitialBalance)

	return simulationResult{
		Trades:      completedTrades,
		EquityCurve: equityCurve,
		Summary:     summary,
	}
}

// backtestBuildSubsetKlines builds klines subset up to current candle
func (s *Services) backtestBuildSubsetKlines(allKlines map[string][]binance.KlineInfo, currentOpenTime int64) map[string][]binance.KlineInfo {
	subsetKlines := make(map[string][]binance.KlineInfo)
	for tf, klines := range allKlines {
		subset := make([]binance.KlineInfo, 0)
		for _, k := range klines {
			if k.OpenTime <= currentOpenTime {
				subset = append(subset, k)
			}
		}
		if len(subset) > 0 {
			subsetKlines[tf] = subset
		}
	}
	return subsetKlines
}

// backtestValidate5Gates validates the 5 gates from TRADE_EXECUTE_FLOW
func (s *Services) backtestValidate5Gates(
	wallet *virtualWallet,
	dailyStats *dailyStats,
	plan *dtos.TradingPlan,
	mmConfig *config.MMConfig,
) bool {
	// Gate 1.5: Minimum Risk/Reward Ratio Hard Limit
	if plan.RiskRewardRatio < float64(mmConfig.RISK_REWARD_RATIO) {
		fmt.Printf("🚫 [GATE 1.5] R:R too low: %.2f < %.2f\n", 
			plan.RiskRewardRatio, mmConfig.RISK_REWARD_RATIO)
		return false
	}

	// Gate 2A: Consecutive Loss Limit
	if dailyStats.ConsecutiveLosses >= int(mmConfig.MAX_DAILY_LOSS_COUNT) {
		fmt.Printf("🚫 [GATE 2A] Consecutive losses limit reached: %d >= %d\n",
			dailyStats.ConsecutiveLosses, mmConfig.MAX_DAILY_LOSS_COUNT)
		return false
	}

	// Gate 2B: Daily Loss Percentage
	maxDailyLoss := wallet.Balance * float64(mmConfig.MAX_DAILY_LOSS_PERCENT)
	if dailyStats.PnL < 0 && math.Abs(dailyStats.PnL) >= maxDailyLoss {
		fmt.Printf("🚫 [GATE 2B] Daily loss limit reached: %.2f >= %.2f\n",
			math.Abs(dailyStats.PnL), maxDailyLoss)
		return false
	}

	// Gate 3: Minimum Balance
	minBalance := wallet.Balance * 0.98
	if minBalance < 3.0 {
		fmt.Printf("🚫 [GATE 3] Insufficient balance: %.2f < 3.0\n", minBalance)
		return false
	}

	// Gate 5: Daily Trade Count (with R:R exception)
	if dailyStats.Count >= int(mmConfig.MAX_DAILY_TRADES) {
		if plan.RiskRewardRatio < float64(mmConfig.RISK_REWARD_TARGET) {
			fmt.Printf("🚫 [GATE 5] Daily trades limit reached: %d >= %d (R:R %.2f < target %.2f)\n",
				dailyStats.Count, mmConfig.MAX_DAILY_TRADES, plan.RiskRewardRatio, mmConfig.RISK_REWARD_TARGET)
			return false
		}
		fmt.Printf("✅ [GATE 5] Exception: High R:R trade (%.2f >= %.2f)\n",
			plan.RiskRewardRatio, mmConfig.RISK_REWARD_TARGET)
	}

	return true
}

// backtestCreateActiveTrade creates a new active trade from trading plan
func (s *Services) backtestCreateActiveTrade(
	tradeNum int,
	side string,
	signal string,
	confidence float64,
	plan *dtos.TradingPlan,
	candle binance.KlineInfo,
	candleTime time.Time,
	wallet *virtualWallet,
	dailyStats *dailyStats,
	mmConfig *config.MMConfig,
) *activeTrade {
	trade := &activeTrade{
		TradeNum:        tradeNum,
		Side:            side,
		Signal:          signal,
		Confidence:      confidence,
		TradingMode:     plan.Mode,
		Entries:         make([]*tradeEntry, 0),
		TotalQty:        0,
		AvgEntryPrice:   0,
		TakeProfit:      plan.TakeProfit,
		StopLoss:        plan.StopLoss,
		RiskRewardRatio: plan.RiskRewardRatio,
		EntryTime:       candleTime,
		FilledTime:      nil,
		TPPrice:         plan.TakeProfit,
		SLPrice:         plan.StopLoss,
		DailyTradeCount: dailyStats.Count,
		DailyPnL:        dailyStats.PnL,
		ConsecutiveLoss: dailyStats.ConsecutiveLosses,
	}

	// Calculate capital allocation based on signal strength
	strengthPercent := 1.0
	if confidence < 70 {
		strengthPercent = 0.8
	}

	maxPositionValue := wallet.Balance * float64(mmConfig.MAX_POSITION_SIZE)
	capitalAllocation := maxPositionValue * strengthPercent

	// Process each entry from trading plan
	for _, entry := range plan.Entries {
		entryPrice := entry.EntryPrice
		entryType := "LIMIT"
		entryStatus := "PENDING"
		var filledTime *time.Time

		// Aggressive mode: Entry 1 is MARKET order (instant fill)
		if mmConfig.IS_AGRESSIVE && entry.EntryNumber == 1 {
			entryPrice = candle.Close
			entryType = "MARKET"
			entryStatus = "FILLED"
			ft := candleTime
			filledTime = &ft

			if trade.FilledTime == nil {
				trade.FilledTime = filledTime
			}
		}

		tradeEntry := &tradeEntry{
			EntryNum:  entry.EntryNumber,
			Type:      entryType,
			Price:     entryPrice,
			Qty:       entry.PositionQty,
			Timestamp: candleTime,
			Status:    entryStatus,
			CreatedAt: candleTime,
		}

		trade.Entries = append(trade.Entries, tradeEntry)

		// If filled, update trade totals
		if entryStatus == "FILLED" {
			trade.TotalQty += entry.PositionQty
			if trade.AvgEntryPrice == 0 {
				trade.AvgEntryPrice = entryPrice
			} else {
				totalValue := (trade.AvgEntryPrice * (trade.TotalQty - entry.PositionQty)) + (entryPrice * entry.PositionQty)
				trade.AvgEntryPrice = totalValue / trade.TotalQty
			}
		}
	}

	trade.CapitalAllocated = capitalAllocation

	return trade
}

// backtestCheckTPSL checks if TP or SL is hit on a candle
func (s *Services) backtestCheckTPSL(trade *activeTrade, candle binance.KlineInfo) string {
	if trade.TotalQty == 0 {
		return ""
	}

	if trade.Side == "BUY" {
		if trade.StopLoss > 0 && candle.Low <= trade.StopLoss {
			return exitReasonHitSL
		}
		if trade.TakeProfit > 0 && candle.High >= trade.TakeProfit {
			return exitReasonHitTP
		}
	} else {
		if trade.StopLoss > 0 && candle.High >= trade.StopLoss {
			return exitReasonHitSL
		}
		if trade.TakeProfit > 0 && candle.Low <= trade.TakeProfit {
			return exitReasonHitTP
		}
	}
	return ""
}

// backtestSyncPendingEntries checks and fills pending limit orders
func (s *Services) backtestSyncPendingEntries(trade *activeTrade, candle binance.KlineInfo, candleTime time.Time, expirationHours int) {
	for _, entry := range trade.Entries {
		if entry.Status != "PENDING" {
			continue
		}

		// Check expired
		elapsed := candleTime.Sub(entry.CreatedAt)
		if elapsed.Hours() >= float64(expirationHours) {
			entry.Status = "CANCELLED"
			fmt.Printf("⏰ [ENTRY #%d] LIMIT order expired @ %s\n", entry.EntryNum, candleTime.Format("2006-01-02 15:04"))
			continue
		}

		// Check if price hit entry
		filled := false
		if trade.Side == "BUY" {
			filled = candle.Low <= entry.Price
		} else {
			filled = candle.High >= entry.Price
		}

		if filled {
			entry.Status = "FILLED"
			entry.Timestamp = candleTime

			if trade.FilledTime == nil {
				trade.FilledTime = &candleTime
			}

			// Auto-Adapt TP/SL: Update TotalQty and recalculate AvgEntryPrice
			oldQty := trade.TotalQty
			oldAvgPrice := trade.AvgEntryPrice

			trade.TotalQty += entry.Qty

			if oldQty == 0 {
				trade.AvgEntryPrice = entry.Price
			} else {
				totalValue := (oldAvgPrice * oldQty) + (entry.Price * entry.Qty)
				trade.AvgEntryPrice = totalValue / trade.TotalQty
			}

			fmt.Printf("✅ [ENTRY #%d] LIMIT filled @ $%.4f (Total Qty: %.4f, Avg: $%.4f)\n",
				entry.EntryNum, entry.Price, trade.TotalQty, trade.AvgEntryPrice)
		}
	}
}

// backtestIsDeadSignal checks if all entries are cancelled and no position opened
func (s *Services) backtestIsDeadSignal(trade *activeTrade) bool {
	if trade.TotalQty > 0 {
		return false
	}

	for _, entry := range trade.Entries {
		if entry.Status != "CANCELLED" {
			return false
		}
	}

	fmt.Printf("💀 [DEAD SIGNAL] All entries cancelled, closing trade #%d\n", trade.TradeNum)
	return true
}

// backtestCloseTrade closes an active trade and updates wallet & stats
func (s *Services) backtestCloseTrade(
	trade *activeTrade,
	candle binance.KlineInfo,
	exitReason string,
	exitTime time.Time,
	wallet *virtualWallet,
	dailyStats *dailyStats,
) models.BacktestTrade {
	var exitPrice float64

	switch exitReason {
	case exitReasonHitTP:
		exitPrice = trade.TakeProfit
	case exitReasonHitSL:
		exitPrice = trade.StopLoss
	default:
		exitPrice = candle.Close
	}

	// Calculate PnL
	var pnl float64
	if trade.TotalQty > 0 {
		if trade.Side == "BUY" {
			pnl = (exitPrice - trade.AvgEntryPrice) * trade.TotalQty
		} else {
			pnl = (trade.AvgEntryPrice - exitPrice) * trade.TotalQty
		}
	}

	// Apply taker fee (0.04% for Binance Futures)
	takerFee := 0.0004
	entryValue := trade.AvgEntryPrice * trade.TotalQty
	exitValue := exitPrice * trade.TotalQty
	totalFee := (entryValue + exitValue) * takerFee
	pnl -= totalFee

	pnlPercent := 0.0
	if trade.CapitalAllocated > 0 {
		pnlPercent = (pnl / trade.CapitalAllocated) * 100
	}

	// Update wallet
	wallet.Balance += pnl

	// Update daily stats
	dailyStats.PnL += pnl
	if pnl < 0 {
		dailyStats.ConsecutiveLosses++
		dailyStats.SLHits++
	} else if pnl > 0 {
		dailyStats.ConsecutiveLosses = 0
		dailyStats.TPHits++
	}

	// Calculate duration
	var durationMinutes int64
	if trade.FilledTime != nil {
		durationMinutes = int64(exitTime.Sub(*trade.FilledTime).Minutes())
	}

	// Convert entries to JSON
	entriesDTO := make([]dtos.TradeEntry, len(trade.Entries))
	for i, e := range trade.Entries {
		entriesDTO[i] = dtos.TradeEntry{
			EntryNum:  e.EntryNum,
			Type:      e.Type,
			Price:     e.Price,
			Qty:       e.Qty,
			Timestamp: e.Timestamp,
			Status:    e.Status,
		}
	}
	entriesJSON, _ := json.Marshal(entriesDTO)

	exitTimePtr := &exitTime
	filledTimePtr := trade.FilledTime

	return models.BacktestTrade{
		TradeNum:        trade.TradeNum,
		Side:            trade.Side,
		Signal:          trade.Signal,
		Confidence:      trade.Confidence,
		TradingMode:     trade.TradingMode,
		TakeProfit:      trade.TakeProfit,
		StopLoss:        trade.StopLoss,
		RiskRewardRatio: trade.RiskRewardRatio,
		EntryTime:       trade.EntryTime,
		FilledTime:      filledTimePtr,
		ExitTime:        exitTimePtr,
		TotalQty:        trade.TotalQty,
		AvgEntryPrice:   trade.AvgEntryPrice,
		ExitPrice:       exitPrice,
		TotalCapital:    trade.CapitalAllocated,
		PnL:             helpers.RoundFloat(pnl, 2),
		PnLPercent:      helpers.RoundFloat(pnlPercent, 2),
		ExitReason:      exitReason,
		Status:          "CLOSED",
		DurationMinutes: durationMinutes,
		EntriesJSON:     string(entriesJSON),
		DailyTradeCount: trade.DailyTradeCount,
		DailyPnL:        trade.DailyPnL,
		ConsecutiveLoss: trade.ConsecutiveLoss,
	}
}

// backtestCreateDeadSignalTrade creates a trade record for dead signal
func (s *Services) backtestCreateDeadSignalTrade(trade *activeTrade, exitTime time.Time) models.BacktestTrade {
	entriesDTO := make([]dtos.TradeEntry, len(trade.Entries))
	for i, e := range trade.Entries {
		entriesDTO[i] = dtos.TradeEntry{
			EntryNum:  e.EntryNum,
			Type:      e.Type,
			Price:     e.Price,
			Qty:       e.Qty,
			Timestamp: e.Timestamp,
			Status:    e.Status,
		}
	}
	entriesJSON, _ := json.Marshal(entriesDTO)

	exitTimePtr := &exitTime

	return models.BacktestTrade{
		TradeNum:        trade.TradeNum,
		Side:            trade.Side,
		Signal:          trade.Signal,
		Confidence:      trade.Confidence,
		TradingMode:     trade.TradingMode,
		TakeProfit:      trade.TakeProfit,
		StopLoss:        trade.StopLoss,
		RiskRewardRatio: trade.RiskRewardRatio,
		EntryTime:       trade.EntryTime,
		FilledTime:      nil,
		ExitTime:        exitTimePtr,
		TotalQty:        0,
		AvgEntryPrice:   0,
		ExitPrice:       0,
		TotalCapital:    0,
		PnL:             0,
		PnLPercent:      0,
		ExitReason:      exitReasonDeadSignal,
		Status:          "CANCELLED",
		DurationMinutes: 0,
		EntriesJSON:     string(entriesJSON),
		DailyTradeCount: trade.DailyTradeCount,
		DailyPnL:        trade.DailyPnL,
		ConsecutiveLoss: trade.ConsecutiveLoss,
	}
}

// backtestCalculateSummary calculates comprehensive backtest summary
func (s *Services) backtestCalculateSummary(trades []models.BacktestTrade, initialCapital float64) dtos.BacktestSummary {
	summary := dtos.BacktestSummary{
		InitialBalance: initialCapital,
	}

	if len(trades) == 0 {
		return summary
	}

	var totalProfit, totalLoss float64
	var totalDuration int64
	filledCount := 0
	winningCount := 0

	for _, trade := range trades {
		switch trade.ExitReason {
		case exitReasonExpired:
			summary.ExpiredTrades++
		case exitReasonDeadSignal:
			summary.CancelledTrades++
		default:
			filledCount++
			summary.TotalTrades++
			summary.NetProfit += trade.PnL
			totalDuration += trade.DurationMinutes

			if trade.PnL > 0 {
				totalProfit += trade.PnL
				winningCount++
				summary.WinningTrades++
				if trade.PnL > summary.LargestWin {
					summary.LargestWin = trade.PnL
				}
			} else if trade.PnL < 0 {
				totalLoss += math.Abs(trade.PnL)
				summary.LosingTrades++
				if math.Abs(trade.PnL) > summary.LargestLoss {
					summary.LargestLoss = math.Abs(trade.PnL)
				}
			}
		}
	}

	// Derived metrics
	summary.FinalBalance = initialCapital + summary.NetProfit
	summary.NetProfitPercent = (summary.NetProfit / initialCapital) * 100

	if filledCount > 0 {
		summary.WinRate = helpers.RoundFloat(float64(winningCount)/float64(filledCount)*100, 2)
		if winningCount > 0 {
			summary.AvgWin = helpers.RoundFloat(totalProfit/float64(winningCount), 2)
		}
		losingCount := filledCount - winningCount
		if losingCount > 0 {
			summary.AvgLoss = helpers.RoundFloat(totalLoss/float64(losingCount), 2)
		}
	}

	// Profit factor
	if totalLoss > 0 {
		summary.ProfitFactor = helpers.RoundFloat(totalProfit/totalLoss, 2)
	} else if totalProfit > 0 {
		summary.ProfitFactor = 999.99
	}

	// Max drawdown
	var runningPnL, peakPnL float64
	for _, trade := range trades {
		if trade.ExitReason != exitReasonExpired && trade.ExitReason != exitReasonDeadSignal {
			runningPnL += trade.PnL
			if runningPnL > peakPnL {
				peakPnL = runningPnL
			}
			drawdown := peakPnL - runningPnL
			if drawdown > summary.MaxDrawdown {
				summary.MaxDrawdown = drawdown
			}
		}
	}

	if initialCapital > 0 {
		summary.MaxDrawdown = (summary.MaxDrawdown / initialCapital) * 100
	}

	// Round values
	summary.NetProfit = helpers.RoundFloat(summary.NetProfit, 2)
	summary.NetProfitPercent = helpers.RoundFloat(summary.NetProfitPercent, 2)
	summary.MaxDrawdown = helpers.RoundFloat(summary.MaxDrawdown, 2)

	return summary
}
