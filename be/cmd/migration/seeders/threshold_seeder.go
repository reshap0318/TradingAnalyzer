package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedThreshold inserts default threshold data
func SeedThreshold(db *gorm.DB) {
	fmt.Println("Seeding thresholds...")

	// Default threshold data
	defaultThresholds := []models.Threshold{
		{
			Category:     "STRONG_BUY",
			MinValue:     70,
			MaxValue:     100,
			Action:       "BUY",
			Color:        "green",
			OrderDisplay: 1,
		},
		{
			Category:     "BUY",
			MinValue:     35,
			MaxValue:     70,
			Action:       "BUY",
			Color:        "light-green",
			OrderDisplay: 2,
		},
		{
			Category:     "WAIT",
			MinValue:     -35,
			MaxValue:     35,
			Action:       "WAIT",
			Color:        "gray",
			OrderDisplay: 3,
		},
		{
			Category:     "SELL",
			MinValue:     -70,
			MaxValue:     -35,
			Action:       "SELL",
			Color:        "red",
			OrderDisplay: 4,
		},
		{
			Category:     "STRONG_SELL",
			MinValue:     -100,
			MaxValue:     -70,
			Action:       "SELL",
			Color:        "dark-red",
			OrderDisplay: 5,
		},
	}

	count := 0
	for _, threshold := range defaultThresholds {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Threshold
		result := db.Where(models.Threshold{Category: threshold.Category, MinValue: threshold.MinValue}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&threshold).Error; err != nil {
				log.Printf("Failed to create threshold %s: %v", threshold.Category, err)
			} else {
				count++
			}
		} else if result.Error != nil {
			log.Printf("Failed to check threshold %s: %v", threshold.Category, result.Error)
		}
	}

	fmt.Printf("✓ Seeded %d thresholds\n", count)
}
