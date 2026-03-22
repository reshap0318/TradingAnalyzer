package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type StrategyTimeframeRepository struct {
	*GenericRepository[models.StrategyTimeframe]
}

// NewStrategyTimeframeRepository creates a new strategy timeframe repository
func NewStrategyTimeframeRepository(db *gorm.DB) *StrategyTimeframeRepository {
	return &StrategyTimeframeRepository{
		GenericRepository: NewGenericRepository(db, &models.StrategyTimeframe{}),
	}
}

// DeleteByStrategyID deletes all timeframes for a strategy
func (r *StrategyTimeframeRepository) DeleteByStrategyID(tx *gorm.DB, strategyID uint) error {
	db := r.getDB(tx)
	return db.Model(&models.StrategyTimeframe{}).Where("strategy_id = ?", strategyID).Delete(&models.StrategyTimeframe{}).Error
}

type StrategyIndicatorRepository struct {
	*GenericRepository[models.StrategyIndicator]
}

// NewStrategyIndicatorRepository creates a new strategy indicator repository
func NewStrategyIndicatorRepository(db *gorm.DB) *StrategyIndicatorRepository {
	return &StrategyIndicatorRepository{
		GenericRepository: NewGenericRepository(db, &models.StrategyIndicator{}),
	}
}

// DeleteByStrategyID deletes all indicators for a strategy
func (r *StrategyIndicatorRepository) DeleteByStrategyID(tx *gorm.DB, strategyID uint) error {
	db := r.getDB(tx)
	return db.Model(&models.StrategyIndicator{}).Where("strategy_id = ?", strategyID).Delete(&models.StrategyIndicator{}).Error
}

type StrategyMoneyMgmtRepository struct {
	*GenericRepository[models.StrategyMoneyMgmt]
}

// NewStrategyMoneyMgmtRepository creates a new strategy money management repository
func NewStrategyMoneyMgmtRepository(db *gorm.DB) *StrategyMoneyMgmtRepository {
	return &StrategyMoneyMgmtRepository{
		GenericRepository: NewGenericRepository(db, &models.StrategyMoneyMgmt{}),
	}
}

// DeleteByStrategyID deletes all money management configs for a strategy
func (r *StrategyMoneyMgmtRepository) DeleteByStrategyID(tx *gorm.DB, strategyID uint) error {
	db := r.getDB(tx)
	return db.Model(&models.StrategyMoneyMgmt{}).Where("strategy_id = ?", strategyID).Delete(&models.StrategyMoneyMgmt{}).Error
}

type StrategySymbolRepository struct {
	*GenericRepository[models.StrategySymbol]
}

// NewStrategySymbolRepository creates a new strategy symbol repository
func NewStrategySymbolRepository(db *gorm.DB) *StrategySymbolRepository {
	return &StrategySymbolRepository{
		GenericRepository: NewGenericRepository(db, &models.StrategySymbol{}),
	}
}

// DeleteByStrategyID deletes all symbols for a strategy
func (r *StrategySymbolRepository) DeleteByStrategyID(tx *gorm.DB, strategyID uint) error {
	db := r.getDB(tx)
	return db.Model(&models.StrategySymbol{}).Where("strategy_id = ?", strategyID).Delete(&models.StrategySymbol{}).Error
}
