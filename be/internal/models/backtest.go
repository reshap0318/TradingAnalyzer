package models

import (
	"time"
)

// Backtest represents a backtest run
type Backtest struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:100;not null" json:"name"`
	Symbol          string     `gorm:"size:20;not null" json:"symbol"`
	StrategyID      uint       `gorm:"not null" json:"strategy_id"`
	StartTime       time.Time  `gorm:"not null" json:"start_time"`
	EndTime         time.Time  `gorm:"not null" json:"end_time"`
	Capital         float64    `gorm:"type:decimal(20,2);not null" json:"capital"`

	// Results
	TotalTrades     int      `gorm:"default:0" json:"total_trades"`
	WinningTrades   int      `gorm:"default:0" json:"winning_trades"`
	LosingTrades    int      `gorm:"default:0" json:"losing_trades"`
	TotalPnL        float64  `gorm:"type:decimal(20,2);default:0" json:"total_pnl"`
	TotalPnLPercent float64  `gorm:"type:decimal(10,2);default:0" json:"total_pnl_percent"`
	MaxDrawdown     float64  `gorm:"type:decimal(20,2);default:0" json:"max_drawdown"`
	WinRate         float64  `gorm:"type:decimal(5,2);default:0" json:"win_rate"`
	ProfitFactor    float64  `gorm:"type:decimal(10,2);default:0" json:"profit_factor"`

	// Status
	Status          string `gorm:"size:20;default:'PENDING'" json:"status"`
	ErrorMessage    string `gorm:"type:text" json:"error_message"`

	// Timestamps
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

func (Backtest) TableName() string {
	return "backtest"
}
