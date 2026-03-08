package service

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

func (s *Services) TimeframeGetAll(ctx *gin.Context) (res []dtos.TimeframeData, err error) {
	Timeframes, err := s.repo.Timeframe.FindAll(nil)
	if err != nil {
		return nil, err
	}

	for _, Timeframe := range Timeframes {
		res = append(res, dtos.TimeframeData{
			Name:      Timeframe.Name,
			InMinutes: Timeframe.InMinutes,
			CreatedAt: helpers.FormatDateTime(Timeframe.CreatedAt),
		})
	}

	return
}

func (s *Services) TimeframeGetByName(ctx *gin.Context, name string) (res *dtos.TimeframeData, err error) {
	Timeframe, err := s.repo.Timeframe.FindByName(nil, name)
	if err != nil {
		return nil, err
	}

	res = &dtos.TimeframeData{
		Name:      Timeframe.Name,
		InMinutes: Timeframe.InMinutes,
		CreatedAt: helpers.FormatDateTime(Timeframe.CreatedAt),
	}

	return
}

func (s *Services) TimeframeCreate(ctx *gin.Context, req *dtos.TimeframeRequest) (res *dtos.TimeframeData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		Timeframe := &models.Timeframe{
			Name:      req.Name,
			InMinutes: req.InMinutes,
		}

		Timeframe, err := s.repo.Timeframe.Create(tx, Timeframe)
		if err != nil {
			return nil, err
		}

		return &dtos.TimeframeData{
			Name:      Timeframe.Name,
			InMinutes: Timeframe.InMinutes,
			CreatedAt: helpers.FormatDateTime(Timeframe.CreatedAt),
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.TimeframeData), nil
}

func (s *Services) TimeframeUpdate(ctx *gin.Context, name string, req *dtos.TimeframeRequest) (res *dtos.TimeframeData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		Timeframe := &models.Timeframe{
			Name:      req.Name,
			InMinutes: req.InMinutes,
		}

		Timeframe, err := s.repo.Timeframe.Update(tx, &models.Timeframe{Name: name}, Timeframe)
		if err != nil {
			return nil, err
		}

		return &dtos.TimeframeData{
			Name:      Timeframe.Name,
			InMinutes: Timeframe.InMinutes,
			CreatedAt: helpers.FormatDateTime(Timeframe.CreatedAt),
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.TimeframeData), nil
}

func (s *Services) TimeframeDelete(ctx *gin.Context, name string) (res *dtos.TimeframeData, err error) {
	result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		data, err := s.repo.Timeframe.DeleteByName(tx, name)
		if err != nil {
			return nil, err
		}

		return &dtos.TimeframeData{
			Name:      data.Name,
			InMinutes: data.InMinutes,
			CreatedAt: helpers.FormatDateTime(data.CreatedAt),
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*dtos.TimeframeData), nil
}
