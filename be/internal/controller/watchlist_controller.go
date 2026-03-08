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
