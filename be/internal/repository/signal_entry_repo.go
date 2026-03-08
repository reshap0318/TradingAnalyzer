package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type SignalEntryRepository struct {
	*GenericRepository[models.SignalEntry]
}

// NewSignalEntryRepository creates new signal entry repository
func NewSignalEntryRepository(db *gorm.DB) *SignalEntryRepository {
	return &SignalEntryRepository{
		GenericRepository: NewGenericRepository(db, &models.SignalEntry{}),
	}
}

// FindBySignalID finds all entries for a signal ordered by entry_number
func (r *SignalEntryRepository) FindBySignalID(tx *gorm.DB, signalID uint) ([]models.SignalEntry, error) {
	db := r.getDB(tx)
	var entries []models.SignalEntry
	err := db.Where("signal_id = ?", signalID).Order("entry_number ASC").Find(&entries).Error
	return entries, err
}

// FindBySignalIDAndStatus finds entries for a signal with specific status
func (r *SignalEntryRepository) FindBySignalIDAndStatus(tx *gorm.DB, signalID uint, status string) ([]models.SignalEntry, error) {
	db := r.getDB(tx)
	var entries []models.SignalEntry
	err := db.Where("signal_id = ? AND status = ?", signalID, status).Order("entry_number ASC").Find(&entries).Error
	return entries, err
}

// FindFirstPending finds the first pending entry for a signal
func (r *SignalEntryRepository) FindFirstPending(tx *gorm.DB, signalID uint) (*models.SignalEntry, error) {
	db := r.getDB(tx)
	var entry models.SignalEntry
	err := db.Where("signal_id = ? AND status = ?", signalID, "PENDING").Order("entry_number ASC").First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteBySignalID deletes all entries for a signal
func (r *SignalEntryRepository) DeleteBySignalID(tx *gorm.DB, signalID uint) error {
	db := r.getDB(tx)
	return db.Where("signal_id = ?", signalID).Delete(&models.SignalEntry{}).Error
}

// UpdateFilled updates entry with execution details
func (r *SignalEntryRepository) UpdateFilled(
	tx *gorm.DB,
	entryID uint,
	filledPrice float64,
	filledQty float64,
	binanceOrderID int64,
	binanceStatus string,
) (*models.SignalEntry, error) {
	db := r.getDB(tx)

	entry := &models.SignalEntry{
		ID:             entryID,
		FilledPrice:    filledPrice,
		FilledQty:      filledQty,
		BinanceOrderID: binanceOrderID,
		BinanceStatus:  binanceStatus,
		Status:         "FILLED",
	}

	err := db.Model(&models.SignalEntry{}).Where("id = ?", entryID).Updates(entry).Error
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// UpdateStatus updates entry status
func (r *SignalEntryRepository) UpdateStatus(tx *gorm.DB, entryID uint, status string, rejectReason string) error {
	db := r.getDB(tx)

	updates := map[string]interface{}{
		"status": status,
	}

	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}

	err := db.Model(&models.SignalEntry{}).Where("id = ?", entryID).Updates(updates).Error
	return err
}

// CountBySignalAndStatus counts entries for a signal with specific status
func (r *SignalEntryRepository) CountBySignalAndStatus(tx *gorm.DB, signalID uint, status string) (int64, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.SignalEntry{}).Where("signal_id = ? AND status = ?", signalID, status).Count(&count).Error
	return count, err
}

// GetTotalFilledQty calculates total filled quantity for a signal
func (r *SignalEntryRepository) GetTotalFilledQty(tx *gorm.DB, signalID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		TotalQty float64
	}

	var result Result
	err := db.Model(&models.SignalEntry{}).
		Select("COALESCE(SUM(filled_qty), 0) as total_qty").
		Where("signal_id = ? AND status = ?", signalID, "FILLED").
		Scan(&result).Error

	return result.TotalQty, err
}

// GetAvgFilledPrice calculates average filled price for a signal
func (r *SignalEntryRepository) GetAvgFilledPrice(tx *gorm.DB, signalID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		AvgPrice float64
	}

	var result Result
	err := db.Model(&models.SignalEntry{}).
		Select("COALESCE(AVG(filled_price), 0) as avg_price").
		Where("signal_id = ? AND status = ?", signalID, "FILLED").
		Scan(&result).Error

	return result.AvgPrice, err
}

// GetWeightedAvgFilledPrice calculates weighted average filled price (by qty)
func (r *SignalEntryRepository) GetWeightedAvgFilledPrice(tx *gorm.DB, signalID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		WeightedAvg float64
	}

	var result Result
	err := db.Model(&models.SignalEntry{}).
		Select("COALESCE(SUM(filled_price * filled_qty) / NULLIF(SUM(filled_qty), 0), 0) as weighted_avg").
		Where("signal_id = ? AND status = ?", signalID, "FILLED").
		Scan(&result).Error

	return result.WeightedAvg, err
}
