package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
)

var (
	tradeBotActive       bool                  // Is trade bot service active?
	tradeBotCancel       context.CancelFunc    // Cancel function to stop goroutines
	tradeBotMutex        sync.Mutex            // Mutex for thread safety
	tradeBotStrategy     *uint                 // Active strategy ID
	tradeBotTime         *time.Time            // Timestamp when bot was activated
	tradeExecutorRunning bool                  // Track if trade execution cycle is currently running
	tradeMonitorRunning  bool                  // Track if trade monitor cycle is currently running
	tradeExecutorLogger  *helpers.WorkerLogger // Logger for trade executor
	tradeMonitorLogger   *helpers.WorkerLogger // Logger for trade monitor
)

// TradeBotActivate activates the automated trade bot
// Starts 2 background workers:
// 1. Trade Executor - Scans watchlist & executes new trades (interval: strategy timeframe)
// 2. Trade Monitor - Monitors active trades (interval: 1 minute)
func (s *Services) TradeBotActivate(ctx *gin.Context, strategyID *uint) (res map[string]interface{}, err error) {
	tradeBotMutex.Lock()
	defer tradeBotMutex.Unlock()

	if tradeBotActive {
		return nil, fmt.Errorf("trade bot is already active. Deactivate first before activating again")
	}

	// 🧹 Clean up old logs (older than 30 days)
	helpers.CleanupOldLogs()

	// Get strategy - optimize query: use provided ID or fallback to active strategy
	var strategy *dtos.StrategyData
	if strategyID != nil && *strategyID > 0 {
		strategy, err = s.StrategyGetByID(ctx, *strategyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get strategy: %w", err)
		}
	} else {
		strategy, err = s.StrategyGetActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active strategy: %w", err)
		}
	}

	// Get primary timeframe details
	timeframeName := strategy.PrimaryTF
	timeframe, err := s.repo.Timeframe.FindByName(nil, timeframeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary timeframe %s: %w", timeframeName, err)
	}

	// Use timeframe in_minutes as trade execution interval
	executionInterval := time.Duration(timeframe.InMinutes) * time.Minute

	// Create logger for trade executor session (Executor)
	tradeExecutorLogger, err = helpers.NewWorkerLogger("EXECUTOR", "executor")
	if err != nil {
		return nil, fmt.Errorf("failed to create trade executor logger: %w", err)
	}

	// Create logger for trade monitor session (Monitor)
	tradeMonitorLogger, err = helpers.NewWorkerLogger("MONITOR", "monitor")
	if err != nil {
		tradeExecutorLogger.Close() // Cleanup previous logger
		return nil, fmt.Errorf("failed to create trade monitor logger: %w", err)
	}

	// Create context with cancel
	ctxBot, cancel := context.WithCancel(context.Background())
	tradeBotCancel = cancel
	tradeBotActive = true
	now := time.Now()
	tradeBotTime = &now
	tradeBotStrategy = &strategy.ID

	// Start background goroutine 1: Trade Executor (dynamic interval based on strategy timeframe)
	go s.runBackgroundTradeExecutor(ctxBot, executionInterval, tradeExecutorLogger)

	// Start background goroutine 2: Trade Monitor (fixed 1 minute interval)
	monitorInterval := 1 * time.Minute
	go s.runBackgroundTradeMonitor(ctxBot, monitorInterval, tradeMonitorLogger)

	return map[string]interface{}{
		"is_active":          true,
		"execution_interval": executionInterval.Minutes(),
		"monitor_interval":   1,
		"started_at":         tradeBotTime.Format(time.RFC3339),
	}, nil
}

// TradeBotDeactivate deactivates the automated trade bot
func (s *Services) TradeBotDeactivate(ctx *gin.Context) (res map[string]interface{}, err error) {
	tradeBotMutex.Lock()
	defer tradeBotMutex.Unlock()

	if !tradeBotActive {
		return nil, fmt.Errorf("trade bot is not active")
	}

	// Cancel the context
	if tradeBotCancel != nil {
		tradeBotCancel()
	}
	tradeBotActive = false
	tradeExecutorRunning = false // Reset executor flag
	tradeMonitorRunning = false  // Reset monitor flag
	tradeBotTime = nil           // Reset bot time

	// Close logger file handle for trade executor
	if tradeExecutorLogger != nil {
		tradeExecutorLogger.Banner("🔴 TRADE EXECUTOR STOPPED",
			fmt.Sprintf("⏱  Stopped at: %s", time.Now().Format("2006-01-02 15:04:05")))
		tradeExecutorLogger.Close()
		tradeExecutorLogger = nil
	}

	// Close logger file handle for trade monitor
	if tradeMonitorLogger != nil {
		tradeMonitorLogger.Banner("🔴 TRADE MONITOR STOPPED",
			fmt.Sprintf("⏱  Stopped at: %s", time.Now().Format("2006-01-02 15:04:05")))
		tradeMonitorLogger.Close()
		tradeMonitorLogger = nil
	}

	return map[string]interface{}{
		"is_active": false,
	}, nil
}

