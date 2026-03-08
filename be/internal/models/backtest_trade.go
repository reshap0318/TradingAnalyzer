package models

import (
	"time"
)

// BacktestTrade represents individual trades in a backtest
type BacktestTrade struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	BacktestID uint       `gorm:"not null;index" json:"backtest_id"`
	EntryTime  time.Time  `gorm:"not null" json:"entry_time"`
	ExitTime   *time.Time `json:"exit_time"`
	Side       string     `gorm:"size:10;not null" json:"side"`
	EntryPrice float64    `gorm:"type:decimal(15,8);not null" json:"entry_price"`
	ExitPrice  float64    `gorm:"type:decimal(15,8)" json:"exit_price"`
	Quantity   float64    `gorm:"type:decimal(20,8);not null" json:"quantity"`
	PnL        float64    `gorm:"type:decimal(20,2)" json:"pnl"`
	PnLPercent float64    `gorm:"type:decimal(10,2)" json:"pnl_percent"`
	TakeProfit float64    `gorm:"type:decimal(15,8)" json:"take_profit"`
	StopLoss   float64    `gorm:"type:decimal(15,8)" json:"stop_loss"`
	ExitReason string     `gorm:"size:50" json:"exit_reason"`
	Status     string     `gorm:"size:20;default:'OPEN'" json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (BacktestTrade) TableName() string {
	return "backtest_trade"
}
