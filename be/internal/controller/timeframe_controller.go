package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) TimeframeIndex(ctx *gin.Context) {
	timeframes, err := c.srvc.TimeframeGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", timeframes)
}

func (c *Controller) TimeframeDetail(ctx *gin.Context) {
	name := ctx.Param("name")

	timeframe, err := c.srvc.TimeframeGetByName(ctx, name)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", timeframe)
}

func (c *Controller) TimeframeCreate(ctx *gin.Context) {
	var req dtos.TimeframeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	timeframe, err := c.srvc.TimeframeCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", timeframe)
}

func (c *Controller) TimeframeUpdate(ctx *gin.Context) {
	var req dtos.TimeframeRequest
	name := ctx.Param("name")

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	timeframe, err := c.srvc.TimeframeUpdate(ctx, name, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", timeframe)
}

func (c *Controller) TimeframeDelete(ctx *gin.Context) {
	name := ctx.Param("name")

	timeframe, err := c.srvc.TimeframeDelete(ctx, name)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", timeframe)
}
