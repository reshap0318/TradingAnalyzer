package controller

import (
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
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	tradeBot, err := c.srvc.TradeBotActivate(ctx, req.StrategyID)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to activate trade bot",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Trade bot activated successfully",
		"data":    tradeBot,
	})
}

func (c *Controller) TradeBotDeactivate(ctx *gin.Context) {
	tradeBot, err := c.srvc.TradeBotDeactivate(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to deactivate trade bot",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Trade bot deactivated successfully",
		"data":    tradeBot,
	})
}

func (c *Controller) TradeBotGetStatus(ctx *gin.Context) {
	tradeBot, err := c.srvc.TradeBotGetStatus(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to get trade bot status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Trade bot status retrieved successfully",
		"data":    tradeBot,
	})
}

// TradeMonitorProcessAllActive processes all active trades (manual trigger)
// Called by cron job automatically, but can be triggered manually for debugging
func (c *Controller) TradeMonitorProcessAllActive(ctx *gin.Context) {
	// Create mock context for background processing
	mockCtx := &gin.Context{}

	results, err := c.srvc.TradeMonitorProcessAllActive(mockCtx)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to process active trades",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Trade monitoring completed",
		"data": gin.H{
			"total_processed": len(results),
			"results":         results,
		},
	})
}

// TradeMonitorProcessSingle processes a single active trade by ID
// Useful for debugging or re-processing specific trades
func (c *Controller) TradeMonitorProcessSingle(ctx *gin.Context) {
	var req dtos.TradeMonitorRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	result, err := c.srvc.TradeMonitorProcessSingle(ctx, &req)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to process trade",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Trade processed successfully",
		"data":    result,
	})
}
