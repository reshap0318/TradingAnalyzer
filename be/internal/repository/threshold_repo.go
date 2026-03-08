package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type ThresholdRepository struct {
	*GenericRepository[models.Threshold]
}

// NewThresholdRepository creates a new category repository
func NewThresholdRepository(db *gorm.DB) *ThresholdRepository {
	return &ThresholdRepository{
		GenericRepository: NewGenericRepository(db, &models.Threshold{}),
	}
}