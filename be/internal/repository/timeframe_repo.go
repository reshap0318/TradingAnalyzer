package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type TimeframeRepository struct {
	*GenericRepository[models.Timeframe]
}

// NewTimeframeRepository creates a new timeframe repository
func NewTimeframeRepository(db *gorm.DB) *TimeframeRepository {
	return &TimeframeRepository{
		GenericRepository: NewGenericRepository(db, &models.Timeframe{}),
	}
}

// FindByName finds a timeframe by name
func (r *TimeframeRepository) FindByName(tx *gorm.DB, name string) (*models.Timeframe, error) {
	db := r.getDB(tx)
	var timeframe models.Timeframe
	if err := db.Model(&models.Timeframe{}).Where("name = ?", name).First(&timeframe).Error; err != nil {
		return nil, err
	}
	return &timeframe, nil
}

// DeleteByName deletes a timeframe by name
func (r *TimeframeRepository) DeleteByName(tx *gorm.DB, name string) (*models.Timeframe, error) {
	db := r.getDB(tx)
	var timeframe models.Timeframe
	if err := db.Model(&models.Timeframe{}).Where("name = ?", name).First(&timeframe).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.Timeframe{}).Where("name = ?", name).Delete(&timeframe).Error; err != nil {
		return nil, err
	}

	return &timeframe, nil
}
