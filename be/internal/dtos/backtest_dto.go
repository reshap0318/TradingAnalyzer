package dtos

import (
	"time"
)

// BacktestRequest represents request to create and run a backtest
type BacktestRequest struct {
	Name       string  `json:"name" binding:"required"`
	Symbol     string  `json:"symbol" binding:"required"`
	StrategyID uint    `json:"strategy_id" binding:"required,min=1"`
	Days       int     `json:"days" binding:"required,min=1,max=30"`
	Capital    float64 `json:"capital" binding:"required,min=10"` // Default 1000 USDT
}

// BacktestResponse represents backtest result with equity curve, OHLCV data and detailed trades
type BacktestResponse struct {
	ID           uint               `json:"id"`
	Name         string             `json:"name"`
	Symbol       string             `json:"symbol"`
	StrategyID   uint               `json:"strategy_id"`
	StartTime    time.Time          `json:"start_time"`
	EndTime      time.Time          `json:"end_time"`
	Capital      float64            `json:"capital"`
	Summary      BacktestSummary    `json:"summary"`
	EquityCurve  []EquityPoint      `json:"equity_curve"`
	Trades       []BacktestTradeDTO `json:"trades"`
	Status       string             `json:"status"`
	ErrorMessage string             `json:"error_message,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	CompletedAt  *time.Time         `json:"completed_at"`
	Strategy     *StrategyData      `json:"strategy,omitempty"`
	OHLCV        []CandleData       `json:"ohlcv"` // OHLCV data for charting
}

// CandleData represents OHLCV candle data for charting
type CandleData struct {
	Timestamp int64   `json:"timestamp"` // Unix timestamp in milliseconds
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// BacktestSummary represents summary statistics of backtest
type BacktestSummary struct {
	InitialBalance   float64 `json:"initial_balance"`
	FinalBalance     float64 `json:"final_balance"`
	NetProfit        float64 `json:"net_profit"`
	NetProfitPercent float64 `json:"net_profit_percent"`
	WinRate          float64 `json:"win_rate_pct"`
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
	ExpiredTrades    int     `json:"expired_trades"`
	CancelledTrades  int     `json:"cancelled_trades"`
	MaxDrawdown      float64 `json:"max_drawdown_pct"`
	ProfitFactor     float64 `json:"profit_factor"`
	AvgWin           float64 `json:"avg_win"`
	AvgLoss          float64 `json:"avg_loss"`
	LargestWin       float64 `json:"largest_win"`
	LargestLoss      float64 `json:"largest_loss"`
}

// EquityPoint represents a single point in the equity curve
type EquityPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix timestamp in milliseconds
	Balance   float64 `json:"balance"`
	PnL       float64 `json:"pnl"`
}

// BacktestTradeDTO represents individual trade with detailed entries and exits
type BacktestTradeDTO struct {
	TradeID     uint    `json:"trade_id"`
	TradeNum    int     `json:"trade_num"`
	Side        string  `json:"side"`
	Signal      string  `json:"signal"`
	Confidence  float64 `json:"confidence"`
	TradingMode string  `json:"trading_mode"`
	Status      string  `json:"status"`

	// Targets from trading plan
	Targets TradeTargets `json:"targets"`

	// Entries (can be multiple for multi-entry strategy)
	Entries []TradeEntry `json:"entries"`

	// Exit information
	Exit *TradeExit `json:"exit,omitempty"`

	// Aggregated stats
	TotalQty      float64 `json:"total_qty"`
	AvgEntryPrice float64 `json:"avg_entry_price"`
	TotalCapital  float64 `json:"total_capital"`

	// PnL
	PnL        float64 `json:"pnl"`
	PnLPercent float64 `json:"pnl_percent"`

	// Timing
	EntryTime       time.Time  `json:"entry_time"`
	FilledTime      *time.Time `json:"filled_time,omitempty"`
	ExitTime        *time.Time `json:"exit_time,omitempty"`
	DurationMinutes int64      `json:"duration_minutes,omitempty"`

	// Daily stats snapshot
	DailyStats *DailyStatsSnapshot `json:"daily_stats,omitempty"`
}

// TradeTargets represents TP and SL prices
type TradeTargets struct {
	TPPrice float64 `json:"tp_price"`
	SLPrice float64 `json:"sl_price"`
	Ratio   float64 `json:"ratio"`
}

// TradeEntry represents a single entry in a multi-entry trade
type TradeEntry struct {
	EntryNum  int       `json:"entry_num"`
	Type      string    `json:"type"` // MARKET or LIMIT
	Price     float64   `json:"price"`
	Qty       float64   `json:"qty"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // PENDING, FILLED, CANCELLED, EXPIRED
}

// TradeExit represents exit information
type TradeExit struct {
	Reason    string    `json:"reason"` // HIT_TP, HIT_SL, CLOSED_END, DEAD_SIGNAL, EXPIRED
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// DailyStatsSnapshot represents daily statistics at trade entry
type DailyStatsSnapshot struct {
	TradeCount      int     `json:"trade_count"`
	PnL             float64 `json:"pnl"`
	ConsecutiveLoss int     `json:"consecutive_loss"`
}

// BacktestListItem represents backtest in list view (WITHOUT trades)
type BacktestListItem struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Symbol          string    `json:"symbol"`
	StrategyName    string    `json:"strategy_name"`
	TotalPnL        float64   `json:"total_pnl"`
	TotalPnLPercent float64   `json:"total_pnl_percent"`
	WinRate         float64   `json:"win_rate"`
	TotalTrades     int       `json:"total_trades"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
