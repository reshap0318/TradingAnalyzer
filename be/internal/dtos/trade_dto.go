package dtos

import "time"

// ============== REQUEST DTOs ==============

// TradeRequest represents the request for auto-trading execution
type TradeRequest struct {
	Symbol     string `json:"symbol" binding:"required"`
	StrategyID uint   `json:"strategy_id"` // Optional: if not provided, uses active strategy
}

// ============== RESPONSE DTOs ==============

// TradeResponse represents the result of the auto-trade execution
type TradeResponse struct {
	Symbol           string           `json:"symbol"`
	PrimaryTimeframe string           `json:"primary_timeframe"`
	Timestamp        time.Time        `json:"timestamp"`
	Signal           SignalInfo       `json:"signal,omitempty"`
	Scoring          ScoringBreakdown `json:"scoring,omitempty"`
	ExecutionInfo    ExecutionInfo    `json:"execution_info,omitempty"`
}

// ExecutionInfo contains details about the execution status and orders placed
type ExecutionInfo struct {
	Executed    bool        `json:"executed"`
	Message     string      `json:"message"`
	MarginType  string      `json:"margin_type,omitempty"`
	Leverage    int         `json:"leverage,omitempty"`
	CapitalUsed float64     `json:"capital_used,omitempty"`
	Orders      []OrderInfo `json:"orders,omitempty"`
	TPOrderID   int64       `json:"tp_order_id,omitempty"`
	SLOrderID   int64       `json:"sl_order_id,omitempty"`
}

// OrderInfo contains minimal info about an executed order
type OrderInfo struct {
	EntryNumber    int     `json:"entry_number"`
	BinanceOrderID int64   `json:"binance_order_id"`
	Price          float64 `json:"price"`
	Quantity       float64 `json:"quantity"`
	Type           string  `json:"type"`   // MARKET or LIMIT
	Status         string  `json:"status"` // FILLED, NEW, dll
}

type TradeDayStat struct {
	Active             int8
	Count              int8
	TPHits             int8
	SLHits             int8
	TotalLoss          float64
	TotalProfit        float64
	ConsecutiveLossess int8
	PnL                float64
}

// ============== TRADE FILTER DTO ==============

// TradeFilter represents optional query filters for TradeBotGet endpoint
// All fields are optional — omitting a field means no filter is applied for that field
type TradeFilter struct {
	Status    []string `form:"status"`         // Filter by status (e.g. ACTIVE, CLOSED, CANCELLED) - accepts multiple
	Symbol    []string `form:"symbol"`         // Filter by symbol (e.g. BTCUSDT, ETHUSDT) - accepts multiple
	Interval  string   `form:"interval"`       // Filter by timeframe interval (e.g. 15m, 1h, 4h)
	MinConf   float64  `form:"min_confidence"` // Filter trades with confidence >= this value
	Side      string   `form:"side"`           // Filter by side: BUY or SELL
	DateStart string   `form:"date_start"`     // Filter trades created on or after this date (YYYY-MM-DD)
	DateEnd   string   `form:"date_end"`       // Filter trades created on or before this date (YYYY-MM-DD)
}

// ============== TRADE MONITOR DTOs ==============

// ProcessTradeResult holds the result of processing a trade
type ProcessTradeResult struct {
	TradeID      uint     `json:"trade_id"`
	Symbol       string   `json:"symbol"`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	EntriesSync  int      `json:"entries_sync"`
	TPUpdated    bool     `json:"tp_updated"`
	SLUpdated    bool     `json:"sl_updated"`
	UpdatedCount int      `json:"updated_count"`
	Logs         []string `json:"logs"` // Detailed execution flow logs
}

// TradeMonitorRequest represents the request to process a single trade
type TradeMonitorRequest struct {
	TradeID uint `json:"trade_id" binding:"required"`
}

// ============== TRADE DATA DTO ==============

// TradeData represents trade data in responses
type TradeData struct {
	ID              uint        `json:"id"`
	Symbol          string      `json:"symbol"`
	Interval        string      `json:"interval"`
	Side            string      `json:"side"`
	Confidence      float64     `json:"confidence"`
	TotalScore      float64     `json:"total_score"`
	IsAggressive    bool        `json:"is_aggressive"`
	TPPrice         float64     `json:"tp_price"`
	SLPrice         float64     `json:"sl_price"`
	RiskRewardRatio float64     `json:"risk_reward_ratio"`
	AvgEntryPrice   float64     `json:"avg_entry_price"`
	Leverage        int         `json:"leverage"`
	CapitalUsed     float64     `json:"capital_used"`
	TotalQty        float64     `json:"total_qty"`
	Status          string      `json:"status"`
	Description     string      `json:"description"`
	TPOrderID       int64       `json:"tp_order_id"`
	SLOrderID       int64       `json:"sl_order_id"`
	TPSLStatus      string      `json:"tp_sl_status"`
	ExitReason      string      `json:"exit_reason"`
	ExitPrice       float64     `json:"exit_price"`
	PnL             float64     `json:"pnl"`
	PnLPct          float64     `json:"pnl_pct"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
	ClosedAt        *time.Time  `json:"closed_at"`
	Orders          []OrderInfo `json:"orders,omitempty"`
}
