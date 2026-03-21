package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedTimeframes inserts default timeframe data
func SeedTimeframes(db *gorm.DB) {
	fmt.Println("Seeding timeframes...")

	// Using slice to maintain order as defined in the original map
	timeframeData := []struct {
		name      string
		inMinutes int
	}{
		// Scalping (sangat cepat, 1-15 menit hold)
		{"1m", 1},
		{"3m", 3},
		{"5m", 5},
		// Short-term scalping (15-30 menit hold)
		{"15m", 15},
		{"30m", 30},
		// Day trading (beberapa jam hold)
		{"1h", 60},
		{"2h", 120},
		{"4h", 240},
		{"6h", 360},
		{"8h", 480},
		{"12h", 720},
		// Swing trading (beberapa hari hold)
		{"1d", 1440},
		{"3d", 4320},
		// Position trading (mingguan/bulanan)
		{"1w", 10080},
		{"1M", 43200},
	}

	count := 0
	for _, data := range timeframeData {
		// Create or get interval
		var interval models.Timeframe
		result := db.Where(models.Timeframe{Name: data.name}).First(&interval)

		if result.Error == gorm.ErrRecordNotFound {
			interval = models.Timeframe{
				Name:      data.name,
				InMinutes: data.inMinutes,
			}
			if err := db.Create(&interval).Error; err != nil {
				log.Printf("Failed to create interval %s: %v", data.name, err)
				continue
			}
			count++
		} else if result.Error != nil {
			log.Printf("Failed to check interval %s: %v", data.name, result.Error)
			continue
		}
	}

	fmt.Printf("✓ Seeded %d timeframes\n", count)
}
