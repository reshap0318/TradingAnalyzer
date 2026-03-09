package models

import (
	"time"
)

// Watchlist represents a watchlist symbol in the database
type Watchlist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol    string    `gorm:"column:symbol;type:varchar(20);uniqueIndex;not null" json:"symbol"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Watchlist) TableName() string {
	return "watchlist"
}
