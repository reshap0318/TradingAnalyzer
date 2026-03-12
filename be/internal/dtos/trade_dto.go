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

// ============== TRADE MONITOR DTOs ==============

// ProcessTradeResult holds the result of processing a trade
type ProcessTradeResult struct {
	TradeID      uint   `json:"trade_id"`
	Symbol       string `json:"symbol"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	EntriesSync  int    `json:"entries_sync"`
	TPUpdated    bool   `json:"tp_updated"`
	SLUpdated    bool   `json:"sl_updated"`
	UpdatedCount int    `json:"updated_count"`
}

// TradeMonitorRequest represents the request to process a single trade
type TradeMonitorRequest struct {
	TradeID uint `json:"trade_id" binding:"required"`
}
