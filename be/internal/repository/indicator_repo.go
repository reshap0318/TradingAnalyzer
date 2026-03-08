package repository

import (
	"gorm.io/gorm"

	"github.com/reshap/trading-bot/internal/models"
)

type IndicatorRepository struct {
	*GenericRepository[models.Indicators]
}

// NewIndicatorRepository creates a new Indicator repository
func NewIndicatorRepository(db *gorm.DB) *IndicatorRepository {
	return &IndicatorRepository{
		GenericRepository: NewGenericRepository(db, &models.Indicators{}),
	}
}

// FindAllActive finds all active strategies
func (r *IndicatorRepository) FindAllActive(tx *gorm.DB) ([]models.Indicators, error) {
	db := r.getDB(tx)
	var strategies []models.Indicators
	if err := db.Model(&models.Indicators{}).Where("is_active = ?", true).Order("order_view DESC").Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// FindByName finds a Indicator by name
func (r *IndicatorRepository) FindByIndicator(tx *gorm.DB, indicator string) (*models.Indicators, error) {
	db := r.getDB(tx)
	var Indicator models.Indicators
	if err := db.Model(&models.Indicators{}).Where("indicator = ?", indicator).First(&Indicator).Error; err != nil {
		return nil, err
	}
	return &Indicator, nil
}
