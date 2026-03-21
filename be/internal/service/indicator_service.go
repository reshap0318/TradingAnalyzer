package service

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
)

// IndicatorGetAvailable returns list of all available indicators with their default params
func (s *Services) IndicatorGetAvailable(ctx *gin.Context) (res []dtos.IndicatorInfo, err error) {
	res = []dtos.IndicatorInfo{
		{
			Key:         "rsi",
			Name:        "RSI",
			Description: "Relative Strength Index - Momentum oscillator",
			DefaultParams: map[string]interface{}{
				"period":     14,
				"overbought": 70,
				"oversold":   30,
			},
		},
		{
			Key:         "macd",
			Name:        "MACD",
			Description: "Moving Average Convergence Divergence - Trend following momentum",
			DefaultParams: map[string]interface{}{
				"fast":   12,
				"slow":   26,
				"signal": 9,
			},
		},
		{
			Key:         "stochastic",
			Name:        "Stochastic",
			Description: "Stochastic Oscillator - Momentum indicator",
			DefaultParams: map[string]interface{}{
				"k_period":   14,
				"d_period":   3,
				"smooth":     3,
				"overbought": 80,
				"oversold":   20,
			},
		},
		{
			Key:         "bollinger_bands",
			Name:        "Bollinger Bands",
			Description: "Bollinger Bands - Volatility indicator",
			DefaultParams: map[string]interface{}{
				"period":  20,
				"std_dev": 2.0,
			},
		},
		{
			Key:         "atr",
			Name:        "ATR",
			Description: "Average True Range - Volatility measurement",
			DefaultParams: map[string]interface{}{
				"period": 14,
			},
		},
		{
			Key:         "volume",
			Name:        "Volume",
			Description: "Volume Analysis - Volume-based indicator",
			DefaultParams: map[string]interface{}{
				"ma_period": 20,
			},
		},
		{
			Key:         "moving_average",
			Name:        "Moving Average",
			Description: "Simple and Exponential Moving Averages",
			DefaultParams: map[string]interface{}{
				"sma_periods": []int{20, 50, 200},
				"ema_periods": []int{12, 26},
			},
		},
		{
			Key:           "candle_patterns",
			Name:          "Candle Patterns",
			Description:   "Candlestick Pattern Detection",
			DefaultParams: map[string]interface{}{},
		},
	}
	return
}

func (s *Services) IndicatorGetAll(ctx *gin.Context) (res []dtos.IndicatorData, err error) {
	strategies, err := s.repo.Indicator.FindAll(nil)
	if err != nil {
		return nil, err
	}

	// Get available indicators info
	availableIndicators, _ := s.IndicatorGetAvailable(nil)
	indicatorMap := make(map[string]dtos.IndicatorInfo)
	for _, ind := range availableIndicators {
		indicatorMap[ind.Key] = ind
	}

	for _, indicator := range strategies {
		var params interface{}
		if indicator.Params != nil {
			json.Unmarshal(indicator.Params, &params)
		}

		data := dtos.IndicatorData{
			ID:          indicator.ID,
			Name:        indicator.Name,
			Description: indicator.Description,
			Indicator:   indicator.Indicator,
			Role:        indicator.Role,
			Params:      params,
			IsActive:    indicator.IsActive,
			Weight:      indicator.Weight,
			OrderView:   indicator.OrderView,
			CreatedAt:   helpers.FormatDateTime(indicator.CreatedAt),
		}

		// Add indicator info if available
		if info, ok := indicatorMap[indicator.Indicator]; ok {
			data.IndicatorInfo = &info
		}

		res = append(res, data)
	}

	return
}

func (s *Services) IndicatorGetByID(ctx *gin.Context, id uint) (res *dtos.IndicatorData, err error) {
	indicator, err := s.repo.Indicator.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	var params interface{}
	if indicator.Params != nil {
		json.Unmarshal(indicator.Params, &params)
	}

	res = &dtos.IndicatorData{
		ID:          indicator.ID,
		Name:        indicator.Name,
		Description: indicator.Description,
		Indicator:   indicator.Indicator,
		Role:        indicator.Role,
		Params:      params,
		IsActive:    indicator.IsActive,
		Weight:      indicator.Weight,
		OrderView:   indicator.OrderView,
		CreatedAt:   helpers.FormatDateTime(indicator.CreatedAt),
	}

	return
}

