package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

// TradeExecute handles POST /api/trade/auto endpoint
// Executes automated trading based on SignalAnalyze and Money Management
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