// TradeBotGetStatus gets the current trade bot status
func (s *Services) TradeBotGetStatus(ctx *gin.Context) (res map[string]interface{}, err error) {
	tradeBotMutex.Lock()
	defer tradeBotMutex.Unlock()

	// If bot is not active, return simple status
	if !tradeBotActive {
		return map[string]interface{}{
			"is_active": false,
			"strategy":  nil,
		}, nil
	}

	// If bot is active, get strategy details
	var strategyData *dtos.StrategyData
	if tradeBotStrategy != nil {
		strategyData, err = s.StrategyGetByID(ctx, *tradeBotStrategy)
		if err != nil {
			// Log error but don't fail the entire request
			fmt.Printf("Warning: Failed to get strategy %d: %v\n", *tradeBotStrategy, err)
		}
	}

	// Calculate bot running duration
	var botRunningDuration string
	var botRunningSeconds float64
	if tradeBotTime != nil {
		duration := time.Since(*tradeBotTime)
		botRunningDuration = duration.String()
		botRunningSeconds = duration.Seconds()
	}

	return map[string]interface{}{
		"is_active":            tradeBotActive,
		"strategy":             strategyData,
		"bot_started_at":       tradeBotTime,
		"bot_running_duration": botRunningDuration,
		"bot_running_seconds":  botRunningSeconds,
	}, nil
}

// runBackgroundTradeExecutor runs the background trade execution process
// This worker scans watchlist symbols and executes new trades via TradeExecute()
// Interval: Dynamic based on strategy timeframe (e.g., 15m, 1h, 4h)
func (s *Services) runBackgroundTradeExecutor(ctx context.Context, executionInterval time.Duration, logger *helpers.WorkerLogger) {
	// Add panic recovery for goroutine
	defer func() {
		if r := recover(); r != nil {
			logger.Error("CRITICAL PANIC in trade executor goroutine: %v", r)
			// Reset flags as safety net if goroutine crashes
			tradeBotMutex.Lock()
			tradeBotActive = false
			tradeExecutorRunning = false
			tradeBotMutex.Unlock()
		}
	}()

	// Print start banner (console + file)
	logger.Banner(
		fmt.Sprintf("🟢 TRADE EXECUTOR STARTED — %s", logger.StartedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("⏱  Interval: %.0fm | Strategy: %d", executionInterval.Minutes(), *tradeBotStrategy),
	)

	// ✅ RUN EXECUTION FIRST CYCLE IMMEDIATELY
	logger.Info("Running initial trade execution immediately...")
	s.runTradeExecutionCycle(logger)

	// ✅ Loop: always align to next clock mark (e.g., :00, :15, :30, :45)
	for {
		// Calculate delay until next interval mark
		delay := s.calculateNextIntervalDelay(executionInterval)
		nextExecution := time.Now().Add(delay)
		logger.Info("Next trade execution at %s (waiting %.0fm)",
			nextExecution.Format("2006-01-02 15:04:05"), delay.Minutes())

		// Wait until next mark, with cancellation support
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			logger.Info("Trade executor stopped by user request")
			return
		case <-timer.C:
			logger.Info("Trade execution triggered at %s", time.Now().Format("2006-01-02 15:04:05"))

			// Skip if execution is already running
			tradeBotMutex.Lock()
			if tradeExecutorRunning {
				tradeBotMutex.Unlock()
				logger.Skip("Previous trade execution still running - skipping this interval")
				continue
			}
			tradeExecutorRunning = true
			tradeBotMutex.Unlock()

			// Run trade execution (tradeExecutorRunning will be reset by runTradeExecutionCycle's defer)
			s.runTradeExecutionCycle(logger)
		}
	}
}