func (s *Services) IndicatorCreate(ctx *gin.Context, req *dtos.IndicatorRequest) (res *dtos.IndicatorData, err error) {
	// Validate weight
	if req.Weight <= 0 {
		return nil, fmt.Errorf("weight must be greater than 0")
	}

	// Check unique indicator
	existing, err := s.repo.Indicator.FindByIndicator(nil, req.Indicator)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("indicator already exists")
	}

	var params json.RawMessage
	if req.Params != nil {
		params = json.RawMessage(*req.Params)
	}

	indicator := &models.Indicators{
		Name:        req.Name,
		Description: req.Description,
		Indicator:   req.Indicator,
		Role:        req.Role,
		Params:      params,
		IsActive:    req.IsActive,
		Weight:      req.Weight,
		OrderView:   req.OrderView,
	}

	indicator, err = s.repo.Indicator.Create(nil, indicator)
	if err != nil {
		return nil, err
	}

	var paramsData interface{}
	if indicator.Params != nil {
		json.Unmarshal(indicator.Params, &paramsData)
	}

	return &dtos.IndicatorData{
		ID:          indicator.ID,
		Name:        indicator.Name,
		Description: indicator.Description,
		Indicator:   indicator.Indicator,
		Role:        indicator.Role,
		Params:      paramsData,
		IsActive:    indicator.IsActive,
		Weight:      indicator.Weight,
		OrderView:   indicator.OrderView,
		CreatedAt:   helpers.FormatDateTime(indicator.CreatedAt),
	}, nil
}

func (s *Services) IndicatorUpdate(ctx *gin.Context, id uint, req *dtos.IndicatorRequest) (res *dtos.IndicatorData, err error) {
	// Validate weight
	if req.Weight <= 0 {
		return nil, fmt.Errorf("weight must be greater than 0")
	}

	// Check unique indicator (exclude current indicator)
	existing, err := s.repo.Indicator.FindByIndicator(nil, req.Indicator)
	if err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("indicator already exists")
	}

	var params json.RawMessage
	if req.Params != nil {
		params = json.RawMessage(*req.Params)
	}

	// Direct update by ID (efficient - 1 query)
	updates := &models.Indicators{
		Name:        req.Name,
		Description: req.Description,
		Indicator:   req.Indicator,
		Role:        req.Role,
		Params:      params,
		IsActive:    req.IsActive,
		Weight:      req.Weight,
		OrderView:   req.OrderView,
	}

	_, err = s.repo.Indicator.Update(nil, &models.Indicators{ID: id}, updates)
	if err != nil {
		return nil, err
	}

	// Return response directly from request data (no refetch needed)
	var paramsData interface{}
	if req.Params != nil && len(*req.Params) > 0 {
		json.Unmarshal([]byte(*req.Params), &paramsData)
	}

	return &dtos.IndicatorData{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Indicator:   req.Indicator,
		Role:        req.Role,
		Params:      paramsData,
		IsActive:    req.IsActive,
		Weight:      req.Weight,
		OrderView:   req.OrderView,
		CreatedAt:   "", // Will be filled by client if needed
	}, nil
}

func (s *Services) IndicatorDelete(ctx *gin.Context, id uint) (res *dtos.IndicatorData, err error) {
	indicator, err := s.repo.Indicator.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Indicator.Delete(nil, id)
	if err != nil {
		return nil, err
	}

	var params interface{}
	if indicator.Params != nil {
		json.Unmarshal(indicator.Params, &params)
	}

	return &dtos.IndicatorData{
		ID:          indicator.ID,
		Name:        indicator.Name,
		Description: indicator.Description,
		Indicator:   indicator.Indicator,
		Role:        indicator.Role,
		Params:      params,
		IsActive:    indicator.IsActive,
		Weight:      indicator.Weight,
		OrderView:   indicator.OrderView,
		CreatedAt:   helpers.FormatDateTime(indicator.CreatedAt),
	}, nil
}

func (s *Services) IndicatorGetActive(ctx *gin.Context) (res []dtos.IndicatorData, err error) {
	strategies, err := s.repo.Indicator.FindAllActive(nil)
	if err != nil {
		return nil, err
	}

	for _, indicator := range strategies {
		var params interface{}
		if indicator.Params != nil {
			json.Unmarshal(indicator.Params, &params)
		}

		res = append(res, dtos.IndicatorData{
			ID:          indicator.ID,
			Name:        indicator.Name,
			Description: indicator.Description,
			Indicator:   indicator.Indicator,
			Role:        indicator.Role,
			Params:      params,
			IsActive:    indicator.IsActive,
			Weight:      indicator.Weight,
			OrderView:   indicator.OrderView,
			CreatedAt:   helpers.FormatDateTime(indicator.CreatedAt),
		})
	}

	return
}
