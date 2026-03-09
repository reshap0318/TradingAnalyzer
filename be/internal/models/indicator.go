package models

import (
	"encoding/json"
	"time"
)

// Indicators represents a trading Indicators in the database
type Indicators struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string          `gorm:"column:name;type:varchar(100);uniqueIndex;not null" json:"name"` // e.g., "rsi", "macd", "stochastic"
	Indicator   string          `gorm:"column:indicator;type:varchar(50);uniqueIndex;not null" json:"indicator"`
	Description string          `gorm:"column:description;type:text" json:"description"`
	Params      json.RawMessage `gorm:"column:params;type:json" json:"params"` // Custom parameters in JSON
	IsActive    bool            `gorm:"column:is_active;default:true" json:"is_active"`
	Weight      float64         `gorm:"column:weight;not null;default:1.0" json:"weight"`
	OrderView   int             `gorm:"column:order_view;not null;default:0" json:"order_view"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (Indicators) TableName() string {
	return "m_indicator"
}