// runBackgroundTradeMonitor runs the background trade monitoring process
// This worker checks all active trades for TP/SL hits, syncs entries, and performs netting
// Interval: Fixed 1 minute
func (s *Services) runBackgroundTradeMonitor(ctx context.Context, monitorInterval time.Duration, logger *helpers.WorkerLogger) {
	// Add panic recovery for goroutine
	defer func() {
		if r := recover(); r != nil {
			logger.Error("CRITICAL PANIC in trade monitor goroutine: %v", r)
		}
	}()

	// Print start banner (console + file)
	logger.Banner(
		fmt.Sprintf("🟢 TRADE MONITOR STARTED — %s", logger.StartedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("⏱  Interval: %.0fm", monitorInterval.Minutes()),
	)

	// Ticker for fixed interval (1 minute)
	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	// Initial run immediately
	s.runTradeMonitorCycle(logger)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Trade monitor stopped by user request")
			return
		case <-ticker.C:
			// Default interval triggered
			tradeBotMutex.Lock()
			if tradeMonitorRunning {
				tradeBotMutex.Unlock()
				logger.Skip("Previous trade monitor cycle still active - skipping this interval")
				continue
			}
			tradeMonitorRunning = true
			tradeBotMutex.Unlock()

			s.runTradeMonitorCycle(logger)
		}
	}
}

// calculateNextIntervalDelay calculates duration until the next interval mark
// Uses minutes since midnight to support any interval (5m, 15m, 1h, 4h, etc.)
// Example: if interval=15m and current time=00:05, returns 10m (until 00:15)
// Example: if interval=4h and current time=09:25, returns 2h35m (until 12:00)
func (s *Services) calculateNextIntervalDelay(interval time.Duration) time.Duration {
	now := time.Now()
	intervalMinutes := int(interval.Minutes())
	if intervalMinutes <= 0 {
		return 0
	}

	// Calculate total minutes since midnight
	totalMinutes := now.Hour()*60 + now.Minute()

	// Calculate next interval mark (in minutes since midnight)
	nextMark := ((totalMinutes / intervalMinutes) + 1) * intervalMinutes

	// Calculate target time from start of today
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location())
	target := startOfDay.Add(time.Duration(nextMark) * time.Minute)

	// If nextMark exceeds 24h (1440 minutes), it rolls over to next day automatically
	// since we're adding minutes to startOfDay

	// Safety check: if target is somehow in the past, add one interval
	if target.Before(now) {
		target = target.Add(interval)
	}

	return target.Sub(now)
}

