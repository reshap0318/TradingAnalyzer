package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

func (s *Services) StrategyGetAll(ctx *gin.Context) (res []dtos.StrategyData, err error) {
	strategies, err := s.repo.Strategy.FindAllWithDetails(nil)
	if err != nil {
		return nil, err
	}

	for _, strategy := range strategies {
		res = append(res, s.mapStrategyToDTO(strategy))
	}

	return
}

func (s *Services) StrategyGetActive(ctx *gin.Context) (res *dtos.StrategyData, err error) {
	strategy, err := s.repo.Strategy.FindFirstActive(nil)
	if err != nil {
		return nil, err
	}

	dto := s.mapStrategyToDTO(*strategy)
	return &dto, nil
}

func (s *Services) StrategyGetByID(ctx *gin.Context, id uint) (res *dtos.StrategyData, err error) {
	strategy, err := s.repo.Strategy.FindByIDWithDetails(nil, id)
	if err != nil {
		return nil, err
	}

	dto := s.mapStrategyToDTO(*strategy)
	return &dto, nil
}

func (s *Services) StrategyCreate(ctx *gin.Context, req *dtos.StrategyRequest) (res *dtos.StrategyData, err error) {
	// Validate primary_tf exists in timeframes
	if err := s.validatePrimaryTF(req.PrimaryTF, req.Timeframes); err != nil {
		return nil, err
	}

	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// If strategy is active, deactivate all other strategies first
		if req.IsActive {
			err = s.repo.Strategy.DeactivateAll(tx)
			if err != nil {
				return nil, err
			}
		}

		// 1. Create main strategy
		strategy := &models.Strategy{
			Name:      req.StrategyName,
			PrimaryTF: req.PrimaryTF,
			IsActive:  req.IsActive,
		}

		strategy, err = s.repo.Strategy.Create(tx, strategy)
		if err != nil {
			return nil, err
		}

		// 2. Create strategy timeframes
		for _, tf := range req.Timeframes {
			strategyTF := &models.StrategyTimeframe{
				StrategyID:    strategy.ID,
				TimeframeName: tf.TF,
				Weight:        tf.Weight,
			}
			_, err = s.repo.StrategyTimeframe.Create(tx, strategyTF)
			if err != nil {
				return nil, err
			}
		}

		// 3. Create strategy indicators
		for _, ind := range req.IndicatorWeights {
			strategyInd := &models.StrategyIndicator{
				StrategyID:    strategy.ID,
				IndicatorID:   ind.IndicatorID,
				Weight:        ind.Weight,
				TimeframeName: ind.TimeframeName,
			}
			_, err = s.repo.StrategyIndicator.Create(tx, strategyInd)
			if err != nil {
				return nil, err
			}
		}

		// 4. Create strategy money management
		for _, mm := range req.MoneyManagement {
			strategyMM := &models.StrategyMoneyMgmt{
				StrategyID: strategy.ID,
				Parameter:  mm.Parameter,
				Value:      mm.Value,
			}
			_, err = s.repo.StrategyMoneyMgmt.Create(tx, strategyMM)
			if err != nil {
				return nil, err
			}
		}

		// 5. Create strategy symbols
		for _, sym := range req.Symbols {
			strategySym := &models.StrategySymbol{
				StrategyID: strategy.ID,
				Symbol:     sym.Symbol,
				IsActive:   sym.IsActive,
			}
			_, err = s.repo.StrategySymbol.Create(tx, strategySym)
			if err != nil {
				return nil, err
			}
		}

		// 6. Reload strategy with details
		strategyWithDetails, err := s.repo.Strategy.FindByIDWithDetails(tx, strategy.ID)
		if err != nil {
			return nil, err
		}

		dto := s.mapStrategyToDTO(*strategyWithDetails)
		return &dto, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.StrategyData), nil
}

