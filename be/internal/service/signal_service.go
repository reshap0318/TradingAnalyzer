package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// SignalSave saves a signal to database with all snapshots
// This function expects all data to be pre-fetched - NO additional DB/Binance calls
// Parameters:
//   - ctx: Gin context
//   - analyzeRes: Analysis result (contains all signal data)
//   - tradCapital: Trading capital
//   - strategy: Pre-fetched strategy data (REQUIRED)
//   - primaryKlines: Pre-fetched klines for OHLC snapshot (REQUIRED)
func (s *Services) SignalSave(
	ctx *gin.Context,
	analyzeRes *dtos.SignalAnalyzeResponse,
	tradCapital float64,
	strategy *dtos.StrategyData,
	primaryKlines []binance.KlineInfo,
) (*models.Signal, error) {
	if analyzeRes == nil {
		return nil, fmt.Errorf("analyze response is nil")
	}
	if strategy == nil {
		return nil, fmt.Errorf("strategy is required (must be pre-fetched)")
	}
	if len(primaryKlines) == 0 {
		return nil, fmt.Errorf("primaryKlines is required (must be pre-fetched)")
	}

	// Build strategy snapshot from pre-fetched strategy data
	timeframes := make([]map[string]interface{}, 0, len(strategy.Timeframes))
	for _, tf := range strategy.Timeframes {
		timeframes = append(timeframes, map[string]interface{}{
			"name":   tf.TimeframeName,
			"weight": tf.Weight,
		})
	}

	indicatorWeights := make([]map[string]interface{}, 0, len(strategy.IndicatorWeights))
	for _, iw := range strategy.IndicatorWeights {
		iwMap := map[string]interface{}{
			"indicator": iw.IndicatorDetail.Name,
			"role":      iw.IndicatorDetail.Role,
			"weight":    iw.Weight,
		}
		if iw.TimeframeName != nil {
			iwMap["timeframe"] = *iw.TimeframeName
		}
		indicatorWeights = append(indicatorWeights, iwMap)
	}

	mmConfig := map[string]interface{}{
		"leverage":               strategy.MoneyManagement.LEVERAGE,
		"max_position_size":      strategy.MoneyManagement.MAX_POSITION_SIZE,
		"min_confidence":         strategy.MoneyManagement.MIN_CONFIDENCE,
		"is_aggressive":          strategy.MoneyManagement.IS_AGRESSIVE,
		"risk_entry_buffer":      strategy.MoneyManagement.RISK_ENTRY_BUFFER,
		"max_daily_trades":       strategy.MoneyManagement.MAX_DAILY_TRADES,
		"max_daily_loss_count":   strategy.MoneyManagement.MAX_DAILY_LOSS_COUNT,
		"max_daily_loss_percent": strategy.MoneyManagement.MAX_DAILY_LOSS_PERCENT,
		"risk_reward_ratio":      strategy.MoneyManagement.RISK_REWARD_RATIO,
		"risk_reward_target":     strategy.MoneyManagement.RISK_REWARD_TARGET,
	}

	strategySnapshot := map[string]interface{}{
		"id":                strategy.ID,
		"name":              strategy.StrategyName,
		"timeframes":        timeframes,
		"primary_timeframe": strategy.PrimaryTF,
		"indicator_weights": indicatorWeights,
		"mm_config":         mmConfig,
	}

	// Build OHLCV snapshot from pre-fetched klines
	candles := make([]map[string]interface{}, 0, len(primaryKlines))
	for _, k := range primaryKlines {
		candle := map[string]interface{}{
			"time":   k.OpenTime / 1000,
			"open":   k.Open,
			"high":   k.High,
			"low":    k.Low,
			"close":  k.Close,
			"volume": k.Volume,
		}
		candles = append(candles, candle)
	}

	ohlcSnapshot := map[string]interface{}{
		"timeframe": analyzeRes.PrimaryTimeframe,
		"candles":   candles,
	}

	// Build indicator values
	indicatorValues, err := s.buildIndicatorValues(analyzeRes)
	if err != nil {
		return nil, fmt.Errorf("failed to build indicator values: %w", err)
	}

	// Build entry levels
	entryLevels, err := s.buildEntryLevels(analyzeRes)
	if err != nil {
		return nil, fmt.Errorf("failed to build entry levels: %w", err)
	}

	// Convert strategy snapshot to JSONMap
	strategySnapshotJSON, err := s.convertToJSONMap(strategySnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to convert strategy snapshot: %w", err)
	}

	// Convert OHLC snapshot to JSONMap
	ohlcSnapshotJSON, err := s.convertToJSONMap(ohlcSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OHLC snapshot: %w", err)
	}

	// Convert indicator values to JSONArray
	indicatorValuesJSON, err := s.convertToJSONArray(indicatorValues)
	if err != nil {
		return nil, fmt.Errorf("failed to convert indicator values: %w", err)
	}

	// Convert entry levels to JSONArray
	entryLevelsJSON, err := s.convertToJSONArray(entryLevels)
	if err != nil {
		return nil, fmt.Errorf("failed to convert entry levels: %w", err)
	}

	// Build trading plan data
	tradingPlan := analyzeRes.Signal.TradingPlan
	if tradingPlan == nil {
		return nil, fmt.Errorf("trading plan is nil")
	}

	// Create signal model
	signal := &models.Signal{
		Symbol:              analyzeRes.Symbol,
		StrategyID:          0, // Will be set by caller if needed
		SignalCategory:      analyzeRes.Signal.Signal,
		SignalValid:         analyzeRes.Signal.Valid,
		TotalScore:          analyzeRes.Scoring.TotalScore,
		Confidence:          analyzeRes.Scoring.Confidence,
		CurrentPrice:        analyzeRes.Signal.CurentPrice,
		PrimaryTimeframe:    analyzeRes.PrimaryTimeframe,
		TPPrice:             tradingPlan.TakeProfit,
		SLPrice:             tradingPlan.StopLoss,
		SupportPrice:        tradingPlan.Support,
		ResistancePrice:     tradingPlan.Resistance,
		RiskRewardRatio:     tradingPlan.RiskRewardRatio,
		AvgEntryPrice:       tradingPlan.Summary.AvgEntryPrice,
		EntryMode:           tradingPlan.Mode,
		TradingCapital:      tradCapital,
		TotalPositionValue:  tradingPlan.Summary.TotalPositionValue,
		TotalPositionQty:    tradingPlan.Summary.TotalPositionQty,
		MaxRiskUSDT:         tradingPlan.Summary.MaxRiskUSDT,
		MaxRiskPercent:      tradingPlan.Summary.MaxRiskPercent,
		TargetProfitUSDT:    tradingPlan.Summary.TargetProfitUSDT,
		TargetProfitPercent: tradingPlan.Summary.TargetProfitPercent,
		EffectiveLeverage:   tradingPlan.Summary.EffectiveLeverage,
		Leverage:            0, // Will be set from config
		BufferPercent:       tradingPlan.BufferPercent,
		StrategySnapshot:    strategySnapshotJSON,
		OHLCSnapshot:        ohlcSnapshotJSON,
		IndicatorValues:     indicatorValuesJSON,
		EntryLevels:         entryLevelsJSON,
	}

	// Save to database
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		savedSignal, err := s.repo.Signal.Create(tx, signal)
		if err != nil {
			return nil, err
		}
		return savedSignal, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to save signal: %w", err)
	}

	return result.(*models.Signal), nil
}

