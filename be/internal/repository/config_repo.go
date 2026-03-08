package repository

import (
	"gorm.io/gorm"

	"github.com/reshap/trading-bot/internal/models"
)

type ConfigRepository struct {
	*GenericRepository[models.Config]
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{
		GenericRepository: NewGenericRepository(db, &models.Config{}),
	}
}

// FindByCategory finds all configs by category
func (r *ConfigRepository) FindByCategory(tx *gorm.DB, category string) ([]models.Config, error) {
	db := r.getDB(tx)
	var configs []models.Config
	if err := db.Model(&models.Config{}).Where("category = ?", category).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// FindByKey finds a config by key
func (r *ConfigRepository) FindByKey(tx *gorm.DB, configKey string) (*models.Config, error) {
	db := r.getDB(tx)
	var config models.Config
	if err := db.Model(&models.Config{}).Where("config_key = ?", configKey).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByKeyAndCategory finds a config by key and category
func (r *ConfigRepository) FindByKeyAndCategory(tx *gorm.DB, configKey, category string) (*models.Config, error) {
	db := r.getDB(tx)
	var config models.Config
	if err := db.Model(&models.Config{}).Where("config_key = ? AND category = ?", configKey, category).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