func (s *Services) StrategyUpdate(ctx *gin.Context, id uint, req *dtos.StrategyRequest) (res *dtos.StrategyData, err error) {
	// Validate primary_tf exists in timeframes
	if err := s.validatePrimaryTF(req.PrimaryTF, req.Timeframes); err != nil {
		return nil, err
	}

	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// If strategy is set to active, deactivate all strategies first
		// The current strategy will be set to active in the update below
		if req.IsActive {
			err = s.repo.Strategy.DeactivateAll(tx)
			if err != nil {
				return nil, err
			}
		}

		// 1. Update main strategy
		strategyMap := map[string]interface{}{
			"strategy_name": req.StrategyName,
			"primary_tf":    req.PrimaryTF,
			"is_active":     req.IsActive,
		}

		_, err = s.repo.Strategy.UpdateMap(tx, &models.Strategy{ID: id}, strategyMap)
		if err != nil {
			return nil, err
		}

		// 2. Delete existing relationships
		err = s.repo.StrategyTimeframe.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategyIndicator.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategyMoneyMgmt.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategySymbol.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		// 3. Create new relationships
		for _, tf := range req.Timeframes {
			strategyTF := &models.StrategyTimeframe{
				StrategyID:    id,
				TimeframeName: tf.TF,
				Weight:        tf.Weight,
			}
			_, err = s.repo.StrategyTimeframe.Create(tx, strategyTF)
			if err != nil {
				return nil, err
			}
		}

		for _, ind := range req.IndicatorWeights {
			strategyInd := &models.StrategyIndicator{
				StrategyID:    id,
				IndicatorID:   ind.IndicatorID,
				Weight:        ind.Weight,
				TimeframeName: ind.TimeframeName,
			}
			_, err = s.repo.StrategyIndicator.Create(tx, strategyInd)
			if err != nil {
				return nil, err
			}
		}

		for _, mm := range req.MoneyManagement {
			strategyMM := &models.StrategyMoneyMgmt{
				StrategyID: id,
				Parameter:  mm.Parameter,
				Value:      mm.Value,
			}
			_, err = s.repo.StrategyMoneyMgmt.Create(tx, strategyMM)
			if err != nil {
				return nil, err
			}
		}

		for _, sym := range req.Symbols {
			strategySym := &models.StrategySymbol{
				StrategyID: id,
				Symbol:     sym.Symbol,
				IsActive:   sym.IsActive,
			}
			_, err = s.repo.StrategySymbol.Create(tx, strategySym)
			if err != nil {
				return nil, err
			}
		}

		// 4. Reload strategy with details
		strategyWithDetails, err := s.repo.Strategy.FindByIDWithDetails(tx, id)
		if err != nil {
			return nil, err
		}

		dto := s.mapStrategyToDTO(*strategyWithDetails)
		return &dto, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.StrategyData), nil
}

func (s *Services) StrategyDelete(ctx *gin.Context, id uint) (res *dtos.StrategyData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// 1. Get strategy with details first
		strategy, err := s.repo.Strategy.FindByIDWithDetails(tx, id)
		if err != nil {
			return nil, err
		}

		// 2. Prevent deleting active strategy
		if strategy.IsActive {
			return nil, fmt.Errorf("cannot delete active strategy. Please deactivate it first")
		}

		dto := s.mapStrategyToDTO(*strategy)

		// 3. Delete relationships (cascade)
		err = s.repo.StrategyTimeframe.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategyIndicator.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategyMoneyMgmt.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		err = s.repo.StrategySymbol.DeleteByStrategyID(tx, id)
		if err != nil {
			return nil, err
		}

		// 4. Delete main strategy
		_, err = s.repo.Strategy.Delete(tx, id)
		if err != nil {
			return nil, err
		}

		return &dto, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.StrategyData), nil
}

// mapStrategyToDTO maps strategy model to DTO
func (s *Services) mapStrategyToDTO(strategy models.Strategy) dtos.StrategyData {
	dto := dtos.StrategyData{
		ID:               strategy.ID,
		StrategyName:     strategy.Name,
		PrimaryTF:        strategy.PrimaryTF,
		IsActive:         strategy.IsActive,
		CreatedAt:        helpers.FormatDateTime(strategy.CreatedAt),
		UpdatedAt:        helpers.FormatDateTime(strategy.UpdatedAt),
		Timeframes:       []dtos.StrategyTimeframeData{},
		IndicatorWeights: []dtos.StrategyIndicatorData{},
		MoneyManagement:  s.parseMMConfigFromStrategy(strategy.MoneyManagement),
	}

	// Map timeframes
	for _, tf := range strategy.Timeframes {
		tfData := dtos.StrategyTimeframeData{
			ID:            tf.ID,
			TimeframeName: tf.TimeframeName,
			Weight:        tf.Weight,
		}
		if tf.Timeframe.Name != "" {
			tfData.TimeframeDetail = &dtos.TimeframeData{
				Name:      tf.Timeframe.Name,
				InMinutes: tf.Timeframe.InMinutes,
				CreatedAt: helpers.FormatDateTime(tf.Timeframe.CreatedAt),
			}
		}
		dto.Timeframes = append(dto.Timeframes, tfData)
	}

	// Map indicators
	for _, ind := range strategy.IndicatorWeights {
		indData := dtos.StrategyIndicatorData{
			ID:            ind.ID,
			IndicatorID:   ind.IndicatorID,
			Weight:        ind.Weight,
			TimeframeName: ind.TimeframeName,
		}
		if ind.Indicator.ID != 0 {
			indData.IndicatorDetail = &dtos.IndicatorData{
				ID:          ind.Indicator.ID,
				Name:        ind.Indicator.Name,
				Indicator:   ind.Indicator.Indicator,
				Role:        ind.Indicator.Role,
				Description: ind.Indicator.Description,
				IsActive:    ind.Indicator.IsActive,
				Weight:      ind.Indicator.Weight,
				OrderView:   ind.Indicator.OrderView,
				CreatedAt:   helpers.FormatDateTime(ind.Indicator.CreatedAt),
			}
		}
		dto.IndicatorWeights = append(dto.IndicatorWeights, indData)
	}

	// Map symbols
	for _, sym := range strategy.Symbols {
		dto.Symbols = append(dto.Symbols, dtos.StrategySymbolData{
			ID:       sym.ID,
			Symbol:   sym.Symbol,
			IsActive: sym.IsActive,
		})
	}

	return dto
}

