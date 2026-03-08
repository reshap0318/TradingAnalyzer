package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// BacktestTradeRepository handles backtest trade data access
type BacktestTradeRepository struct {
	*GenericRepository[models.BacktestTrade]
}

// NewBacktestTradeRepository creates a new backtest trade repository
func NewBacktestTradeRepository(db *gorm.DB) *BacktestTradeRepository {
	return &BacktestTradeRepository{
		GenericRepository: NewGenericRepository(db, &models.BacktestTrade{}),
	}
}

// FindByBacktestID finds all trades for a backtest ordered by entry_time ASC
func (r *BacktestTradeRepository) FindByBacktestID(tx *gorm.DB, backtestID uint) ([]models.BacktestTrade, error) {
	db := r.getDB(tx)
	var trades []models.BacktestTrade
	err := db.Where("backtest_id = ?", backtestID).Order("entry_time ASC").Find(&trades).Error
	return trades, err
}

// DeleteByBacktestID deletes all trades for a backtest
func (r *BacktestTradeRepository) DeleteByBacktestID(tx *gorm.DB, backtestID uint) error {
	db := r.getDB(tx)
	return db.Where("backtest_id = ?", backtestID).Delete(&models.BacktestTrade{}).Error
}
