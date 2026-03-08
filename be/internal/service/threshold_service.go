package service

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

func (s *Services) ThresholdGetAll(ctx *gin.Context) (res []dtos.ThresholdData, err error) {
	thresholds, err := s.repo.Threshold.FindAll(nil)
	if err != nil {
		return nil, err
	}

	for _, threshold := range thresholds {
		res = append(res, dtos.ThresholdData{
			ID:           threshold.ID,
			Category:     threshold.Category,
			MinValue:     threshold.MinValue,
			MaxValue:     threshold.MaxValue,
			Action:       threshold.Action,
			Color:        threshold.Color,
			OrderDisplay: threshold.OrderDisplay,
			CreatedAt:    helpers.FormatDateTime(threshold.CreatedAt),
		})
	}

	return
}

func (s *Services) ThresholdGetByID(ctx *gin.Context, id uint) (res *dtos.ThresholdData, err error) {
	threshold, err := s.repo.Threshold.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	res = &dtos.ThresholdData{
		ID:           threshold.ID,
		Category:     threshold.Category,
		MinValue:     threshold.MinValue,
		MaxValue:     threshold.MaxValue,
		Action:       threshold.Action,
		Color:        threshold.Color,
		OrderDisplay: threshold.OrderDisplay,
		CreatedAt:    helpers.FormatDateTime(threshold.CreatedAt),
	}

	return
}

func (s *Services) ThresholdCreate(ctx *gin.Context, req *dtos.ThresholdRequest) (res *dtos.ThresholdData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		threshold, err := s.repo.Threshold.Create(tx, &models.Threshold{
			Category:     req.Category,
			MinValue:     req.MinValue,
			MaxValue:     req.MaxValue,
			Action:       req.Action,
			Color:        req.Color,
			OrderDisplay: req.OrderDisplay,
		})
		if err != nil {
			return nil, err
		}

		return &dtos.ThresholdData{
			ID:           threshold.ID,
			Category:     threshold.Category,
			MinValue:     threshold.MinValue,
			MaxValue:     threshold.MaxValue,
			Action:       threshold.Action,
			Color:        threshold.Color,
			OrderDisplay: threshold.OrderDisplay,
			CreatedAt:    helpers.FormatDateTime(threshold.CreatedAt),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*dtos.ThresholdData), nil
}

func (s *Services) ThresholdUpdate(ctx *gin.Context, id uint, req *dtos.ThresholdRequest) (res *dtos.ThresholdData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		threshold, err := s.repo.Threshold.Update(tx, &models.Threshold{ID: id}, &models.Threshold{
			Category:     req.Category,
			MinValue:     req.MinValue,
			MaxValue:     req.MaxValue,
			Action:       req.Action,
			Color:        req.Color,
			OrderDisplay: req.OrderDisplay,
		})
		if err != nil {
			return nil, err
		}

		return &dtos.ThresholdData{
			ID:           threshold.ID,
			Category:     threshold.Category,
			MinValue:     threshold.MinValue,
			MaxValue:     threshold.MaxValue,
			Action:       threshold.Action,
			Color:        threshold.Color,
			OrderDisplay: threshold.OrderDisplay,
			CreatedAt:    helpers.FormatDateTime(threshold.CreatedAt),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*dtos.ThresholdData), nil
}

func (s *Services) ThresholdDelete(ctx *gin.Context, id uint) (res *dtos.ThresholdData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		threshold, err := s.repo.Threshold.Delete(tx, id)
		if err != nil {
			return nil, err
		}

		return &dtos.ThresholdData{
			ID:           threshold.ID,
			Category:     threshold.Category,
			MinValue:     threshold.MinValue,
			MaxValue:     threshold.MaxValue,
			Action:       threshold.Action,
			Color:        threshold.Color,
			OrderDisplay: threshold.OrderDisplay,
			CreatedAt:    helpers.FormatDateTime(threshold.CreatedAt),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*dtos.ThresholdData), nil
}