// parseMMConfigFromStrategy parses MoneyManagement configs into MMConfigResponse struct
// Similar to LoadConfigDB but for strategy-specific money management
// Default values are 0 or false (zero values), only set if data exists in DB
func (s *Services) parseMMConfigFromStrategy(moneyMgmt []models.StrategyMoneyMgmt) *dtos.MMConfigResponse {
	// Initialize with zero values (default)
	mm := &dtos.MMConfigResponse{
		MIN_CONFIDENCE:         s.cfg.MM.MIN_CONFIDENCE,
		MAX_DAILY_TRADES:       s.cfg.MM.MAX_DAILY_TRADES,
		MAX_DAILY_LOSS_PERCENT: s.cfg.MM.MAX_DAILY_LOSS_PERCENT,
		MAX_DAILY_LOSS_COUNT:   s.cfg.MM.MAX_DAILY_LOSS_COUNT,
		RISK_REWARD_RATIO:      s.cfg.MM.RISK_REWARD_RATIO,
		RISK_REWARD_TARGET:     s.cfg.MM.RISK_REWARD_TARGET,
		RISK_ENTRY_BUFFER:      s.cfg.MM.RISK_ENTRY_BUFFER,
		MAX_POSITION_SIZE:      s.cfg.MM.MAX_POSITION_SIZE,
		LEVERAGE:               s.cfg.MM.LEVERAGE,
		IS_AGRESSIVE:           s.cfg.MM.IS_AGRESSIVE,
		ORDER_EXPIRATION_HOURS: s.cfg.MM.ORDER_EXPIRATION_HOURS,
	}

	// Set values from database
	for _, cfg := range moneyMgmt {
		switch cfg.Parameter {
		case "MIN_CONFIDENCE":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.MIN_CONFIDENCE = int8(val)
			}
		case "MAX_DAILY_TRADES":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.MAX_DAILY_TRADES = int8(val)
			}
		case "MAX_DAILY_LOSS_PERCENT":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.MAX_DAILY_LOSS_PERCENT = float32(val)
			}
		case "MAX_DAILY_LOSS_COUNT":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.MAX_DAILY_LOSS_COUNT = int8(val)
			}
		case "RISK_REWARD_RATIO":
			if val, err := helpers.ParseFloat(cfg.Value, 32); err == nil {
				mm.RISK_REWARD_RATIO = float32(val)
			}
		case "RISK_REWARD_TARGET":
			if val, err := helpers.ParseFloat(cfg.Value, 32); err == nil {
				mm.RISK_REWARD_TARGET = float32(val)
			}
		case "RISK_ENTRY_BUFFER":
			if val, err := helpers.ParseFloat(cfg.Value, 32); err == nil {
				mm.RISK_ENTRY_BUFFER = float32(val)
			}
		case "MAX_POSITION_SIZE":
			if val, err := helpers.ParseFloat(cfg.Value, 32); err == nil {
				mm.MAX_POSITION_SIZE = float32(val)
			}
		case "LEVERAGE":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.LEVERAGE = int8(val)
			}
		case "IS_AGRESSIVE":
			mm.IS_AGRESSIVE = cfg.Value == "true"
		case "ORDER_EXPIRATION_HOURS":
			if val, err := helpers.ParseFloat(cfg.Value, 8); err == nil {
				mm.ORDER_EXPIRATION_HOURS = int8(val)
			}
		}
	}

	return mm
}

// validatePrimaryTF validates that primary_tf exists in the timeframes list
func (s *Services) validatePrimaryTF(primaryTF string, timeframes []dtos.StrategyTimeframeRequest) error {
	if primaryTF == "" {
		return fmt.Errorf("primary_tf is required")
	}

	if len(timeframes) == 0 {
		return fmt.Errorf("timeframes must have at least one timeframe")
	}

	found := false
	for _, tf := range timeframes {
		if tf.TF == primaryTF {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("primary_tf '%s' must be included in timeframes list", primaryTF)
	}

	return nil
}
