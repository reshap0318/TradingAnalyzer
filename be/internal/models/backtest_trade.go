package models

import (
	"time"
)

// BacktestTrade represents individual trades in a backtest
type BacktestTrade struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	BacktestID      uint       `gorm:"column:backtest_id;not null;index" json:"backtest_id"`
	EntryTime       time.Time  `gorm:"column:entry_time;not null" json:"entry_time"`
	FilledTime      *time.Time `gorm:"column:filled_time" json:"filled_time"`
	ExitTime        *time.Time `gorm:"column:exit_time" json:"exit_time"`
	Side            string     `gorm:"column:side;size:10;not null" json:"side"`
	EntryPrice      float64    `gorm:"column:entry_price;type:decimal(15,8);not null" json:"entry_price"`
	ExitPrice       float64    `gorm:"column:exit_price;type:decimal(15,8)" json:"exit_price"`
	Quantity        float64    `gorm:"column:quantity;type:decimal(20,8);not null" json:"quantity"`
	PnL             float64    `gorm:"column:pnl;type:decimal(20,2)" json:"pnl"`
	PnLPercent      float64    `gorm:"column:pnl_percent;type:decimal(10,2)" json:"pnl_percent"`
	TakeProfit      float64    `gorm:"column:take_profit;type:decimal(15,8)" json:"take_profit"`
	StopLoss        float64    `gorm:"column:stop_loss;type:decimal(15,8)" json:"stop_loss"`
	ExitReason      string     `gorm:"column:exit_reason;size:50" json:"exit_reason"`
	Status          string     `gorm:"column:status;size:20;default:'OPEN'" json:"status"`
	DurationMinutes int64      `gorm:"column:duration_minutes;default:0" json:"duration_minutes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (BacktestTrade) TableName() string {
	return "backtest_trade"
}
