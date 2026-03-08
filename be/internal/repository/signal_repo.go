package repository

import (
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type SignalRepository struct {
	*GenericRepository[models.Signal]
	Entry *SignalEntryRepository
}

// NewSignalRepository creates new signal repository
func NewSignalRepository(db *gorm.DB) *SignalRepository {
	return &SignalRepository{
		GenericRepository: NewGenericRepository(db, &models.Signal{}),
	}
}

// SignalFindActiveSignal finds active signal for symbol+interval+side
func (r *SignalRepository) SignalFindActiveBySymbol(tx *gorm.DB, symbol string) (*models.Signal, error) {
	db := r.getDB(tx)
	var signal models.Signal
	err := db.Where("symbol = ? AND status = ?", symbol, "ACTIVE").First(&signal).Error
	if err != nil {
		return nil, err
	}
	return &signal, nil
}

// SignalFindAllActive finds all active signals
func (r *SignalRepository) SignalFindAllActive(tx *gorm.DB) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	err := db.Preload("Entries").Where("status = ?", "ACTIVE").Find(&signals).Error
	return signals, err
}

func (r *SignalRepository) SignalFindToday(tx *gorm.DB) ([]models.Signal, error) {
	db := r.getDB(tx)

	var signals []models.Signal
	err := db.Where("date(created_at) = date(now()) and status != ?", "CANCELLED").Find(&signals).Error
	return signals, err
}

// FindWithEntries finds signal with all entries loaded
func (r *SignalRepository) FindWithEntries(tx *gorm.DB, id uint) (*models.Signal, error) {
	db := r.getDB(tx)
	var signal models.Signal
	err := db.Preload("Entries").Where("id = ?", id).First(&signal).Error
	if err != nil {
		return nil, err
	}
	return &signal, nil
}

// FindAllActiveWithEntries finds all active signals with entries loaded
func (r *SignalRepository) FindAllActiveWithEntries(tx *gorm.DB) ([]models.Signal, error) {
	db := r.getDB(tx)
	var signals []models.Signal
	err := db.Preload("Entries").Where("status = ?", "ACTIVE").Find(&signals).Error
	if err != nil {
		return nil, err
	}
	return signals, err
}

// UpdateAvgEntryPrice calculates and updates average entry price from filled entries
func (r *SignalRepository) UpdateAvgEntryPrice(tx *gorm.DB, signalID uint) error {
	db := r.getDB(tx)

	// Calculate weighted average price and total qty from filled entries
	type Result struct {
		WeightedAvg float64
		TotalQty    float64
	}

	var result Result
	err := db.Model(&models.SignalEntry{}).
		Select("COALESCE(SUM(filled_price * filled_qty) / NULLIF(SUM(filled_qty), 0), 0) as weighted_avg, COALESCE(SUM(filled_qty), 0) as total_qty").
		Where("signal_id = ? AND status = ?", signalID, "FILLED").
		Scan(&result).Error

	if err != nil {
		return err
	}

	// Update signal with calculated values
	err = db.Model(&models.Signal{}).
		Where("id = ?", signalID).
		Updates(map[string]interface{}{
			"avg_entry_price": result.WeightedAvg,
			"total_qty":       result.TotalQty,
		}).Error

	return err
}

// UpdateTotalQty updates total quantity from filled entries
func (r *SignalRepository) UpdateTotalQty(tx *gorm.DB, signalID uint) error {
	db := r.getDB(tx)

	totalQty, err := r.Entry.GetTotalFilledQty(tx, signalID)
	if err != nil {
		return err
	}

	err = db.Model(&models.Signal{}).
		Where("id = ?", signalID).
		Update("total_qty", totalQty).Error

	return err
}

// CountActiveBySymbol counts active signals for a symbol
func (r *SignalRepository) CountActiveBySymbol(tx *gorm.DB, symbol string) (int64, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.Signal{}).Where("symbol = ? AND status = ?", symbol, "ACTIVE").Count(&count).Error
	return count, err
}