// runTradeExecutionCycle runs a single trade execution cycle
// This function:
// 1. Gets all active symbols from watchlist
// 2. Calls TradeExecute() for each symbol to analyze signal & execute trade
func (s *Services) runTradeExecutionCycle(logger *helpers.WorkerLogger) {
	cycleStart := time.Now()

	// 1. Get all active symbols from watchlist
	watchlists, err := s.repo.Watchlist.FindAllActive(nil)
	if err != nil {
		logger.Error("Failed to get watchlist: %v", err)
		return
	}

	if len(watchlists) == 0 {
		logger.Warn("No active symbols in watchlist")
		return
	}

	logger.Success("Found %d active symbols in watchlist", len(watchlists))
	logger.CycleStart(len(watchlists))
	logger.Info("Execution list: %s", extractSymbolsFromWatchlist(watchlists))

	// Track trade stats
	tradesExecuted := 0
	tradesSkipped := 0

	// Ensure tradeExecutorRunning is reset when cycle completes (panic recovery)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in runTradeExecutionCycle: %v", r)
		}
		duration := time.Since(cycleStart)
		logger.CycleSummary(len(watchlists), tradesExecuted, tradesSkipped, duration)
		tradeBotMutex.Lock()
		tradeExecutorRunning = false
		tradeBotMutex.Unlock()
	}()

	// 2. Execute trade for each symbol
	for _, wl := range watchlists {
		// Check if trade bot is still active
		tradeBotMutex.Lock()
		if !tradeBotActive {
			tradeBotMutex.Unlock()
			logger.Warn("Trade bot deactivated during execution cycle")
			return
		}
		tradeBotMutex.Unlock()

		logger.Info("Analyzing & Executing: %s", wl.Symbol)

		// 3. Call TradeExecute for each symbol
		// TradeExecute will:
		// - Analyze signal with confidence scoring
		// - Validate against money management rules
		// - Execute trade on Binance if signal is valid
		tradeReq := &dtos.TradeRequest{
			Symbol:     wl.Symbol,
			StrategyID: *tradeBotStrategy,
		}

		// Create mock gin context for TradeExecute
		ginCtx := &gin.Context{}

		// Execute trade
		tradeRes, err := s.TradeExecute(ginCtx, tradeReq)

		if err != nil {
			logger.Error("Trade execution failed for %s: %v", wl.Symbol, err)
			tradesSkipped++
		} else if tradeRes != nil && tradeRes.ExecutionInfo.Executed {
			logger.Trade(true, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesExecuted++
		} else {
			logger.Trade(false, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesSkipped++
		}

		// Delay between symbols to avoid rate limiting
		if len(watchlists) > 1 {
			time.Sleep(30 * time.Second)
		}
	}
}

// runTradeMonitorCycle runs a single trade monitor check cycle
// This function:
// 1. Calls TradeMonitorProcessAllActive() to check all active trades
// 2. Logs results for each trade (TP/SL hit, entry sync, netting)
func (s *Services) runTradeMonitorCycle(logger *helpers.WorkerLogger) {
	cycleStart := time.Now()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC recovered in runTradeMonitorCycle: %v", r)
		}
		duration := time.Since(cycleStart)

		tradeBotMutex.Lock()
		tradeMonitorRunning = false
		tradeBotMutex.Unlock()

		logger.Info("Trade monitor cycle completed in %v", duration)
	}()

	logger.Info("Running trade monitor check for all active trades...")

	// Run trade monitor (using mock/background context)
	mockCtx := &gin.Context{}
	results, err := s.TradeMonitorProcessAllActive(mockCtx)

	if err != nil {
		logger.Error("Failed to monitor active trades: %v", err)
		return
	}

	// Calculate statistics from results
	processed := 0
	errors := 0
	for _, res := range results {
		// Log detailed execution flow
		if len(res.Logs) > 0 {
			logger.Info("Flow Logs for Trade ID %d (%s):", res.TradeID, res.Symbol)
			for _, flowLog := range res.Logs {
				logger.Info("  -> %s", flowLog)
			}
		}

		if res.Status == "ERROR" {
			errors++
			logger.Error("Trade ID %d (%s) failed: %s", res.TradeID, res.Symbol, res.Message)
		} else if res.Status != "SKIPPED" {
			processed++
			logger.Success("Trade ID %d (%s) processed. Status: %s. Msg: %s", res.TradeID, res.Symbol, res.Status, res.Message)
		}
	}

	logger.Info("Monitor cycle done. Processed: %d, Errors: %d", processed, errors)
}

// extractSymbolsFromWatchlist extracts symbol names from watchlist slice for logging
func extractSymbolsFromWatchlist(watchlists []models.Watchlist) string {
	if len(watchlists) == 0 {
		return "0 symbols"
	}

	symbols := make([]string, 0, len(watchlists))
	for _, wl := range watchlists {
		symbols = append(symbols, wl.Symbol)
	}

	if len(symbols) <= 5 {
		result := ""
		for i, s := range symbols {
			if i > 0 {
				result += ", "
			}
			result += s
		}
		return result
	}

	// Show first 3 and last 2 if more than 5
	result := symbols[0]
	for i := 1; i < 3 && i < len(symbols); i++ {
		result += ", " + symbols[i]
	}
	result += ", ..., " + symbols[len(symbols)-2] + ", " + symbols[len(symbols)-1]
	return result
}

