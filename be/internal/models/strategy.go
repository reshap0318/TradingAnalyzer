package models

import (
	"time"
)

// Strategy represents a trading strategy configuration
type Strategy struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string    `gorm:"column:strategy_name;type:varchar(100);uniqueIndex;not null" json:"strategy_name"`
	PrimaryTF string    `gorm:"column:primary_tf;type:char(5);not null" json:"primary_tf"` // Reference to m_timeframe(name)
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Timeframes       []StrategyTimeframe `gorm:"foreignKey:StrategyID" json:"timeframes,omitempty"`
	IndicatorWeights []StrategyIndicator `gorm:"foreignKey:StrategyID" json:"indicator_weights,omitempty"`
	MoneyManagement  []StrategyMoneyMgmt `gorm:"foreignKey:StrategyID" json:"money_management,omitempty"`
}

func (Strategy) TableName() string {
	return "strategy"
}

// StrategyTimeframe represents the many-to-many relationship between Strategy and Timeframe
type StrategyTimeframe struct {
	ID            uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StrategyID    uint      `gorm:"column:strategy_id;not null;index" json:"strategy_id"`
	TimeframeName string    `gorm:"column:tf;type:char(5);not null" json:"tf"` // Reference to m_timeframe(name)
	Weight        float64   `gorm:"column:weight;type:decimal(5,4);not null" json:"weight"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Reference
	Timeframe Timeframe `gorm:"foreignKey:TimeframeName;references:Name" json:"timeframe_detail,omitempty"`
}

func (StrategyTimeframe) TableName() string {
	return "strategy_timeframe"
}

// StrategyIndicator represents the many-to-many relationship between Strategy and Indicator
type StrategyIndicator struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StrategyID  uint      `gorm:"column:strategy_id;not null;index" json:"strategy_id"`
	IndicatorID uint      `gorm:"column:indicator_id;not null;index" json:"indicator_id"` // Reference to m_indicator(id)
	Weight      float64   `gorm:"column:weight;type:decimal(5,4);not null" json:"weight"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Reference
	Indicator Indicators `gorm:"foreignKey:IndicatorID;references:ID" json:"indicator_detail,omitempty"`
}

func (StrategyIndicator) TableName() string {
	return "strategy_indicator"
}

// StrategyMoneyMgmt represents the many-to-many relationship between Strategy and Config
type StrategyMoneyMgmt struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	StrategyID uint      `gorm:"column:strategy_id;not null;index" json:"strategy_id"`
	Parameter  string    `gorm:"column:parameter;type:varchar(50);not null" json:"parameter"` // Reference to m_config(config_key)
	Value      string    `gorm:"column:value;type:text;not null" json:"value"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Reference
	Config Config `gorm:"foreignKey:Parameter;references:ConfigKey" json:"config_detail,omitempty"`
}

func (StrategyMoneyMgmt) TableName() string {
	return "strategy_mm" // strategy_money_management
}
