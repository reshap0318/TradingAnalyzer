package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedConfig inserts default configuration data
func SeedConfig(db *gorm.DB) {
	fmt.Println("Seeding configs...")

	// Default config data
	defaultConfigs := []models.Config{
		{
			ConfigKey: "MIN_CONFIDENCE",
			Value:     "45",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_TRADES",
			Value:     "10",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_LOSS_PERCENT",
			Value:     "0.05",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_DAILY_LOSS_COUNT",
			Value:     "7",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_POSITION_SIZE",
			Value:     "0.15",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "RISK_ENTRY_BUFFER",
			Value:     "0.0075",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "MAX_RISK_PER_TRADE",
			Value:     "0.04",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "RISK_REWARD_RATIO",
			Value:     "1.5",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "RISK_REWARD_TARGET",
			Value:     "3.0",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "LEVERAGE",
			Value:     "5",
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "IS_AGRESSIVE",
			Value:     "0", // true | false
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "ORDER_EXPIRATION_HOURS",
			Value:     "4", // in hour
			Category:  "MONEY_MANAGEMENT",
		},
		{
			ConfigKey: "BINANCE_TESTNET",
			Value:     "true", // true | false
			Category:  "BINANCE",
		},
	}

	count := 0
	for _, cfg := range defaultConfigs {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Config
		result := db.Where(models.Config{ConfigKey: cfg.ConfigKey, Category: cfg.Category}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&cfg).Error; err != nil {
				log.Printf("Failed to create config %s: %v", cfg.ConfigKey, err)
			} else {
				count++
			}
		} else if result.Error != nil {
			log.Printf("Failed to check config %s: %v", cfg.ConfigKey, result.Error)
		}
	}

	fmt.Printf("✓ Seeded %d configs\n", count)
}
