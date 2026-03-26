package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Trade represents an executed trade
type Trade struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Symbol   string `gorm:"column:symbol;size:20;not null" json:"symbol"`
	Interval string `gorm:"column:interval;size:10;not null" json:"interval"`
	Side     string `gorm:"column:side;size:10;not null" json:"side"`

	// Signal Quality
	Confidence  float64 `gorm:"column:confidence" json:"confidence"`
	TotalScore  float64 `gorm:"column:total_score;type:DECIMAL(10,3)" json:"total_score"`
	SignalLogID *uint   `gorm:"column:signal_log_id;index" json:"signal_log_id"`

	// Trading Mode
	IsAggressive bool `gorm:"column:is_aggressive;default:false" json:"is_aggressive"`

	// 🎯 SINGLE TP/SL (for total position)
	TPPrice         float64 `gorm:"column:tp_price;type:decimal(15,8)" json:"tp_price"`
	SLPrice         float64 `gorm:"column:sl_price;type:decimal(15,8)" json:"sl_price"`
	RiskRewardRatio float64 `gorm:"column:risk_reward_ratio;type:decimal(10,2)" json:"risk_reward_ratio"`

	// Average Entry (calculated)
	AvgEntryPrice float64 `gorm:"column:avg_entry_price;type:decimal(15,8)" json:"avg_entry_price"`

	// Money Management
	Leverage    int     `gorm:"column:leverage;default:5" json:"leverage"`
	CapitalUsed float64 `gorm:"column:capital_used;type:decimal(20,2)" json:"capital_used"` //usdt use
	TotalQty    float64 `gorm:"column:total_qty;type:decimal(20,8)" json:"total_qty"`       // total coin

	// Status
	Status      string `gorm:"column:status;size:20;default:'ACTIVE'" json:"status"`
	Description string `gorm:"column:description;type:text" json:"description"`

	// 🎯 TP/SL Order IDs (single set)
	TPOrderID  int64  `gorm:"column:tp_order_id" json:"tp_order_id"`
	SLOrderID  int64  `gorm:"column:sl_order_id" json:"sl_order_id"`
	TPSLStatus string `gorm:"column:tp_sl_status;size:20" json:"tp_sl_status"`

	// Exit Info
	ExitPrice  float64 `gorm:"column:exit_price;type:decimal(15,8)" json:"exit_price"`
	ExitReason string  `gorm:"column:exit_reason;size:50" json:"exit_reason"`
	PnL        float64 `gorm:"column:pnl;type:decimal(15,8)" json:"pnl"`
	PnLPct     float64 `gorm:"column:pnl_pct;type:decimal(5,2)" json:"pnl_pct"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `gorm:"column:closed_at" json:"closed_at"`

	// Relations
	Entries []TradeEntry `gorm:"foreignKey:TradeID" json:"entries"`
}

func (Trade) TableName() string {
	return "trade"
}

type TradeEntry struct {
	ID          uint `gorm:"primaryKey;column:id" json:"id"`
	TradeID     uint `gorm:"column:trade_id;not null;index" json:"trade_id"`
	EntryNumber int  `gorm:"column:entry_number;not null" json:"entry_number"`

	// Entry Price
	EntryPrice float64 `gorm:"column:entry_price;type:decimal(15,8);not null" json:"entry_price"`
	EntryType  string  `gorm:"column:entry_type;type:ENUM('LIMIT','MARKET');default:'LIMIT'" json:"entry_type"`

	// Position Sizing
	PositionSize  string  `gorm:"column:position_size;size:10" json:"position_size"`
	PositionValue float64 `gorm:"column:position_value;type:decimal(20,2)" json:"position_value"`
	PositionQty   float64 `gorm:"column:position_qty;type:decimal(20,8)" json:"position_qty"`

	// Binance Order Tracking
	BinanceOrderID int64  `gorm:"column:binance_order_id" json:"binance_order_id"`
	BinanceStatus  string `gorm:"column:binance_status;size:20" json:"binance_status"`

	// Execution Details
	FilledPrice float64    `gorm:"column:filled_price;type:decimal(15,8)" json:"filled_price"`
	FilledQty   float64    `gorm:"column:filled_qty;type:decimal(20,8)" json:"filled_qty"`
	FilledAt    *time.Time `gorm:"column:filled_at" json:"filled_at"`

	// Status
	Status       string    `gorm:"column:status;type:ENUM('PENDING','FILLED','CANCELLED','REJECTED');default:'PENDING'" json:"status"`
	RejectReason string    `gorm:"column:reject_reason;size:255" json:"reject_reason"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TradeEntry) TableName() string {
	return "trade_entry"
}

// JSONMap for JSON fields
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON: %v", value)
	}
	return json.Unmarshal(bytes, m)
}
