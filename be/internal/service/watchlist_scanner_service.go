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
	scannerActive   bool
	scannerCancel   context.CancelFunc
	scannerMutex    sync.Mutex
	scannerStrategy *uint
	scannerRunning  bool // Track if scan is currently running
	scannerLogger   *helpers.WorkerLogger

	// Shared log counter across all workers (scanner, tpsl, etc.)
	// Reset when service restarts
	workerLogCounter      int
	workerLogCounterMutex sync.Mutex
)

// getNextLogNumber returns the next shared session counter for all workers
func getNextLogNumber() int {
	workerLogCounterMutex.Lock()
	defer workerLogCounterMutex.Unlock()
	workerLogCounter++
	return workerLogCounter
}

// WatchlistScannerActivate activates the background scanner
func (s *Services) WatchlistScannerActivate(ctx *gin.Context, strategyID *uint) (res map[string]interface{}, err error) {
	scannerMutex.Lock()
	defer scannerMutex.Unlock()

	if scannerActive {
		return nil, fmt.Errorf("scanner is already active. Deactivate first before activating again")
	}

	// Get shared session counter
	currentLogNumber := getNextLogNumber()

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

	// Use timeframe in_minutes as scan interval
	scanInterval := time.Duration(timeframe.InMinutes) * time.Minute

	// Create logger for this scanner session
	scannerLogger, err = helpers.NewWorkerLogger("SCANNER", "watcher", currentLogNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner logger: %w", err)
	}

	// Create context with cancel
	ctxScan, cancel := context.WithCancel(context.Background())
	scannerCancel = cancel
	scannerActive = true
	scannerStrategy = &strategy.ID

	// Start background goroutine with dynamic interval
	go s.runBackgroundScanner(ctxScan, scanInterval, scannerLogger)

	return map[string]interface{}{
		"is_active":     true,
		"message":       "Scanner activated successfully",
		"scan_interval": scanInterval.Minutes(),
		"strategy_id":   strategy.ID,
		"log_number":    currentLogNumber,
	}, nil
}

// WatchlistScannerDeactivate deactivates the background scanner
func (s *Services) WatchlistScannerDeactivate(ctx *gin.Context) (res map[string]interface{}, err error) {
	scannerMutex.Lock()
	defer scannerMutex.Unlock()

	if !scannerActive {
		return nil, fmt.Errorf("scanner is not active")
	}

	// Cancel the context
	if scannerCancel != nil {
		scannerCancel()
	}
	scannerActive = false
	scannerRunning = false // Reset running flag

	// Close logger file handle
	if scannerLogger != nil {
		scannerLogger.Banner("🔴 WATCHLIST SCANNER STOPPED",
			fmt.Sprintf("⏱  Stopped at: %s", time.Now().Format("2006-01-02 15:04:05")))
		scannerLogger.Close()
		scannerLogger = nil
	}

	return map[string]interface{}{
		"is_active": false,
		"message":   "Scanner deactivated successfully",
	}, nil
}

// WatchlistScannerGetStatus gets the current scanner status
func (s *Services) WatchlistScannerGetStatus(ctx *gin.Context) (res map[string]interface{}, err error) {
	scannerMutex.Lock()
	defer scannerMutex.Unlock()

	return map[string]interface{}{
		"is_active": scannerActive,
	}, nil
}

// runBackgroundScanner runs the background scanning process
func (s *Services) runBackgroundScanner(ctx context.Context, scanInterval time.Duration, logger *helpers.WorkerLogger) {
	// Add panic recovery for goroutine
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC in scanner goroutine: %v", r)
		}
	}()

	// Print start banner (console + file)
	logger.Banner(
		fmt.Sprintf("🟢 WATCHLIST SCANNER STARTED — Session #%03d", logger.LogNumber),
		fmt.Sprintf("⏱  Interval: %.0fm | Strategy: %d", scanInterval.Minutes(), *scannerStrategy),
	)

	// ✅ RUN SCAN PERTAMA LANGSUNG
	logger.Info("Running initial scan immediately...")
	s.runScanCycle(logger)

	// ✅ Loop: always align to next clock mark (e.g., :00, :15, :30, :45)
	for {
		// Calculate delay until next interval mark
		delay := s.calculateNextIntervalDelay(scanInterval)
		nextExecution := time.Now().Add(delay)
		logger.Info("Next scan at %s (waiting %.0fm)",
			nextExecution.Format("2006-01-02 15:04:05"), delay.Minutes())

		// Wait until next mark, with cancellation support
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			logger.Info("Scanner stopped by user request")
			return
		case <-timer.C:
			logger.Info("Scan triggered at %s", time.Now().Format("2006-01-02 15:04:05"))

			// Skip if scan is already running
			scannerMutex.Lock()
			if scannerRunning {
				scannerMutex.Unlock()
				logger.Skip("Previous scan still running - skipping this interval")
				continue
			}
			scannerRunning = true
			scannerMutex.Unlock()

			// Run scan (scannerRunning will be reset by runScanCycle's defer)
			s.runScanCycle(logger)
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

// runScanCycle runs a single scan cycle
func (s *Services) runScanCycle(logger *helpers.WorkerLogger) {
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
	logger.Info("Scan list: %s", extractSymbolsFromWatchlist(watchlists))

	// Track trade stats
	tradesExecuted := 0
	tradesSkipped := 0

	// Ensure scannerRunning is reset when scan completes (panic recovery)
	defer func() {
		duration := time.Since(cycleStart)
		logger.CycleSummary(len(watchlists), tradesExecuted, tradesSkipped, duration)
		scannerMutex.Lock()
		scannerRunning = false
		scannerMutex.Unlock()
	}()

	// 2. Scan each symbol
	for _, wl := range watchlists {
		// Check if scanner is still active
		scannerMutex.Lock()
		if !scannerActive {
			scannerMutex.Unlock()
			logger.Warn("Scanner deactivated during scan cycle")
			return
		}
		scannerMutex.Unlock()

		logger.Info("Scanning: %s", wl.Symbol)

		// 3. Call TradeExecute for each symbol
		tradeReq := &dtos.TradeRequest{
			Symbol:     wl.Symbol,
			StrategyID: *scannerStrategy,
		}

		// Create mock gin context for TradeExecute
		ginCtx := &gin.Context{}

		// Execute trade
		tradeRes, err := s.TradeExecute(ginCtx, tradeReq)

		if err != nil {
			logger.Error("Trade failed for %s: %v", wl.Symbol, err)
			tradesSkipped++
		} else if tradeRes != nil && tradeRes.ExecutionInfo.Executed {
			logger.Trade(true, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesExecuted++
		} else {
			logger.Trade(false, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesSkipped++
		}

		// Small delay between symbols to avoid rate limiting
		if len(watchlists) > 1 {
			time.Sleep(30 * time.Second)
		}
	}
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
