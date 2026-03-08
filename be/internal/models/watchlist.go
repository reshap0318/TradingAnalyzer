package models

import (
	"time"
)

// Watchlist represents a watchlist symbol in the database
type Watchlist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol    string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"symbol"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Watchlist) TableName() string {
	return "watchlist"
}
