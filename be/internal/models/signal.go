package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Signal represents a trading signal
type Signal struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Symbol   string `gorm:"size:20;not null" json:"symbol"`
	Interval string `gorm:"size:10;not null" json:"interval"`
	Side     string `gorm:"size:10;not null" json:"side"`

	// Signal Quality
	Confidence float64 `json:"confidence"`
	TotalScore float64 `gorm:"type:DECIMAL(10,3)" json:"total_score"`
	RawSignal  JSONMap `gorm:"type:json" json:"raw_signal"`

	// Trading Mode
	IsAggressive bool `gorm:"default:false" json:"is_aggressive"`

	// 🎯 SINGLE TP/SL (for total position)
	TPPrice         float64 `gorm:"type:decimal(15,8);column:tp_price" json:"tp_price"`
	SLPrice         float64 `gorm:"type:decimal(15,8);column:sl_price" json:"sl_price"`
	RiskRewardRatio float64 `gorm:"type:decimal(10,2);column:risk_reward_ratio" json:"risk_reward_ratio"`

	// Average Entry (calculated)
	AvgEntryPrice float64 `gorm:"type:decimal(15,8);column:avg_entry_price" json:"avg_entry_price"`

	// Money Management
	Leverage    int     `gorm:"default:5" json:"leverage"`
	CapitalUsed float64 `gorm:"type:decimal(20,2);column:capital_used" json:"capital_used"` //usdt use
	TotalQty    float64 `gorm:"type:decimal(20,8);column:total_qty" json:"total_qty"`       // total coin

	// Status
	Status      string `gorm:"size:20;default:'ACTIVE'" json:"status"`
	Description string `gorm:"type:text" json:"description"`

	// 🎯 TP/SL Order IDs (single set)
	TPOrderID  int64  `gorm:"column:tp_order_id" json:"tp_order_id"`
	SLOrderID  int64  `gorm:"column:sl_order_id" json:"sl_order_id"`
	TPSLStatus string `gorm:"size:20;column:tp_sl_status" json:"tp_sl_status"`

	// Exit Info
	ExitPrice  float64 `gorm:"type:decimal(15,8);column:exit_price" json:"exit_price"`
	ExitReason string  `gorm:"size:50;column:exit_reason" json:"exit_reason"`
	PnL        float64 `gorm:"type:decimal(15,8)" json:"pnl"`
	PnLPct     float64 `gorm:"type:decimal(5,2);column:pnl_pct" json:"pnl_pct"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`

	// Relations
	Entries []SignalEntry `gorm:"foreignKey:SignalID" json:"entries"`
}

func (Signal) TableName() string {
	return "signal"
}

type SignalEntry struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	SignalID    uint `gorm:"not null;index" json:"signal_id"`
	EntryNumber int  `gorm:"not null" json:"entry_number"`

	// Entry Price
	EntryPrice float64 `gorm:"type:decimal(15,8);not null" json:"entry_price"`
	EntryType  string  `gorm:"type:ENUM('LIMIT','MARKET');default:'LIMIT'" json:"entry_type"`

	// Position Sizing
	PositionSize  string  `gorm:"size:10" json:"position_size"`
	PositionValue float64 `gorm:"type:decimal(20,2)" json:"position_value"`
	PositionQty   float64 `gorm:"type:decimal(20,8)" json:"position_qty"`

	// Binance Order Tracking
	BinanceOrderID int64  `json:"binance_order_id"`
	BinanceStatus  string `gorm:"size:20" json:"binance_status"`

	// Execution Details
	FilledPrice float64    `gorm:"type:decimal(15,8)" json:"filled_price"`
	FilledQty   float64    `gorm:"type:decimal(20,8)" json:"filled_qty"`
	FilledAt    *time.Time `json:"filled_at"`

	// Status
	Status       string    `gorm:"type:ENUM('PENDING','FILLED','CANCELLED','REJECTED');default:'PENDING'" json:"status"`
	RejectReason string    `gorm:"size:255" json:"reject_reason"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SignalEntry) TableName() string {
	return "signal_entry"
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
