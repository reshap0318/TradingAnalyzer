package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
)

var (
	scannerActive   bool
	scannerCancel   context.CancelFunc
	scannerMutex    sync.Mutex
	scannerStrategy *uint
	scannerRunning  bool // Track if scan is currently running
)

// WatchlistScannerActivate activates the background scanner
func (s *Services) WatchlistScannerActivate(ctx *gin.Context, strategyID *uint) (res map[string]interface{}, err error) {
	scannerMutex.Lock()
	defer scannerMutex.Unlock()

	if scannerActive {
		return nil, fmt.Errorf("scanner is already active. Deactivate first before activating again")
	}

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

	// Create context with cancel
	ctxScan, cancel := context.WithCancel(context.Background())
	scannerCancel = cancel
	scannerActive = true
	scannerStrategy = &strategy.ID

	// Start background goroutine with dynamic interval
	go s.runBackgroundScanner(ctxScan, scanInterval)

	return map[string]interface{}{
		"is_active":     true,
		"message":       "Scanner activated successfully",
		"scan_interval": scanInterval.Minutes(),
		"strategy_id":   strategy.ID,
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
func (s *Services) runBackgroundScanner(ctx context.Context, scanInterval time.Duration) {
	// Add panic recovery for goroutine
	defer func() {
		if r := recover(); r != nil {
			s.logScannerError(fmt.Sprintf("🚨 PANIC in scanner goroutine: %v", r))
		}
	}()

	s.logScannerInfo("╔══════════════════════════════════════════════════════════╗")
	s.logScannerInfo("║     WATCHLIST SCANNER STARTING                           ║")
	s.logScannerInfo(fmt.Sprintf("║     Interval: %.0f minutes                                    ║", scanInterval.Minutes()))
	s.logScannerInfo("╚══════════════════════════════════════════════════════════╝")

	// ✅ RUN SCAN PERTAMA LANGSUNG
	s.logScannerInfo("Running initial scan immediately...")
	s.runScanCycle()

	// ✅ Calculate time until next interval mark (e.g., 00:15, 00:30, 00:45)
	initialDelay := s.calculateNextIntervalDelay(scanInterval)
	if initialDelay > 0 {
		nextExecution := time.Now().Add(initialDelay)
		s.logScannerInfo(fmt.Sprintf("⏳ Initial wait: %.0f minutes (next scan at %s)", 
			initialDelay.Minutes(), nextExecution.Format("15:04:05")))

		// Use a ticker for the initial delay to allow cancellation
		delayTicker := time.NewTicker(initialDelay)
		select {
		case <-ctx.Done():
			delayTicker.Stop()
			s.logScannerInfo("Scanner stopped during initial delay")
			return
		case <-delayTicker.C:
			delayTicker.Stop()
			s.logScannerInfo(fmt.Sprintf("⏰ Initial wait complete, running scan at %s", time.Now().Format("15:04:05")))
			// ✅ RUN SCAN immediately after initial wait
			s.runScanCycle()
		}
	}

	// ✅ Start periodic ticker for subsequent scans
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logScannerInfo("🛑 Scanner stopped by user request")
			return
		case <-ticker.C:
			s.logScannerInfo(fmt.Sprintf("🔔 Ticker triggered at %s", time.Now().Format("15:04:05")))
			
			// Option 1: Skip if scan is already running
			scannerMutex.Lock()
			if scannerRunning {
				scannerMutex.Unlock()
				s.logScannerSkip("⏰ Previous scan still running - skipping this interval")
				continue
			}
			scannerRunning = true
			scannerMutex.Unlock()

			// Run scan (scannerRunning will be reset by runScanCycle's defer)
			s.runScanCycle()
		}
	}
}

// calculateNextIntervalDelay calculates duration until the next interval mark
// Example: if interval=15m and current time=00:05, returns 10m (until 00:15)
func (s *Services) calculateNextIntervalDelay(interval time.Duration) time.Duration {
	now := time.Now()
	intervalMinutes := int(interval.Minutes())

	// Get current minute of hour
	currentMinute := now.Minute()

	// Calculate next interval mark
	nextMark := ((currentMinute / intervalMinutes) + 1) * intervalMinutes

	// Handle overflow (e.g., 00:55 + 15m = 01:00, not 01:10)
	if nextMark >= 60 {
		nextMark = 0 // Next hour's 00 minute
	}

	// Calculate target time
	target := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), nextMark, 0, 0, now.Location())

	// If we calculated a time in the past (edge case at :00), add an hour
	if target.Before(now) {
		target = target.Add(1 * time.Hour)
	}

	// Return duration until target
	return target.Sub(now)
}

