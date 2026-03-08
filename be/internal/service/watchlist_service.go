package service

import (
	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
)

func (s *Services) WatchlistGetAll(ctx *gin.Context) (res []dtos.WatchlistData, err error) {
	watchlists, err := s.repo.Watchlist.FindAll(nil)
	if err != nil {
		return nil, err
	}

	for _, wl := range watchlists {
		res = append(res, dtos.WatchlistData{
			ID:        wl.ID,
			Symbol:    wl.Symbol,
			IsActive:  wl.IsActive,
			CreatedAt: helpers.FormatDateTime(wl.CreatedAt),
		})
	}

	return
}

func (s *Services) WatchlistGetByID(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error) {
	watchlist, err := s.repo.Watchlist.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	res = &dtos.WatchlistData{
		ID:        watchlist.ID,
		Symbol:    watchlist.Symbol,
		IsActive:  watchlist.IsActive,
		CreatedAt: helpers.FormatDateTime(watchlist.CreatedAt),
	}

	return
}

func (s *Services) WatchlistCreate(ctx *gin.Context, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error) {
	watchlist := &models.Watchlist{
		Symbol:   req.Symbol,
		IsActive: req.IsActive,
	}

	watchlist, err = s.repo.Watchlist.Create(nil, watchlist)
	if err != nil {
		return nil, err
	}

	return &dtos.WatchlistData{
		ID:        watchlist.ID,
		Symbol:    watchlist.Symbol,
		IsActive:  watchlist.IsActive,
		CreatedAt: helpers.FormatDateTime(watchlist.CreatedAt),
	}, nil
}

func (s *Services) WatchlistUpdate(ctx *gin.Context, id uint, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error) {
	// Get existing watchlist first
	existing, err := s.repo.Watchlist.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	// Update with filter using existing symbol
	filter := &models.Watchlist{Symbol: existing.Symbol}
	watchlist := &models.Watchlist{
		Symbol:   req.Symbol,
		IsActive: req.IsActive,
	}

	_, err = s.repo.Watchlist.Update(nil, filter, watchlist)
	if err != nil {
		return nil, err
	}

	// Fetch updated record
	updated, err := s.repo.Watchlist.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	return &dtos.WatchlistData{
		ID:        updated.ID,
		Symbol:    updated.Symbol,
		IsActive:  updated.IsActive,
		CreatedAt: helpers.FormatDateTime(updated.CreatedAt),
	}, nil
}

func (s *Services) WatchlistDelete(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error) {
	watchlist, err := s.repo.Watchlist.FindByID(nil, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Watchlist.Delete(nil, id)
	if err != nil {
		return nil, err
	}

	return &dtos.WatchlistData{
		ID:        watchlist.ID,
		Symbol:    watchlist.Symbol,
		IsActive:  watchlist.IsActive,
		CreatedAt: helpers.FormatDateTime(watchlist.CreatedAt),
	}, nil
}