// SignalDelete deletes a signal by ID
func (s *Services) SignalDelete(ctx *gin.Context, signalID uint) error {
	_, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// Check if signal is referenced by any trades
		var tradeCount int64
		err := tx.Model(&models.Trade{}).Where("signal_log_id = ?", signalID).Count(&tradeCount).Error
		if err != nil {
			return nil, fmt.Errorf("failed to check trade references: %w", err)
		}

		if tradeCount > 0 {
			return nil, fmt.Errorf("cannot delete signal: referenced by %d trade(s). Delete associated trades first", tradeCount)
		}

		// Delete signal if no references
		_, err = s.repo.Signal.Delete(tx, signalID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

// SignalCleanupOld deletes signals older than specified hours
// Only deletes signals that are NOT referenced by any trades (no error thrown)
// Default: 720 hours (30 days)
func (s *Services) SignalCleanupOld(ctx *gin.Context, olderThanHours int) (int64, error) {
	if olderThanHours <= 0 {
		olderThanHours = 720 // Default 30 days (720 hours)
	}

	olderThan := time.Now().Add(-time.Duration(olderThanHours) * time.Hour)

	var deletedCount int64
	_, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// Delete signals that are older than specified time and NOT referenced by trades
		count, err := s.repo.Signal.DeleteOlderThanWithoutTrades(tx, olderThan)
		if err != nil {
			return 0, err
		}
		deletedCount = count
		return count, nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old signals: %w", err)
	}

	return deletedCount, nil
}

// SignalGetByID gets a signal by ID
func (s *Services) SignalGetByID(ctx *gin.Context, signalID uint) (*models.Signal, error) {
	return s.repo.Signal.FindByID(nil, signalID)
}

// SignalGetRecent gets recent signals
func (s *Services) SignalGetRecent(ctx *gin.Context, symbol string, limit int) ([]models.Signal, error) {
	return s.repo.Signal.FindRecent(nil, symbol, limit)
}

// SignalGetPaginated gets paginated signals with filters
func (s *Services) SignalGetPaginated(ctx *gin.Context, req *dtos.SignalIndexRequest) ([]models.Signal, int64, error) {
	return s.repo.Signal.FindPaginated(nil, req)
}

// SignalGetDetail gets a signal by ID with all snapshots
func (s *Services) SignalGetDetail(ctx *gin.Context, signalID uint) (*dtos.Signal, error) {
	signal, err := s.repo.Signal.FindByID(nil, signalID)
	if err != nil {
		return nil, err
	}

	// Convert model to DTO
	response := &dtos.Signal{
		ID:                  signal.ID,
		Symbol:              signal.Symbol,
		StrategyID:          signal.StrategyID,
		SignalCategory:      signal.SignalCategory,
		SignalValid:         signal.SignalValid,
		TotalScore:          signal.TotalScore,
		Confidence:          signal.Confidence,
		CurrentPrice:        signal.CurrentPrice,
		PrimaryTimeframe:    signal.PrimaryTimeframe,
		TPPrice:             signal.TPPrice,
		SLPrice:             signal.SLPrice,
		SupportPrice:        signal.SupportPrice,
		ResistancePrice:     signal.ResistancePrice,
		RiskRewardRatio:     signal.RiskRewardRatio,
		AvgEntryPrice:       signal.AvgEntryPrice,
		EntryMode:           signal.EntryMode,
		TradingCapital:      signal.TradingCapital,
		TotalPositionValue:  signal.TotalPositionValue,
		TotalPositionQty:    signal.TotalPositionQty,
		MaxRiskUSDT:         signal.MaxRiskUSDT,
		MaxRiskPercent:      signal.MaxRiskPercent,
		TargetProfitUSDT:    signal.TargetProfitUSDT,
		TargetProfitPercent: signal.TargetProfitPercent,
		EffectiveLeverage:   signal.EffectiveLeverage,
		Leverage:            signal.Leverage,
		BufferPercent:       signal.BufferPercent,
		CreatedAt:           signal.CreatedAt,
		UpdatedAt:           signal.UpdatedAt,
	}

	// Parse JSON snapshots
	if signal.StrategySnapshot != nil {
		response.StrategySnapshot = signal.StrategySnapshot
	}
	if signal.OHLCSnapshot != nil {
		response.OHLCSnapshot = signal.OHLCSnapshot
	}

	// Convert JSONArray to []map[string]interface{}
	if signal.IndicatorValues != nil {
		indicatorValues := make([]map[string]interface{}, 0)
		for _, v := range signal.IndicatorValues {
			if m, ok := v.(map[string]interface{}); ok {
				indicatorValues = append(indicatorValues, m)
			}
		}
		response.IndicatorValues = indicatorValues
	}

	if signal.EntryLevels != nil {
		entryLevels := make([]map[string]interface{}, 0)
		for _, v := range signal.EntryLevels {
			if m, ok := v.(map[string]interface{}); ok {
				entryLevels = append(entryLevels, m)
			}
		}
		response.EntryLevels = entryLevels
	}

	return response, nil
}

// buildIndicatorValues builds indicator values array from breakdown
func (s *Services) buildIndicatorValues(analyzeRes *dtos.SignalAnalyzeResponse) ([]map[string]interface{}, error) {
	values := make([]map[string]interface{}, 0)

	for _, tfBreakdown := range analyzeRes.Scoring.Breakdown {
		for _, indicator := range tfBreakdown.Indicator {
			// Build values map based on indicator details
			indicatorValues := make(map[string]interface{})

			// Extract values from Details (slice of strings)
			if len(indicator.Details) > 0 {
				// Convert slice to map with index keys
				for i, detail := range indicator.Details {
					indicatorValues[fmt.Sprintf("detail_%d", i)] = detail
				}
			}

			// Add Value if available
			if indicator.Value != nil {
				indicatorValues["value"] = indicator.Value
			}

			valueMap := map[string]interface{}{
				"timeframe":    tfBreakdown.Timeframe,
				"name":         indicator.Name,
				"role":         indicator.Role,
				"values":       indicatorValues,
				"raw_signal":   indicator.RawSignal,
				"weight":       indicator.Weight,
				"contribution": indicator.Contribution,
			}

			if indicator.Zone != "" {
				valueMap["zone"] = indicator.Zone
			}

			values = append(values, valueMap)
		}
	}

	return values, nil
}

// buildEntryLevels builds entry levels array from trading plan
func (s *Services) buildEntryLevels(analyzeRes *dtos.SignalAnalyzeResponse) ([]map[string]interface{}, error) {
	if analyzeRes.Signal.TradingPlan == nil {
		return []map[string]interface{}{}, nil
	}

	entries := make([]map[string]interface{}, 0, len(analyzeRes.Signal.TradingPlan.Entries))

	for _, entry := range analyzeRes.Signal.TradingPlan.Entries {
		entryMap := map[string]interface{}{
			"entry_number":          entry.EntryNumber,
			"entry_price":           entry.EntryPrice,
			"position_size_percent": parsePositionSizePercent(entry.PositionSize),
			"position_value":        entry.PositionValue,
			"position_qty":          entry.PositionQty,
		}
		entries = append(entries, entryMap)
	}

	return entries, nil
}

// convertToJSONMap converts interface{} to JSONMap
func (s *Services) convertToJSONMap(data interface{}) (models.JSONMap, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var jsonMap models.JSONMap
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return nil, err
	}

	return jsonMap, nil
}

// convertToJSONArray converts interface{} to JSONArray
func (s *Services) convertToJSONArray(data interface{}) (models.JSONArray, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var jsonArray models.JSONArray
	if err := json.Unmarshal(b, &jsonArray); err != nil {
		return nil, err
	}

	return jsonArray, nil
}

// parsePositionSizePercent parses position size string to float
func parsePositionSizePercent(sizeStr string) float64 {
	// Remove "%" and parse
	if len(sizeStr) > 0 && sizeStr[len(sizeStr)-1] == '%' {
		sizeStr = sizeStr[:len(sizeStr)-1]
	}

	var result float64
	fmt.Sscanf(sizeStr, "%f", &result)
	return helpers.RoundFloat(result, 2)
}
