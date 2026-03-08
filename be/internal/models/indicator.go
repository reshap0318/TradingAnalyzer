package models

import (
	"encoding/json"
	"time"
)

// Indicators represents a trading Indicators in the database
type Indicators struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // e.g., "rsi", "macd", "stochastic"
	Indicator   string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"indicator"`
	Description string          `gorm:"type:text" json:"description"`
	Params      json.RawMessage `gorm:"type:json" json:"params"` // Custom parameters in JSON
	IsActive    bool            `gorm:"default:true" json:"is_active"`
	Weight      float64         `gorm:"not null;default:1.0" json:"weight"`
	OrderView   int             `gorm:"not null;default:0" json:"order_view"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (Indicators) TableName() string {
	return "m_indicator"
}
