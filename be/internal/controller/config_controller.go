package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
)

func (c *Controller) ConfigIndex(ctx *gin.Context) {
	configs, err := c.srvc.ConfigGetAll(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", configs)
}

func (c *Controller) ConfigDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	config, err := c.srvc.ConfigGetByID(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", config)
}

func (c *Controller) ConfigCreate(ctx *gin.Context) {
	var req dtos.ConfigRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	config, err := c.srvc.ConfigCreate(ctx, &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", config)
}

func (c *Controller) ConfigUpdate(ctx *gin.Context) {
	var req dtos.ConfigRequest
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.ResponseJsonNotValid(ctx)
		return
	}

	config, err := c.srvc.ConfigUpdate(ctx, uint(id), &req)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", config)
}

func (c *Controller) ConfigDelete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	config, err := c.srvc.ConfigDelete(ctx, uint(id))
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", config)
}

func (c *Controller) ConfigGetByCategory(ctx *gin.Context) {
	category := ctx.Param("category")
	if category == "" {
		helpers.RespondWithMessage(ctx, 400, "category is required")
		return
	}

	configs, err := c.srvc.ConfigGetByCategory(ctx, category)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "success", configs)
}

func (c *Controller) ConfigReload(ctx *gin.Context) {
	result, err := c.srvc.ConfigReload(ctx)
	if err != nil {
		helpers.RespondError(ctx, err)
		return
	}

	helpers.ResponsedWithData(ctx, 200, "Configuration reloaded successfully", result)
}
