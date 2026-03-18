package models

import (
	"time"
)

// BacktestTrade represents individual trades in a backtest with detailed entry/exit info
type BacktestTrade struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	BacktestID uint `gorm:"column:backtest_id;not null;index" json:"backtest_id"`
	TradeNum   int  `gorm:"column:trade_num;not null" json:"trade_num"` // Sequential trade number

	// Parent trade info
	Side        string  `gorm:"column:side;size:10;not null" json:"side"` // BUY or SELL
	Signal      string  `gorm:"column:signal;size:20" json:"signal"`      // STRONG_BUY, BUY, SELL, STRONG_SELL
	Confidence  float64 `gorm:"column:confidence;type:decimal(5,2)" json:"confidence"`
	TradingMode string  `gorm:"column:trading_mode;size:20" json:"trading_mode"` // AGGRESSIVE or CONSERVATIVE

	// Target prices (from trading plan)
	TakeProfit      float64 `gorm:"column:take_profit;type:decimal(15,8)" json:"take_profit"`
	StopLoss        float64 `gorm:"column:stop_loss;type:decimal(15,8)" json:"stop_loss"`
	RiskRewardRatio float64 `gorm:"column:risk_reward_ratio;type:decimal(10,2)" json:"risk_reward_ratio"`

	// Timing
	EntryTime  time.Time  `gorm:"column:entry_time;not null;index" json:"entry_time"` // When signal was generated
	FilledTime *time.Time `gorm:"column:filled_time" json:"filled_time"`              // When first entry was filled
	ExitTime   *time.Time `gorm:"column:exit_time" json:"exit_time"`                  // When position was closed

	// Aggregated fill info (for multi-entry positions)
	TotalQty      float64 `gorm:"column:total_qty;type:decimal(20,8)" json:"total_qty"`
	AvgEntryPrice float64 `gorm:"column:avg_entry_price;type:decimal(15,8)" json:"avg_entry_price"`
	ExitPrice     float64 `gorm:"column:exit_price;type:decimal(15,8)" json:"exit_price"`
	TotalCapital  float64 `gorm:"column:total_capital;type:decimal(20,2)" json:"total_capital"` // Total capital used

	// PnL calculation
	PnL        float64 `gorm:"column:pnl;type:decimal(20,2)" json:"pnl"`
	PnLPercent float64 `gorm:"column:pnl_percent;type:decimal(10,2)" json:"pnl_percent"`

	// Exit info
	ExitReason      string `gorm:"column:exit_reason;size:50" json:"exit_reason"`             // HIT_TP, HIT_SL, CLOSED_END, DEAD_SIGNAL, EXPIRED
	Status          string `gorm:"column:status;size:20;default:'ACTIVE'" json:"status"`      // ACTIVE, CLOSED, CANCELLED, EXPIRED
	DurationMinutes int64  `gorm:"column:duration_minutes;default:0" json:"duration_minutes"` // From first fill to exit

	// Entries data (stored as JSON)
	// Format: [{"entry_num": 1, "type": "MARKET", "price": 50000, "qty": 0.01, "timestamp": 1709251200000, "status": "FILLED"}, ...]
	EntriesJSON string `gorm:"column:entries_json;type:text" json:"entries_json"`

	// Daily stats snapshot at trade entry
	DailyTradeCount int     `gorm:"column:daily_trade_count;default:0" json:"daily_trade_count"`
	DailyPnL        float64 `gorm:"column:daily_pnl;type:decimal(20,2);default:0" json:"daily_pnl"`
	ConsecutiveLoss int     `gorm:"column:consecutive_loss;default:0" json:"consecutive_loss"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BacktestTrade) TableName() string {
	return "backtest_trade"
}
