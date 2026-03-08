package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

// BacktestIndex handles GET /api/backtests
// Returns list of all backtests (without trades)
func (c *Controller) BacktestIndex(ctx *gin.Context) {
	backtests, err := c.srvc.BacktestGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", backtests)
}

// BacktestDetail handles GET /api/backtests/:id
// Returns backtest detail with trades and strategy
func (c *Controller) BacktestDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	backtest, err := c.srvc.BacktestGetByID(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", backtest)
}

// BacktestCreate handles POST /api/backtests
// Creates and runs a new backtest
func (c *Controller) BacktestCreate(ctx *gin.Context) {
	var req dtos.BacktestRequest

	// 1. Validate input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	// 2. Call service (WRITE operation with transaction)
	backtest, err := c.srvc.BacktestCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	// 3. Return response
	helpers.ResponsedWithData(ctx, 200, "success", backtest)
}

// BacktestDelete handles DELETE /api/backtests/:id
// Deletes a backtest and its trades
func (c *Controller) BacktestDelete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	backtest, err := c.srvc.BacktestDelete(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", backtest)
}
