package service

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

func (s *Services) ConfigGetAll(ctx *gin.Context) (res []dtos.ConfigData, err error) {
	configs, err := s.repo.Config.FindAll(nil)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		res = append(res, dtos.ConfigData{
			ID:        cfg.ID,
			ConfigKey: cfg.ConfigKey,
			Value:     cfg.Value,
			Category:  cfg.Category,
			CreatedAt: helpers.FormatDateTime(cfg.CreatedAt),
		})
	}

	return
}

func (s *Services) ConfigGetByID(ctx *gin.Context, id uint) (res *dtos.ConfigData, err error) {
	config, err := s.repo.Config.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	res = &dtos.ConfigData{
		ID:        config.ID,
		ConfigKey: config.ConfigKey,
		Value:     config.Value,
		Category:  config.Category,
		CreatedAt: helpers.FormatDateTime(config.CreatedAt),
	}

	return
}

func (s *Services) ConfigCreate(ctx *gin.Context, req *dtos.ConfigRequest) (res *dtos.ConfigData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		config := &models.Config{
			ConfigKey:  req.ConfigKey,
			Value:      req.Value,
			Category:   req.Category,
		}

		config, err = s.repo.Config.Create(tx, config)
		if err != nil {
			return nil, err
		}

		return config, nil
	})

	if err != nil {
		return nil, err
	}

	config := result.(*models.Config)
	return &dtos.ConfigData{
		ID:        config.ID,
		ConfigKey: config.ConfigKey,
		Value:     config.Value,
		Category:  config.Category,
		CreatedAt: helpers.FormatDateTime(config.CreatedAt),
	}, nil
}

func (s *Services) ConfigUpdate(ctx *gin.Context, id uint, req *dtos.ConfigRequest) (res *dtos.ConfigData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// Get existing config first
		existing, err := s.repo.Config.FindByID(tx, id)
		if err != nil {
			return nil, err
		}

		// Update with filter using existing key and category
		filter := &models.Config{ConfigKey: existing.ConfigKey, Category: existing.Category}
		config := &models.Config{
			ConfigKey:  req.ConfigKey,
			Value:      req.Value,
			Category:   req.Category,
		}

		_, err = s.repo.Config.Update(tx, filter, config)
		if err != nil {
			return nil, err
		}

		// Fetch updated record
		updated, err := s.repo.Config.FindByID(tx, id)
		if err != nil {
			return nil, err
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	updated := result.(*models.Config)
	return &dtos.ConfigData{
		ID:        updated.ID,
		ConfigKey: updated.ConfigKey,
		Value:     updated.Value,
		Category:  updated.Category,
		CreatedAt: helpers.FormatDateTime(updated.CreatedAt),
	}, nil
}

func (s *Services) ConfigDelete(ctx *gin.Context, id uint) (res *dtos.ConfigData, err error) {
	config, err := s.repo.Config.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Config.Delete(nil, id)
	if err != nil {
		return nil, err
	}

	return &dtos.ConfigData{
		ID:        config.ID,
		ConfigKey: config.ConfigKey,
		Value:     config.Value,
		Category:  config.Category,
		CreatedAt: helpers.FormatDateTime(config.CreatedAt),
	}, nil
}

func (s *Services) ConfigGetByCategory(ctx *gin.Context, category string) (res []dtos.ConfigData, err error) {
	configs, err := s.repo.Config.FindByCategory(nil, category)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		res = append(res, dtos.ConfigData{
			ID:        cfg.ID,
			ConfigKey: cfg.ConfigKey,
			Value:     cfg.Value,
			Category:  cfg.Category,
			CreatedAt: helpers.FormatDateTime(cfg.CreatedAt),
		})
	}

	return
}

// ConfigReload reloads configuration from database and updates running services
// This function:
// 1. Loads config from database using LoadConfigDB
// 2. Updates the config pointer in services
// 3. Recreates BinanceClient with new config (if BINANCE_TESTNET changed)
func (s *Services) ConfigReload(ctx *gin.Context) (res *dtos.ConfigData, err error) {
	// 1. Load config from database (this returns new config struct)
	newCfg := config.LoadConfigDB(s)

	// 2. Update config reference in services
	// Note: This uses reflection-like approach by updating the cfg field
	// The cfg field is private, so we need to update it via the service struct
	s.cfg = newCfg

	// 3. Reload BinanceClient if testnet setting changed
	// Check if we need to recreate the client
	if s.BinanceClient != nil {
		oldConfig := s.BinanceClient.GetConfig()
		needReload := oldConfig.IsTestnet != newCfg.BINANCE.IsTestnet ||
			oldConfig.APIKey != newCfg.BINANCE.APIKey ||
			oldConfig.SecretKey != newCfg.BINANCE.SecretKey

		if needReload {
			// Close old client (cleanup resources)
			s.BinanceClient.Close()

			// Create new client with updated config
			s.BinanceClient = binance.NewClient(&newCfg.BINANCE, s.RedisClient)
		}
	}

	// Return success response
	return &dtos.ConfigData{
		ID:        0, // Not applicable for reload operation
		ConfigKey: "RELOAD",
		Value:     "success",
		Category:  "SYSTEM",
		CreatedAt: helpers.GetCurrentDateTime(),
	}, nil
}
