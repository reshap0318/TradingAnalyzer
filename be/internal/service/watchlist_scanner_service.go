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
)

var (
	scannerActive   bool
	scannerCancel   context.CancelFunc
	scannerMutex    sync.Mutex
	scannerStrategy *uint
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
	s.logScannerInfo(fmt.Sprintf("Background scanner started (interval: %.0f minutes)", scanInterval.Minutes()))

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logScannerInfo("Scanner stopped")
			return
		case <-ticker.C:
			// Run scan
			s.runScanCycle()
		}
	}
}

// runScanCycle runs a single scan cycle
func (s *Services) runScanCycle() {
	s.logScannerInfo("Starting scan cycle...")

	// 1. Get all active symbols from watchlist
	watchlists, err := s.repo.Watchlist.FindAllActive(nil)
	if err != nil {
		s.logScannerError(fmt.Sprintf("Failed to get watchlist: %v", err))
		return
	}

	if len(watchlists) == 0 {
		s.logScannerInfo("No active symbols in watchlist")
		return
	}

	s.logScannerInfo(fmt.Sprintf("Found %d active symbols in watchlist", len(watchlists)))

	// 2. Scan each symbol
	for _, wl := range watchlists {
		// Check if scanner is still active
		scannerMutex.Lock()
		if !scannerActive {
			scannerMutex.Unlock()
			s.logScannerInfo("Scanner deactivated during scan cycle")
			return
		}
		scannerMutex.Unlock()

		s.logScannerInfo(fmt.Sprintf("Scanning symbol: %s", wl.Symbol))

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
		} else if tradeRes != nil && tradeRes.ExecutionInfo.Executed {
			s.logScannerInfo(fmt.Sprintf("✅ Trade executed for %s: %s", wl.Symbol, tradeRes.ExecutionInfo.Message))
		} else {
			s.logScannerInfo(fmt.Sprintf("⏭️ No trade for %s: %s", wl.Symbol, tradeRes.ExecutionInfo.Message))
		}

		// Small delay between symbols to avoid rate limiting
		time.Sleep(30 * time.Second)
	}

	s.logScannerInfo("Scan cycle completed")
}

// logScannerInfo logs info messages to file
func (s *Services) logScannerInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [SCANNER] [INFO] %s\n", time.Now().Format(time.RFC3339), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
}

// logScannerError logs error messages to file
func (s *Services) logScannerError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [SCANNER] [ERROR] %s\n", time.Now().Format(time.RFC3339), message)
	s.writeScannerLog(logLine)
	fmt.Println(logLine) // Also print to console
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
