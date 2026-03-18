package models

import (
	"time"
)

// Backtest represents a backtest run
type Backtest struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"column:name;size:100;not null" json:"name"`
	Symbol     string    `gorm:"column:symbol;size:20;not null" json:"symbol"`
	StrategyID uint      `gorm:"column:strategy_id;not null" json:"strategy_id"`
	StartTime  time.Time `gorm:"column:start_time;not null" json:"start_time"`
	EndTime    time.Time `gorm:"column:end_time;not null" json:"end_time"`
	Capital    float64   `gorm:"column:capital;type:decimal(20,2);not null" json:"capital"`

	// Results - Summary
	TotalTrades     int     `gorm:"column:total_trades;default:0" json:"total_trades"`
	WinningTrades   int     `gorm:"column:winning_trades;default:0" json:"winning_trades"`
	LosingTrades    int     `gorm:"column:losing_trades;default:0" json:"losing_trades"`
	ExpiredTrades   int     `gorm:"column:expired_trades;default:0" json:"expired_trades"`
	CancelledTrades int     `gorm:"column:cancelled_trades;default:0" json:"cancelled_trades"`
	TotalPnL        float64 `gorm:"column:total_pnl;type:decimal(20,2);default:0" json:"total_pnl"`
	TotalPnLPercent float64 `gorm:"column:total_pnl_percent;type:decimal(10,2);default:0" json:"total_pnl_percent"`
	MaxDrawdownPct  float64 `gorm:"column:max_drawdown_pct;type:decimal(10,2);default:0" json:"max_drawdown_pct"`
	WinRate         float64 `gorm:"column:win_rate;type:decimal(5,2);default:0" json:"win_rate"`
	ProfitFactor    float64 `gorm:"column:profit_factor;type:decimal(10,2);default:0" json:"profit_factor"`
	AvgWin          float64 `gorm:"column:avg_win;type:decimal(20,2);default:0" json:"avg_win"`
	AvgLoss         float64 `gorm:"column:avg_loss;type:decimal(20,2);default:0" json:"avg_loss"`
	LargestWin      float64 `gorm:"column:largest_win;type:decimal(20,2);default:0" json:"largest_win"`
	LargestLoss     float64 `gorm:"column:largest_loss;type:decimal(20,2);default:0" json:"largest_loss"`

	// Equity curve data (stored as JSON)
	EquityCurveJSON string `gorm:"column:equity_curve_json;type:json" json:"equity_curve_json"`

	// Strategy snapshot (stored as JSON) - to preserve strategy config at time of backtest
	StrategyJSON string `gorm:"column:strategy_json;type:json" json:"strategy_json"`

	// Status
	Status       string `gorm:"column:status;size:20;default:'PENDING'" json:"status"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
}

func (Backtest) TableName() string {
	return "backtest"
}
