package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type TradeEntryRepository struct {
	*GenericRepository[models.TradeEntry]
}

// NewTradeEntryRepository creates new trade entry repository
func NewTradeEntryRepository(db *gorm.DB) *TradeEntryRepository {
	return &TradeEntryRepository{
		GenericRepository: NewGenericRepository(db, &models.TradeEntry{}),
	}
}

// FindByTradeID finds all entries for a trade ordered by entry_number
func (r *TradeEntryRepository) FindByTradeID(tx *gorm.DB, tradeID uint) ([]models.TradeEntry, error) {
	db := r.getDB(tx)
	var entries []models.TradeEntry
	err := db.Where("trade_id = ?", tradeID).Order("entry_number ASC").Find(&entries).Error
	return entries, err
}

// FindByTradeIDAndStatus finds entries for a trade with specific status
func (r *TradeEntryRepository) FindByTradeIDAndStatus(tx *gorm.DB, tradeID uint, status string) ([]models.TradeEntry, error) {
	db := r.getDB(tx)
	var entries []models.TradeEntry
	err := db.Where("trade_id = ? AND status = ?", tradeID, status).Order("entry_number ASC").Find(&entries).Error
	return entries, err
}

// FindFirstPending finds the first pending entry for a trade
func (r *TradeEntryRepository) FindFirstPending(tx *gorm.DB, tradeID uint) (*models.TradeEntry, error) {
	db := r.getDB(tx)
	var entry models.TradeEntry
	err := db.Where("trade_id = ? AND status = ?", tradeID, "PENDING").Order("entry_number ASC").First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteByTradeID deletes all entries for a trade
func (r *TradeEntryRepository) DeleteByTradeID(tx *gorm.DB, tradeID uint) error {
	db := r.getDB(tx)
	return db.Where("trade_id = ?", tradeID).Delete(&models.TradeEntry{}).Error
}

// UpdateFilled updates entry with execution details
func (r *TradeEntryRepository) UpdateFilled(
	tx *gorm.DB,
	entryID uint,
	filledPrice float64,
	filledQty float64,
	binanceOrderID int64,
	binanceStatus string,
) (*models.TradeEntry, error) {
	db := r.getDB(tx)

	entry := &models.TradeEntry{
		ID:             entryID,
		FilledPrice:    filledPrice,
		FilledQty:      filledQty,
		BinanceOrderID: binanceOrderID,
		BinanceStatus:  binanceStatus,
		Status:         "FILLED",
	}

	err := db.Model(&models.TradeEntry{}).Where("id = ?", entryID).Updates(entry).Error
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// UpdateStatus updates entry status
func (r *TradeEntryRepository) UpdateStatus(tx *gorm.DB, entryID uint, status string, rejectReason string) error {
	db := r.getDB(tx)

	updates := map[string]interface{}{
		"status": status,
	}

	if rejectReason != "" {
		updates["reject_reason"] = rejectReason
	}

	err := db.Model(&models.TradeEntry{}).Where("id = ?", entryID).Updates(updates).Error
	return err
}

// CountByTradeAndStatus counts entries for a trade with specific status
func (r *TradeEntryRepository) CountByTradeAndStatus(tx *gorm.DB, tradeID uint, status string) (int64, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.TradeEntry{}).Where("trade_id = ? AND status = ?", tradeID, status).Count(&count).Error
	return count, err
}

// GetTotalFilledQty calculates total filled quantity for a trade
func (r *TradeEntryRepository) GetTotalFilledQty(tx *gorm.DB, tradeID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		TotalQty float64
	}

	var result Result
	err := db.Model(&models.TradeEntry{}).
		Select("COALESCE(SUM(filled_qty), 0) as total_qty").
		Where("trade_id = ? AND status = ?", tradeID, "FILLED").
		Scan(&result).Error

	return result.TotalQty, err
}

// GetAvgFilledPrice calculates average filled price for a trade
func (r *TradeEntryRepository) GetAvgFilledPrice(tx *gorm.DB, tradeID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		AvgPrice float64
	}

	var result Result
	err := db.Model(&models.TradeEntry{}).
		Select("COALESCE(AVG(filled_price), 0) as avg_price").
		Where("trade_id = ? AND status = ?", tradeID, "FILLED").
		Scan(&result).Error

	return result.AvgPrice, err
}

// GetWeightedAvgFilledPrice calculates weighted average filled price (by qty)
func (r *TradeEntryRepository) GetWeightedAvgFilledPrice(tx *gorm.DB, tradeID uint) (float64, error) {
	db := r.getDB(tx)

	type Result struct {
		WeightedAvg float64
	}

	var result Result
	err := db.Model(&models.TradeEntry{}).
		Select("COALESCE(SUM(filled_price * filled_qty) / NULLIF(SUM(filled_qty), 0), 0) as weighted_avg").
		Where("trade_id = ? AND status = ?", tradeID, "FILLED").
		Scan(&result).Error

	return result.WeightedAvg, err
}
