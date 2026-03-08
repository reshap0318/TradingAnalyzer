package models

import (
	"time"
)

// Threshold represents a threshold configuration in the database
type Threshold struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Category     string    `gorm:"type:char(20);uniqueIndex;not null" json:"category"`
	MinValue     int       `gorm:"column:min_value;not null" json:"min_value"`
	MaxValue     int       `gorm:"column:max_value;not null" json:"max_value"`
	Action       string    `gorm:"type:char(10);not null" json:"action"`
	Color        string    `gorm:"type:char(20);not null" json:"color"`
	OrderDisplay int       `gorm:"column:order_display;not null" json:"order_display"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Threshold) TableName() string {
	return "m_threshold"
}
