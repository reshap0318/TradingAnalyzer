package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

// SignalRawIndex handles GET /api/signal-raw endpoint
// Returns raw OHLCV data for multiple timeframes
func (c *Controller) SignalRawIndex(ctx *gin.Context) {
	var req dtos.SignalRawRequest

	// 1. Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Call service (READ operation - no transaction)
	response, err := c.srvc.SignalRawGet(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}

// SignalAnalyzeIndex handles POST /api/signal/analyze endpoint
// Analyzes market and generates trading signal
func (c *Controller) SignalAnalyzeIndex(ctx *gin.Context) {
	var req dtos.SignalAnalyzeRequest

	// 1. Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Call service (READ operation - no transaction)
	response, err := c.srvc.SignalAnalyze(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}

// SignalIndex handles GET /api/signals endpoint
// Returns paginated list of signals with optional filters
func (c *Controller) SignalIndex(ctx *gin.Context) {
	var req dtos.SignalIndexRequest

	// 1. Bind query parameters
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 3. Call service
	signals, total, err := c.srvc.SignalGetPaginated(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 4. Build response with pagination
	var response dtos.SignalIndexResponse
	response.Signals = make([]dtos.SignalData, 0, len(signals))
	
	for _, signal := range signals {
		signalData := dtos.SignalData{
			ID:                 signal.ID,
			Symbol:             signal.Symbol,
			StrategyID:         signal.StrategyID,
			SignalCategory:     signal.SignalCategory,
			SignalValid:        signal.SignalValid,
			TotalScore:         signal.TotalScore,
			Confidence:         signal.Confidence,
			CurrentPrice:       signal.CurrentPrice,
			PrimaryTimeframe:   signal.PrimaryTimeframe,
			TPPrice:            signal.TPPrice,
			SLPrice:            signal.SLPrice,
			SupportPrice:       signal.SupportPrice,
			ResistancePrice:    signal.ResistancePrice,
			RiskRewardRatio:    signal.RiskRewardRatio,
			AvgEntryPrice:      signal.AvgEntryPrice,
			EntryMode:          signal.EntryMode,
			TradingCapital:     signal.TradingCapital,
			TotalPositionValue: signal.TotalPositionValue,
			MaxRiskUSDT:        signal.MaxRiskUSDT,
			TargetProfitUSDT:   signal.TargetProfitUSDT,
			Leverage:           signal.Leverage,
			CreatedAt:          signal.CreatedAt,
			UpdatedAt:          signal.UpdatedAt,
		}
		response.Signals = append(response.Signals, signalData)
	}

	// Calculate pagination
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	response.Pagination.Page = req.Page
	response.Pagination.PageSize = req.PageSize
	response.Pagination.TotalItems = total
	response.Pagination.TotalPages = totalPages

	// 5. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}

// SignalDetail handles GET /api/signals/:id endpoint
// Returns detailed signal information with all snapshots
func (c *Controller) SignalDetail(ctx *gin.Context) {
	// 1. Parse signal ID from URL
	signalIDStr := ctx.Param("id")
	signalID, err := strconv.ParseUint(signalIDStr, 10, 64)
	if err != nil {
		helpers.RespondError(ctx, helpers.ErrNotFound)
		return
	}

	// 2. Call service
	signal, err := c.srvc.SignalGetDetail(ctx, uint(signalID))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Build response
	response := dtos.SignalDetailResponse{
		Signal: *signal,
	}

	// 4. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}

// SignalDelete handles DELETE /api/signals/:id endpoint
// Deletes a single signal by ID
func (c *Controller) SignalDelete(ctx *gin.Context) {
	// 1. Parse signal ID from URL
	signalIDStr := ctx.Param("id")
	signalID, err := strconv.ParseUint(signalIDStr, 10, 64)
	if err != nil {
		helpers.RespondError(ctx, helpers.ErrNotFound)
		return
	}

	// 2. Call service
	if err := c.srvc.SignalDelete(ctx, uint(signalID)); err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Return response
	helpers.ResponsedWithData(ctx, 200, "success", nil)
}

// SignalCleanup handles POST /api/signals/cleanup endpoint
// Cleans up signals older than specified hours
func (c *Controller) SignalCleanup(ctx *gin.Context) {
	var req dtos.SignalCleanupRequest

	// 1. Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Call service
	deletedCount, err := c.srvc.SignalCleanupOld(ctx, req.OlderThanHours)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Build response
	response := dtos.SignalCleanupResponse{
		DeletedCount:   deletedCount,
		OlderThanHours: req.OlderThanHours,
		Message:        "Successfully cleaned up old signals",
	}

	// 4. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}
