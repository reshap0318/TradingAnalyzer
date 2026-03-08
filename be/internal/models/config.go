package models

import (
	"time"
)

// Config represents a configuration in the database
type Config struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"config_key"`
	Value     string    `gorm:"type:varchar(50);not null" json:"value"`
	Category  string    `gorm:"type:varchar(30);index;not null" json:"category"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Config) TableName() string {
	return "m_config"
}
