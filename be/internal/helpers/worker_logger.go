package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// WorkerLogger provides structured logging for background workers.
// Supports multiple worker types (SCANNER, TPSL, etc.) with timestamp-based file naming.
// Logs go to file by default; only start/stop banners and cycle summaries print to console.
type WorkerLogger struct {
	WorkerName string   // Display name: "SCANNER", "TPSL"
	filePrefix string   // File prefix: "watcher", "tpsl"
	StartedAt  time.Time // Session start timestamp
	file       *os.File // Persistent file handle
	mu         sync.Mutex
}

const (
	logTimeFormat    = "2006-01-02 15:04:05"
	logsDir          = "./logs"
	logRetentionDays = 30 // Auto-delete logs older than this
)

// NewWorkerLogger creates a new logger instance and opens the log file.
// File naming: ./logs/{filePrefix}_{YYYY-MM-DD_HH-MM-SS}.log
// Example: ./logs/executor_2026-03-16_01-15-00.log
func NewWorkerLogger(workerName, filePrefix string) (*WorkerLogger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Use timestamp for unique session file
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logsDir, fmt.Sprintf("%s_%s.log", filePrefix, timestamp))

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}

	return &WorkerLogger{
		WorkerName: workerName,
		filePrefix: filePrefix,
		StartedAt:  now,
		file:       f,
	}, nil
}

// CleanupOldLogs removes log files older than logRetentionDays.
// Called once on application startup or bot activation.
func CleanupOldLogs() {
	cutoff := time.Now().AddDate(0, 0, -logRetentionDays)
	
	// Pattern: {prefix}_{YYYY-MM-DD_HH-MM-SS}.log
	pattern := regexp.MustCompile(`^[a-z]+_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.log$`)
	
	deletedCount := 0
	
	filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Only process files matching our pattern
		filename := info.Name()
		if !pattern.MatchString(filename) {
			return nil
		}
		
		// Delete if older than cutoff
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: Failed to delete old log %s: %v\n", path, err)
				return nil
			}
			deletedCount++
		}
		
		return nil
	})
	
	if deletedCount > 0 {
		fmt.Printf("🧹 Cleaned up %d old log files (older than %d days)\n", deletedCount, logRetentionDays)
	}
}

// getLogDiskUsage returns total size of log files in MB.
func getLogDiskUsage() float64 {
	var totalSize int64
	
	filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			totalSize += info.Size()
		}
		
		return nil
	})
	
	return float64(totalSize) / (1024 * 1024) // Convert to MB
}

// Close closes the underlying log file handle.
func (l *WorkerLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

// ─── Core Logging Methods ──────────────────────────────────────────────────────
// These write to FILE ONLY. Use Console() or Banner() for stdout output.

// Info logs an informational message (file only).
func (l *WorkerLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] 🔹 [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
}

// Error logs an error message (file only).
func (l *WorkerLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] ❌ [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
}

// Success logs a success message (file only).
func (l *WorkerLogger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] ✅ [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
}

// Warn logs a warning message (file only).
func (l *WorkerLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] ⚠️  [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
}

// Skip logs a skipped action (file only).
func (l *WorkerLogger) Skip(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] 🔸 [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
}

// ─── Trade Logging ─────────────────────────────────────────────────────────────

// Trade logs a trade execution result (file only).
func (l *WorkerLogger) Trade(executed bool, symbol, message string) {
	icon := "🔸"
	status := "NO TRADE"
	if executed {
		icon = "🚀"
		status = "EXECUTED"
	}
	line := fmt.Sprintf("[%s] %s [%s] [%s] %s → %s\n",
		time.Now().Format(logTimeFormat), icon, l.WorkerName, status, symbol, message)
	l.writeToFile(line)
}

// ─── Cycle Logging ─────────────────────────────────────────────────────────────

// CycleStart logs the beginning of a scan cycle (file only).
func (l *WorkerLogger) CycleStart(symbolCount int) {
	line := fmt.Sprintf("[%s] 🔄 [%s] Cycle started — scanning %d symbols\n",
		time.Now().Format(logTimeFormat), l.WorkerName, symbolCount)
	l.writeToFile(line)
}

// CycleSummary logs cycle completion stats (file + console).
func (l *WorkerLogger) CycleSummary(symbolCount, executed, skipped int, duration time.Duration) {
	now := time.Now().Format(logTimeFormat)

	// File log — clean box format
	separator := "───────────────────────────────────────────────────"
	fileOutput := separator + "\n"
	fileOutput += fmt.Sprintf("  ✅ [%s] Cycle Complete\n", l.WorkerName)
	fileOutput += fmt.Sprintf("  📊 %d symbols | %d executed | %d skipped\n", symbolCount, executed, skipped)
	fileOutput += fmt.Sprintf("  ⏱  Duration: %.1fs | Finished: %s\n", duration.Seconds(), now)
	fileOutput += separator + "\n"
	l.writeToFile(fileOutput)

	// Console — banner style matching scanner start
	border := "═══════════════════════════════════════════════════"
	consoleOutput := border + "\n"
	consoleOutput += fmt.Sprintf("  ✅ [%s] Cycle Complete — %s\n", l.WorkerName, now)
	consoleOutput += fmt.Sprintf("  📊 %d symbols | %d exec | %d skip | ⏱ %.1fs\n", symbolCount, executed, skipped, duration.Seconds())
	consoleOutput += border + "\n"
	fmt.Print(consoleOutput)
}

// ─── Banner / Console ──────────────────────────────────────────────────────────

// Banner prints a decorated box banner to both file and console.
// Use for start/stop events that the user should see in terminal.
func (l *WorkerLogger) Banner(bannerLines ...string) {
	border := "═══════════════════════════════════════════════════"
	output := border + "\n"
	for _, line := range bannerLines {
		output += "  " + line + "\n"
	}
	output += border + "\n"

	l.writeToFile(output)
	fmt.Print(output)
}

// Console logs a message to both file and console.
// Use sparingly — only for critical information the user must see in terminal.
func (l *WorkerLogger) Console(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] 🔹 [%s] %s\n", time.Now().Format(logTimeFormat), l.WorkerName, msg)
	l.writeToFile(line)
	fmt.Print(line)
}

// ─── Internal ──────────────────────────────────────────────────────────────────

// writeToFile writes a log line to the file handle.
func (l *WorkerLogger) writeToFile(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		// File was closed or never opened — fallback to console
		fmt.Print(message)
		return
	}

	if _, err := l.file.WriteString(message); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
	}
}
