package controller

import (
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