// TradeBotGetSessionSummary gets summary statistics for current trading session
func (s *Services) TradeBotGetSessionSummary(ctx *gin.Context) (res map[string]interface{}, err error) {
	tradeBotMutex.Lock()
	defer tradeBotMutex.Unlock()

	// Validate bot is active
	if !tradeBotActive || tradeBotTime == nil {
		return nil, fmt.Errorf("trade bot is not running. Please activate the bot first")
	}

	// Get trades from current session (using bot start time)
	trades, err := s.repo.Trade.FindTradesInSession(nil, *tradeBotTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get session trades: %w", err)
	}

	// Calculate statistics
	totalTrades := len(trades)
	executed := 0
	skipped := 0
	totalPnL := 0.0
	symbolsMap := make(map[string]bool)

	for _, trade := range trades {
		// Track symbols
		symbolsMap[trade.Symbol] = true

		// Count executed vs skipped
		if trade.Status == "ACTIVE" || trade.Status == "COMPLETED" {
			executed++
		} else {
			skipped++
		}

		// Calculate PnL for closed trades
		if trade.PnL != 0 {
			totalPnL += trade.PnL
		}
	}

	// Calculate success rate
	successRate := 0.0
	if executed > 0 {
		// Count profitable trades
		profitable := 0
		for _, trade := range trades {
			if trade.Status == "COMPLETED" && trade.PnL > 0 {
				profitable++
			}
		}
		successRate = float64(profitable) / float64(executed) * 100
	}

	// Convert symbols map to slice
	symbols := make([]string, 0, len(symbolsMap))
	for symbol := range symbolsMap {
		symbols = append(symbols, symbol)
	}

	return map[string]interface{}{
		"total_trades":    totalTrades,
		"executed":        executed,
		"skipped":         skipped,
		"success_rate":    successRate,
		"total_pnl":       totalPnL,
		"symbols_traded":  symbols,
		"session_started": tradeBotTime.Format(time.RFC3339),
	}, nil
}

// TradeBotGetExecutedTrades gets list of trades executed in current session
func (s *Services) TradeBotGetExecutedTrades(ctx *gin.Context) (res []dtos.TradeData, err error) {
	tradeBotMutex.Lock()
	defer tradeBotMutex.Unlock()

	// Validate bot is active
	if !tradeBotActive || tradeBotTime == nil {
		return nil, fmt.Errorf("trade bot is not running. Please activate the bot first")
	}

	// Get trades from current session (using bot start time)
	trades, err := s.repo.Trade.FindTradesInSession(nil, *tradeBotTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed trades: %w", err)
	}

	// Convert to DTOs
	for _, trade := range trades {
		dto := s.convertTradeToDTO(trade)
		res = append(res, dto)
	}

	return res, nil
}

// TradeBotGetAll gets list of trades with optional filters
// All filter params are optional — omit to get all trades without that filter
func (s *Services) TradeBotGetAll(ctx *gin.Context, filter dtos.TradeFilter) (res []dtos.TradeData, err error) {
	trades, err := s.repo.Trade.FindWithFilter(nil, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	for _, trade := range trades {
		dto := s.convertTradeToDTO(trade)
		res = append(res, dto)
	}

	return res, nil
}

// convertTradeToDTO converts model.Trade to dtos.TradeData
func (s *Services) convertTradeToDTO(trade models.Trade) dtos.TradeData {
	orders := []dtos.OrderInfo{}

	for _, order := range trade.Entries {
		o := dtos.OrderInfo{
			EntryNumber:    order.EntryNumber,
			BinanceOrderID: order.BinanceOrderID,
			Price:          order.EntryPrice,
			Quantity:       order.PositionQty,
			Type:           order.EntryType,
			Status:         order.Status,
		}
		orders = append(orders, o)
	}

	return dtos.TradeData{
		ID:              trade.ID,
		Symbol:          trade.Symbol,
		Interval:        trade.Interval,
		Side:            trade.Side,
		Confidence:      trade.Confidence,
		TotalScore:      trade.TotalScore,
		IsAggressive:    trade.IsAggressive,
		TPPrice:         trade.TPPrice,
		SLPrice:         trade.SLPrice,
		RiskRewardRatio: trade.RiskRewardRatio,
		AvgEntryPrice:   trade.AvgEntryPrice,
		Leverage:        trade.Leverage,
		CapitalUsed:     trade.CapitalUsed,
		TotalQty:        trade.TotalQty,
		Status:          trade.Status,
		Description:     trade.Description,
		TPOrderID:       trade.TPOrderID,
		SLOrderID:       trade.SLOrderID,
		TPSLStatus:      trade.TPSLStatus,
		ExitReason:      trade.ExitReason,
		ExitPrice:       trade.ExitPrice,
		PnL:             trade.PnL,
		PnLPct:          trade.PnLPct,
		CreatedAt:       trade.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       trade.UpdatedAt.Format(time.RFC3339),
		ClosedAt:        trade.ClosedAt,
		Orders:          orders,
	}
}
