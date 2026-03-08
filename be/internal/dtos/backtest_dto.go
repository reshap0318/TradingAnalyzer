package dtos

import (
	"time"
)

// BacktestRequest represents request to create and run a backtest
type BacktestRequest struct {
	Name         string  `json:"name" binding:"required"`
	Symbol       string  `json:"symbol" binding:"required"`
	Interval     string  `json:"interval" binding:"required"`          // e.g., "15m", "1h"
	Days         int     `json:"days" binding:"required,min=1,max=30"` // Backtest duration in days
	Capital      float64 `json:"capital" binding:"required,min=30"`
	Leverage     int     `json:"leverage" binding:"required,min=1,max=125"`
	IsAggressive bool    `json:"is_aggressive"`
}

// BacktestResponse represents backtest result (WITH trades)
type BacktestResponse struct {
	ID              uint               `json:"id"`
	Name            string             `json:"name"`
	Symbol          string             `json:"symbol"`
	Interval        string             `json:"interval"`
	StartTime       time.Time          `json:"start_time"`
	EndTime         time.Time          `json:"end_time"`
	Capital         float64            `json:"capital"`
	Leverage        int                `json:"leverage"`
	IsAggressive    bool               `json:"is_aggressive"`
	TotalTrades     int                `json:"total_trades"`
	WinningTrades   int                `json:"winning_trades"`
	LosingTrades    int                `json:"losing_trades"`
	TotalPnL        float64            `json:"total_pnl"`
	TotalPnLPercent float64            `json:"total_pnl_percent"`
	MaxDrawdown     float64            `json:"max_drawdown"`
	WinRate         float64            `json:"win_rate"`
	ProfitFactor    float64            `json:"profit_factor"`
	Status          string             `json:"status"`
	ErrorMessage    string             `json:"error_message"`
	CreatedAt       time.Time          `json:"created_at"`
	CompletedAt     *time.Time         `json:"completed_at"`
	Trades          []BacktestTradeDTO `json:"trades"`
}

// BacktestTradeDTO represents individual trade in response
type BacktestTradeDTO struct {
	ID         uint       `json:"id"`
	EntryTime  time.Time  `json:"entry_time"`
	ExitTime   *time.Time `json:"exit_time"`
	Side       string     `json:"side"`
	EntryPrice float64    `json:"entry_price"`
	ExitPrice  float64    `json:"exit_price"`
	Quantity   float64    `json:"quantity"`
	PnL        float64    `json:"pnl"`
	PnLPercent float64    `json:"pnl_percent"`
	ExitReason string     `json:"exit_reason"`
	Status     string     `json:"status"`
}

// BacktestListItem represents backtest in list view (WITHOUT trades)
type BacktestListItem struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Symbol          string    `json:"symbol"`
	Interval        string    `json:"interval"`
	TotalPnL        float64   `json:"total_pnl"`
	TotalPnLPercent float64   `json:"total_pnl_percent"`
	WinRate         float64   `json:"win_rate"`
	TotalTrades     int       `json:"total_trades"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
