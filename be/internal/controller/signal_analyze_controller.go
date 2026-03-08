package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

// SignalAnalyzeIndex handles POST /api/signal/analyze endpoint
// Analyzes market data and generates trading signal
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
