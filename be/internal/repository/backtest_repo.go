package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// BacktestRepository handles backtest data access
type BacktestRepository struct {
	*GenericRepository[models.Backtest]
}

// NewBacktestRepository creates a new backtest repository
func NewBacktestRepository(db *gorm.DB) *BacktestRepository {
	return &BacktestRepository{
		GenericRepository: NewGenericRepository(db, &models.Backtest{}),
	}
}

// FindAllOrderByCreatedAtDESC finds all backtests ordered by created_at DESC (for list view)
func (r *BacktestRepository) FindAllOrderByCreatedAtDESC(tx *gorm.DB) ([]models.Backtest, error) {
	db := r.getDB(tx)
	var backtests []models.Backtest
	err := db.Order("created_at DESC").Find(&backtests).Error
	return backtests, err
}