// runScanCycle runs a single scan cycle
func (s *Services) runScanCycle() {
	cycleStart := time.Now()
	
	s.logScannerInfo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. Get all active symbols from watchlist
	watchlists, err := s.repo.Watchlist.FindAllActive(nil)
	if err != nil {
		s.logScannerError(fmt.Sprintf("Failed to get watchlist: %v", err))
		return
	}

	if len(watchlists) == 0 {
		s.logScannerInfo("⚠️  No active symbols in watchlist")
		return
	}

	s.logScannerSuccess(fmt.Sprintf("Found %d active symbols in watchlist", len(watchlists)))
	
	// Log cycle start with actual count
	s.logScannerCycle("start", len(watchlists), 0)
	s.logScannerInfo(fmt.Sprintf("📋 Scan list: %v", extractSymbolsFromWatchlist(watchlists)))

	// Track trade stats
	tradesExecuted := 0
	tradesSkipped := 0

	// Ensure scannerRunning is reset when scan completes (panic recovery)
	defer func() {
		duration := time.Since(cycleStart)
		s.logScannerCycle("complete", len(watchlists), duration)
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
			s.logScannerInfo("🚫 Scanner deactivated during scan cycle")
			return
		}
		scannerMutex.Unlock()

		s.logScannerInfo(fmt.Sprintf("🔍 Scanning: %s", wl.Symbol))

		// 3. Call TradeExecute for each symbol
		tradeReq := &dtos.TradeRequest{
			Symbol:     wl.Symbol,
			StrategyID: 0, // Will use strategy from scanner if provided
		}

		// Create mock gin context for TradeExecute
		ginCtx := &gin.Context{}

		// Execute trade
		tradeRes, err := s.TradeExecute(ginCtx, tradeReq)

		if err != nil {
			s.logScannerError(fmt.Sprintf("Trade failed for %s: %v", wl.Symbol, err))
			tradesSkipped++
		} else if tradeRes != nil && tradeRes.ExecutionInfo.Executed {
			s.logScannerTrade(true, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesExecuted++
		} else {
			s.logScannerTrade(false, wl.Symbol, tradeRes.ExecutionInfo.Message)
			tradesSkipped++
		}

		// Small delay between symbols to avoid rate limiting
		if len(watchlists) > 1 {
			time.Sleep(30 * time.Second)
		}
	}

	// Summary
	s.logScannerInfo("───────────────────────────────────────────────────────────")
	s.logScannerSuccess(fmt.Sprintf("Cycle Summary: %d executed, %d skipped", tradesExecuted, tradesSkipped))
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

// logScannerInfo logs info messages to file
func (s *Services) logScannerInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] 🟢 [SCANNER] [INFO] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
}

// logScannerError logs error messages to file
func (s *Services) logScannerError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] 🔴 [SCANNER] [ERROR] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
}

// logScannerSuccess logs success messages with checkmark
func (s *Services) logScannerSuccess(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] ✅ [SCANNER] [SUCCESS] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
}

// logScannerSkip logs skip messages with warning icon
func (s *Services) logScannerSkip(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] ⏭️ [SCANNER] [SKIP] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
}

// logScannerTrade logs trade execution messages
func (s *Services) logScannerTrade(executed bool, symbol, message string) {
	icon := "⏭️"
	status := "NO TRADE"
	if executed {
		icon = "🚀"
		status = "EXECUTED"
	}
	logLine := fmt.Sprintf("[%s] %s [SCANNER] [%s] %s → %s\n",
		time.Now().Format("2006-01-02 15:04:05"), icon, status, symbol, message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine)
}

// logScannerCycle logs scan cycle start and completion
func (s *Services) logScannerCycle(event string, symbolCount int, duration time.Duration) {
	var icon string
	switch event {
	case "start":
		icon = "🔄"
		logLine := fmt.Sprintf("[%s] %s [SCANNER] [CYCLE START] Beginning scan of %d symbols...\n",
			time.Now().Format("2006-01-02 15:04:05"), icon, symbolCount)
		s.writeScannerLog(logLine)
		fmt.Println(logLine)
	case "complete":
		icon = "✨"
		logLine := fmt.Sprintf("[%s] %s [SCANNER] [CYCLE COMPLETE] Scanned %d symbols in %.1fs\n",
			time.Now().Format("2006-01-02 15:04:05"), icon, symbolCount, duration.Seconds())
		s.writeScannerLog(logLine)
		fmt.Println(logLine)
	}
}

// writeScannerLog writes log messages to file
func (s *Services) writeScannerLog(message string) {
	// Create logs directory if not exists
	logsDir := "./logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		return
	}

	// Log file path
	logFile := filepath.Join(logsDir, "watchlist_scanner.log")

	// Append to log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(message); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}
