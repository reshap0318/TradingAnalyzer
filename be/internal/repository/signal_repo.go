package repository

import (
	"time"

	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SignalRepository provides signal data access operations
type SignalRepository struct {
	*GenericRepository[models.Signal]
}

// NewSignalRepository creates a new signal repository
func NewSignalRepository(db *gorm.DB) *SignalRepository {
	return &SignalRepository{
		GenericRepository: NewGenericRepository(db, &models.Signal{}),
	}
}

// FindBySymbol finds signals by symbol
func (r *SignalRepository) FindBySymbol(tx *gorm.DB, symbol string) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	if err := db.Model(&models.Signal{}).Where("symbol = ?", symbol).Order("created_at DESC").Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// FindBySymbolAndTimeframe finds signals by symbol and timeframe
func (r *SignalRepository) FindBySymbolAndTimeframe(tx *gorm.DB, symbol, timeframe string) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	if err := db.Model(&models.Signal{}).Where("symbol = ? AND primary_timeframe = ?", symbol, timeframe).Order("created_at DESC").Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// FindRecent finds recent signals with limit
func (r *SignalRepository) FindRecent(tx *gorm.DB, symbol string, limit int) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	query := db.Model(&models.Signal{}).Order("created_at DESC")
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// FindByStrategy finds signals by strategy ID
func (r *SignalRepository) FindByStrategy(tx *gorm.DB, strategyID uint) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	if err := db.Model(&models.Signal{}).Where("strategy_id = ?", strategyID).Order("created_at DESC").Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// FindValidSignals finds valid signals (signal_valid = true)
func (r *SignalRepository) FindValidSignals(tx *gorm.DB, symbol string, limit int) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	query := db.Model(&models.Signal{}).Where("signal_valid = ?", true).Order("created_at DESC")
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// FindByDateRange finds signals within a date range
func (r *SignalRepository) FindByDateRange(tx *gorm.DB, startTime, endTime time.Time) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	if err := db.Model(&models.Signal{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Order("created_at DESC").
		Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

// DeleteOlderThan deletes signals older than specified duration
func (r *SignalRepository) DeleteOlderThan(tx *gorm.DB, olderThan time.Time) (int64, error) {
	db := r.getDB(tx)
	var signal models.Signal
	result := db.Model(&signal).Where("created_at < ?", olderThan).Delete(&signal)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// CountBySymbolAndDate counts signals by symbol and date
func (r *SignalRepository) CountBySymbolAndDate(tx *gorm.DB, symbol string, date time.Time) (int64, error) {
	db := r.getDB(tx)
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := db.Model(&models.Signal{}).
		Where("symbol = ? AND created_at BETWEEN ? AND ?", symbol, startOfDay, endOfDay).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindPaginated finds signals with pagination and filters
func (r *SignalRepository) FindPaginated(tx *gorm.DB, req *dtos.SignalIndexRequest) ([]models.Signal, int64, error) {
	db := r.getDB(tx)
	
	// Build query with filters
	query := db.Model(&models.Signal{})
	
	// Apply filters
	if req.Symbol != "" {
		query = query.Where("symbol = ?", req.Symbol)
	}
	if req.StrategyID > 0 {
		query = query.Where("strategy_id = ?", req.StrategyID)
	}
	if req.SignalCategory != "" {
		query = query.Where("signal_category = ?", req.SignalCategory)
	}
	if req.SignalValid != nil {
		query = query.Where("signal_valid = ?", *req.SignalValid)
	}
	if req.StartTime != "" {
		if startTime, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if req.EndTime != "" {
		if endTime, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			query = query.Where("created_at <= ?", endTime)
		}
	}
	
	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	
	// Execute query with pagination
	var signals []models.Signal
	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&signals).Error; err != nil {
		return nil, 0, err
	}
	
	return signals, total, nil
}
