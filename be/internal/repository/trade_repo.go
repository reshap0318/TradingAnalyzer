package repository

import (
	"time"

	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

type TradeRepository struct {
	*GenericRepository[models.Trade]
	Entry *TradeEntryRepository
}

// NewTradeRepository creates new trade repository
func NewTradeRepository(db *gorm.DB) *TradeRepository {
	return &TradeRepository{
		GenericRepository: NewGenericRepository(db, &models.Trade{}),
	}
}

// TradeFindActiveTrade finds active trade for symbol+interval+side
func (r *TradeRepository) TradeFindActiveBySymbol(tx *gorm.DB, symbol string) (*models.Trade, error) {
	db := r.getDB(tx)
	var trade models.Trade
	err := db.Where("symbol = ? AND status = ?", symbol, "ACTIVE").First(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

// TradeFindAllActive finds all active trades
func (r *TradeRepository) TradeFindAllActive(tx *gorm.DB) ([]models.Trade, error) {
	db := r.getDB(tx)
	var trades []models.Trade
	err := db.Preload("Entries").Where("status = ?", "ACTIVE").Find(&trades).Error
	return trades, err
}

func (r *TradeRepository) TradeFindToday(tx *gorm.DB) ([]models.Trade, error) {
	db := r.getDB(tx)

	var trades []models.Trade
	err := db.Where("date(created_at) = date(now()) and status != ?", "CANCELLED").Order("id desc").Find(&trades).Error
	return trades, err
}

// FindWithEntries finds trade with all entries loaded
func (r *TradeRepository) FindWithEntries(tx *gorm.DB, id uint) (*models.Trade, error) {
	db := r.getDB(tx)
	var trade models.Trade
	err := db.Preload("Entries").Where("id = ?", id).First(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

// FindAllActiveWithEntries finds all active trades with entries loaded
func (r *TradeRepository) FindAllActiveWithEntries(tx *gorm.DB) ([]models.Trade, error) {
	db := r.getDB(tx)
	var trades []models.Trade
	err := db.Preload("Entries").Where("status = ?", "ACTIVE").Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, err
}

// UpdateAvgEntryPrice calculates and updates average entry price from filled entries
func (r *TradeRepository) UpdateAvgEntryPrice(tx *gorm.DB, tradeID uint) error {
	db := r.getDB(tx)

	// Calculate weighted average price and total qty from filled entries
	type Result struct {
		WeightedAvg float64
		TotalQty    float64
	}

	var result Result
	err := db.Model(&models.TradeEntry{}).
		Select("COALESCE(SUM(filled_price * filled_qty) / NULLIF(SUM(filled_qty), 0), 0) as weighted_avg, COALESCE(SUM(filled_qty), 0) as total_qty").
		Where("trade_id = ? AND status = ?", tradeID, "FILLED").
		Scan(&result).Error

	if err != nil {
		return err
	}

	// Update trade with calculated values
	err = db.Model(&models.Trade{}).
		Where("id = ?", tradeID).
		Updates(map[string]interface{}{
			"avg_entry_price": result.WeightedAvg,
			"total_qty":       result.TotalQty,
		}).Error

	return err
}

// UpdateTotalQty updates total quantity from filled entries
func (r *TradeRepository) UpdateTotalQty(tx *gorm.DB, tradeID uint) error {
	db := r.getDB(tx)

	totalQty, err := r.Entry.GetTotalFilledQty(tx, tradeID)
	if err != nil {
		return err
	}

	err = db.Model(&models.Trade{}).
		Where("id = ?", tradeID).
		Update("total_qty", totalQty).Error

	return err
}

// CountActiveBySymbol counts active trades for a symbol
func (r *TradeRepository) CountActiveBySymbol(tx *gorm.DB, symbol string) (int64, error) {
	db := r.getDB(tx)
	var count int64
	err := db.Model(&models.Trade{}).Where("symbol = ? AND status = ?", symbol, "ACTIVE").Count(&count).Error
	return count, err
}

// FindTradesInSession finds trades created after executor start time (current session)
func (r *TradeRepository) FindTradesInSession(tx *gorm.DB, executorStartTime time.Time) ([]models.Trade, error) {
	db := r.getDB(tx)
	var trades []models.Trade
	err := db.Preload("Entries").
		Where("created_at > ?", executorStartTime).
		Order("id desc").
		Find(&trades).Error
	return trades, err
}

// FindAllActiveTrades finds all trades with status ACTIVE
func (r *TradeRepository) FindAllActiveTrades(tx *gorm.DB) ([]models.Trade, error) {
	db := r.getDB(tx)
	var trades []models.Trade
	err := db.Preload("Entries").
		Where("status = ?", "ACTIVE").
		Order("id desc").
		Find(&trades).Error
	return trades, err
}

// FindWithFilter finds trades with optional filters applied dynamically
// All filter fields are optional — zero/empty value means the filter is skipped
func (r *TradeRepository) FindWithFilter(tx *gorm.DB, filter dtos.TradeFilter) ([]models.Trade, error) {
	db := r.getDB(tx)
	query := db.Preload("Entries").Order("id desc")

	// Filter by status (array — IN clause)
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// Filter by symbol (array — IN clause)
	if len(filter.Symbol) > 0 {
		query = query.Where("symbol IN ?", filter.Symbol)
	}

	// Filter by interval (single value)
	if filter.Interval != "" {
		query = query.Where("interval = ?", filter.Interval)
	}

	// Filter by side (single value: BUY or SELL)
	if filter.Side != "" {
		query = query.Where("side = ?", filter.Side)
	}

	// Filter by minimum confidence threshold
	if filter.MinConf > 0 {
		query = query.Where("confidence >= ?", filter.MinConf)
	}

	// Filter by date range (date_start <= created_at)
	if filter.DateStart != "" {
		query = query.Where("DATE(created_at) >= ?", filter.DateStart)
	}

	// Filter by date range (created_at <= date_end)
	if filter.DateEnd != "" {
		query = query.Where("DATE(created_at) <= ?", filter.DateEnd)
	}

	// Apply limit if specified
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100)
	}

	var trades []models.Trade
	err := query.Find(&trades).Error
	return trades, err
}
