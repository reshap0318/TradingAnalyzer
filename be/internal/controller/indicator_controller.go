package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) IndicatorIndex(ctx *gin.Context) {
	indicators, err := c.srvc.IndicatorGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", indicators)
}

func (c *Controller) IndicatorDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	indicator, err := c.srvc.IndicatorGetByID(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", indicator)
}

func (c *Controller) IndicatorCreate(ctx *gin.Context) {
	var req dtos.IndicatorRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	indicator, err := c.srvc.IndicatorCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", indicator)
}

func (c *Controller) IndicatorUpdate(ctx *gin.Context) {
	var req dtos.IndicatorRequest
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	indicator, err := c.srvc.IndicatorUpdate(ctx, uint(id), &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", indicator)
}

func (c *Controller) IndicatorDelete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	indicator, err := c.srvc.IndicatorDelete(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", indicator)
}

func (c *Controller) IndicatorGetActive(ctx *gin.Context) {
	strategies, err := c.srvc.IndicatorGetActive(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", strategies)
}
