package models

import (
	"time"
)

// Timeframe represents a timeframe configuration in the database
type Timeframe struct {
	Name      string    `gorm:"type:char(5) COLLATE utf8mb4_bin;primaryKey;not null" json:"name"`
	InMinutes int       `gorm:"not null" json:"in_minutes"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Timeframe) TableName() string {
	return "m_timeframe"
}
