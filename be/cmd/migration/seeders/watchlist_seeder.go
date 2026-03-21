package seeders

import (
	"fmt"
	"log"

	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SeedWatchlist inserts default watchlist symbols
func SeedWatchlist(db *gorm.DB) {
	fmt.Println("Seeding watchlists...")

	// Default watchlist symbols
	defaultWatchlists := []models.Watchlist{
		{Symbol: "BTCUSDT", IsActive: true},
		{Symbol: "ETHUSDT", IsActive: true},
		{Symbol: "BNBUSDT", IsActive: true},
		{Symbol: "SOLUSDT", IsActive: true},
		{Symbol: "XRPUSDT", IsActive: true},
		{Symbol: "ADAUSDT", IsActive: true},
		{Symbol: "DOGEUSDT", IsActive: true},
		{Symbol: "TRXUSDT", IsActive: true},
		{Symbol: "LINKUSDT", IsActive: true},
		{Symbol: "AVAXUSDT", IsActive: true},
		{Symbol: "SUIUSDT", IsActive: true},
		{Symbol: "LTCUSDT", IsActive: true},
		{Symbol: "NEARUSDT", IsActive: true},
		{Symbol: "UNIUSDT", IsActive: true},
		{Symbol: "FETUSDT", IsActive: true},
	}

	count := 0
	for _, wl := range defaultWatchlists {
		// Use FirstOrCreate to avoid duplicates on re-seed
		var existing models.Watchlist
		result := db.Where(models.Watchlist{Symbol: wl.Symbol}).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&wl).Error; err != nil {
				log.Printf("Failed to create watchlist %s: %v", wl.Symbol, err)
			} else {
				count++
				fmt.Printf("  ✓ Created watchlist: %s\n", wl.Symbol)
			}
		} else if result.Error != nil {
			log.Printf("Failed to check watchlist %s: %v", wl.Symbol, result.Error)
		} else {
			fmt.Printf("  ⊘ Watchlist already exists: %s\n", wl.Symbol)
		}
	}

	fmt.Printf("✓ Seeded %d watchlists\n", count)
}
