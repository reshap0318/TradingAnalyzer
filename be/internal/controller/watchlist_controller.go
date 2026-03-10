package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) WatchlistIndex(ctx *gin.Context) {
	watchlists, err := c.srvc.WatchlistGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", watchlists)
}

func (c *Controller) WatchlistDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	watchlist, err := c.srvc.WatchlistGetByID(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", watchlist)
}

func (c *Controller) WatchlistCreate(ctx *gin.Context) {
	var req dtos.WatchlistRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	watchlist, err := c.srvc.WatchlistCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", watchlist)
}

func (c *Controller) WatchlistUpdate(ctx *gin.Context) {
	var req dtos.WatchlistRequest
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	watchlist, err := c.srvc.WatchlistUpdate(ctx, uint(id), &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", watchlist)
}

func (c *Controller) WatchlistDelete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	watchlist, err := c.srvc.WatchlistDelete(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", watchlist)
}

func (c *Controller) WatchlistScannerActivate(ctx *gin.Context) {
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

	scanner, err := c.srvc.WatchlistScannerActivate(ctx, req.StrategyID)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to activate scanner",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Scanner activated successfully",
		"data":    scanner,
	})
}

func (c *Controller) WatchlistScannerDeactivate(ctx *gin.Context) {
	scanner, err := c.srvc.WatchlistScannerDeactivate(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to deactivate scanner",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Scanner deactivated successfully",
		"data":    scanner,
	})
}

func (c *Controller) WatchlistScannerGetStatus(ctx *gin.Context) {
	scanner, err := c.srvc.WatchlistScannerGetStatus(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    400,
			"message": "Failed to get scanner status",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"code":    200,
		"message": "Scanner status retrieved successfully",
		"data":    scanner,
	})
}
