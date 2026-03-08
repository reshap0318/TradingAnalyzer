package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) ThresholdIndex(ctx *gin.Context) {
	thresholds, err := c.srvc.ThresholdGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", thresholds)
}

func (c *Controller) ThresholdDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	threshold, err := c.srvc.ThresholdGetByID(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", threshold)
}

func (c *Controller) ThresholdCreate(ctx *gin.Context) {
	var req dtos.ThresholdRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	threshold, err := c.srvc.ThresholdCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", threshold)
}

func (c *Controller) ThresholdUpdate(ctx *gin.Context) {
	var req dtos.ThresholdRequest
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	threshold, err := c.srvc.ThresholdUpdate(ctx, uint(id), &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", threshold)
}

func (c *Controller) ThresholdDelete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	threshold, err := c.srvc.ThresholdDelete(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", threshold)
}
