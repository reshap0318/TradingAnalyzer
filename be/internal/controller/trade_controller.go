package controller

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) TradeExecute(ctx *gin.Context) {
	var req dtos.TradeRequest

	// 1. Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Call service (WRITE operation inside - uses transaction internally)
	response, err := c.srvc.TradeExecute(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Return response
	helpers.ResponsedWithData(ctx, 200, "success", response)
}

func (c *Controller) TradeBotActivate(ctx *gin.Context) {
	var req struct {
		StrategyID *uint `json:"strategy_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	tradeBot, err := c.srvc.TradeBotActivate(ctx, req.StrategyID)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Trade bot activated successfully", tradeBot)
}

func (c *Controller) TradeBotDeactivate(ctx *gin.Context) {
	tradeBot, err := c.srvc.TradeBotDeactivate(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Trade bot deactivated successfully", tradeBot)
}

func (c *Controller) TradeBotGetStatus(ctx *gin.Context) {
	tradeBot, err := c.srvc.TradeBotGetStatus(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Trade bot status retrieved successfully", tradeBot)
}

// TradeMonitorProcessAllActive processes all active trades (manual trigger)
func (c *Controller) TradeMonitorProcessAllActive(ctx *gin.Context) {
	// Create mock context for background processing
	mockCtx := &gin.Context{}

	results, err := c.srvc.TradeMonitorProcessAllActive(mockCtx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	response := gin.H{
		"total_processed": len(results),
		"results":         results,
	}
	helpers.ResponsedWithData(ctx, 200, "Trade monitoring completed", response)
}

// TradeMonitorProcessSingle processes a single active trade by ID
func (c *Controller) TradeMonitorProcessSingle(ctx *gin.Context) {
	var req dtos.TradeMonitorRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	result, err := c.srvc.TradeMonitorProcessSingle(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Trade processed successfully", result)
}

// TradeManualClose closes an active trade manually by user request
func (c *Controller) TradeManualClose(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		helpers.RespondError(ctx, fmt.Errorf("invalid trade ID"))
		return
	}

	result, err := c.srvc.TradeManualClose(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Trade closed manually successfully", result)
}

// TradeBotGetSessionSummary gets summary statistics for current trading session
func (c *Controller) TradeBotGetSessionSummary(ctx *gin.Context) {
	summary, err := c.srvc.TradeBotGetSessionSummary(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Session summary retrieved successfully", summary)
}

// TradeBotGetExecutedTrades gets list of trades executed in current session
func (c *Controller) TradeBotGetExecutedTrades(ctx *gin.Context) {
	trades, err := c.srvc.TradeBotGetExecutedTrades(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Executed trades retrieved successfully", trades)
}

// TradeBotGetActiveTrades gets list of currently active trades
func (c *Controller) TradeBotGetActiveTrades(ctx *gin.Context) {
	trades, err := c.srvc.TradeBotGetActiveTrades(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Active trades retrieved successfully", trades)
}
