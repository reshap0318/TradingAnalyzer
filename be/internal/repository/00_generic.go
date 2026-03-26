package repository

import (
	"github.com/reshap/trading-bot/internal/helpers"
	"gorm.io/gorm"
)

// GenericRepository provides generic CRUD operations for any model
type GenericRepository[T any] struct {
	DB    *gorm.DB
	Model *T
}

// NewGenericRepository creates a new generic repository
func NewGenericRepository[T any](db *gorm.DB, model *T) *GenericRepository[T] {
	return &GenericRepository[T]{
		DB:    db,
		Model: model,
	}
}

// getDB returns the transaction DB if provided, otherwise the default DB
func (r *GenericRepository[T]) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// FindByID finds a record by ID
func (r *GenericRepository[T]) FindByID(tx *gorm.DB, id uint) (*T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	if err := db.Model(&intanceModel).Where("id = ?", id).First(&intanceModel).Error; err != nil {
		return nil, helpers.ErrNotFound
	}
	return intanceModel, nil
}

// Create creates a new record
func (r *GenericRepository[T]) Create(tx *gorm.DB, request *T) (*T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	if err := db.Model(&intanceModel).Create(&request).Error; err != nil {
		return nil, err
	}
	return request, nil
}

// Create Many Record
func (r *GenericRepository[T]) CreateMany(tx *gorm.DB, request []T) error {
	db := r.getDB(tx)
	var intanceModel *T
	return db.Model(&intanceModel).Create(&request).Error
}

// Update updates a record by filter
func (r *GenericRepository[T]) Update(tx *gorm.DB, filter *T, update *T) (*T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	if err := db.Model(&intanceModel).Where(filter).First(&intanceModel).Error; err != nil {
		return nil, helpers.ErrNotFound
	}

	if err := db.Model(&intanceModel).Where(filter).Updates(update).Error; err != nil {
		return nil, err
	}

	return intanceModel, nil
}

// UpdateMap updates a record by filter using a map for partial updates (supports zero values)
func (r *GenericRepository[T]) UpdateMap(tx *gorm.DB, filter *T, update map[string]interface{}) (*T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	if err := db.Model(&intanceModel).Where(filter).First(&intanceModel).Error; err != nil {
		return nil, helpers.ErrNotFound
	}

	if err := db.Model(&intanceModel).Where(filter).Updates(update).Error; err != nil {
		return nil, err
	}

	return intanceModel, nil
}

// Delete deletes a record by ID
func (r *GenericRepository[T]) Delete(tx *gorm.DB, id uint) (*T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	if err := db.Model(&intanceModel).Where("id = ?", id).First(&intanceModel).Error; err != nil {
		return nil, helpers.ErrNotFound
	}

	if err := db.Model(&intanceModel).Where("id = ?", id).Delete(&intanceModel).Error; err != nil {
		return nil, err
	}

	return intanceModel, nil
}

// FindAll finds all records
func (r *GenericRepository[T]) FindAll(tx *gorm.DB) ([]T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	datas := []T{}

	if err := db.Model(&intanceModel).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
}

// FindByField finds records by filter
func (r *GenericRepository[T]) FindByField(tx *gorm.DB, filter *T) ([]T, error) {
	db := r.getDB(tx)
	var intanceModel *T
	datas := []T{}

	if err := db.Model(&intanceModel).Where(filter).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
}

// Count counts all records
func (r *GenericRepository[T]) Count(tx *gorm.DB) (int64, error) {
	db := r.getDB(tx)
	var intanceModel *T
	var count int64
	if err := db.Model(&intanceModel).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
