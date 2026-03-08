package repository

import (
	"gorm.io/gorm"

	"github.com/reshap/trading-bot/internal/models"
)

type WatchlistRepository struct {
	*GenericRepository[models.Watchlist]
}

// NewWatchlistRepository creates a new watchlist repository
func NewWatchlistRepository(db *gorm.DB) *WatchlistRepository {
	return &WatchlistRepository{
		GenericRepository: NewGenericRepository(db, &models.Watchlist{}),
	}
}

// FindAllActive finds all active watchlist symbols
func (r *WatchlistRepository) FindAllActive(tx *gorm.DB) ([]models.Watchlist, error) {
	db := r.getDB(tx)
	var watchlists []models.Watchlist
	if err := db.Model(&models.Watchlist{}).Where("is_active = ?", true).Find(&watchlists).Error; err != nil {
		return nil, err
	}
	return watchlists, nil
}

// FindBySymbol finds a watchlist by symbol
func (r *WatchlistRepository) FindBySymbol(tx *gorm.DB, symbol string) (*models.Watchlist, error) {
	db := r.getDB(tx)
	var watchlist models.Watchlist
	if err := db.Model(&models.Watchlist{}).Where("symbol = ?", symbol).First(&watchlist).Error; err != nil {
		return nil, err
	}
	return &watchlist, nil
}
