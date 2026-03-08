package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type StrategyRepository struct {
	*GenericRepository[models.Strategy]
}

// NewStrategyRepository creates a new strategy repository
func NewStrategyRepository(db *gorm.DB) *StrategyRepository {
	return &StrategyRepository{
		GenericRepository: NewGenericRepository(db, &models.Strategy{}),
	}
}

// FindAllWithDetails finds all strategies with their relationships
func (r *StrategyRepository) FindAllWithDetails(tx *gorm.DB) ([]models.Strategy, error) {
	db := r.getDB(tx)
	var strategies []models.Strategy
	if err := db.Model(&models.Strategy{}).
		Preload("Timeframes.Timeframe").
		Preload("IndicatorWeights.Indicator").
		Preload("MoneyManagement.Config").
		Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// FindByIDWithDetails finds a strategy by ID with its relationships
func (r *StrategyRepository) FindByIDWithDetails(tx *gorm.DB, id uint) (*models.Strategy, error) {
	db := r.getDB(tx)
	var strategy models.Strategy
	if err := db.Model(&models.Strategy{}).
		Preload("Timeframes.Timeframe").
		Preload("IndicatorWeights.Indicator").
		Preload("MoneyManagement.Config").
		First(&strategy, id).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// FindByNameWithDetails finds a strategy by name with its relationships
func (r *StrategyRepository) FindByNameWithDetails(tx *gorm.DB, name string) (*models.Strategy, error) {
	db := r.getDB(tx)
	var strategy models.Strategy
	if err := db.Model(&models.Strategy{}).
		Preload("Timeframes.Timeframe").
		Preload("IndicatorWeights.Indicator").
		Preload("MoneyManagement.Config").
		Where("name = ?", name).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// FindActiveWithDetails finds all active strategies with their relationships
func (r *StrategyRepository) FindActiveWithDetails(tx *gorm.DB) ([]models.Strategy, error) {
	db := r.getDB(tx)
	var strategies []models.Strategy
	if err := db.Model(&models.Strategy{}).
		Preload("Timeframes.Timeframe").
		Preload("IndicatorWeights.Indicator").
		Preload("MoneyManagement.Config").
		Where("is_active = ?", true).
		Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// FindFirstActive finds the first active strategy
func (r *StrategyRepository) FindFirstActive(tx *gorm.DB) (*models.Strategy, error) {
	db := r.getDB(tx)
	var strategy models.Strategy
	if err := db.Model(&models.Strategy{}).
		Preload("Timeframes.Timeframe").
		Preload("IndicatorWeights.Indicator").
		Preload("MoneyManagement.Config").
		Where("is_active = ?", true).
		First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// DeactivateAll deactivates all strategies
func (r *StrategyRepository) DeactivateAll(tx *gorm.DB) error {
	db := r.getDB(tx)
	return db.Model(&models.Strategy{}).Where("is_active = ?", true).Update("is_active", false).Error
}

// IsStrategyActive checks if a strategy is active
func (r *StrategyRepository) IsStrategyActive(tx *gorm.DB, id uint) (bool, error) {
	db := r.getDB(tx)
	var strategy models.Strategy
	if err := db.Model(&models.Strategy{}).Select("is_active").First(&strategy, id).Error; err != nil {
		return false, err
	}
	return strategy.IsActive, nil
}
