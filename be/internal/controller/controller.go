package controller

import (
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/service"
)

type Controller struct {
	srvc *service.Services
	cfg  *config.Config
}

func NewController(srvc *service.Services, cfg *config.Config) *Controller {
	return &Controller{srvc: srvc, cfg: cfg}
}
