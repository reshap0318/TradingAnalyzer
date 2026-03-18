package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/config"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// CRUD Operations
// ============================================================================

// BacktestGetAll lists all backtests (without trades)
func (s *Services) BacktestGetAll(ctx *gin.Context) (res []dtos.BacktestListItem, err error) {
	backtests, err := s.repo.Backtest.FindAllOrderByCreatedAtDESC(nil)
	if err != nil {
		return nil, err
	}

	res = make([]dtos.BacktestListItem, len(backtests))
	for i, bt := range backtests {
		strategyName := ""
		strategy, stratErr := s.StrategyGetByID(ctx, bt.StrategyID)
		if stratErr == nil && strategy != nil {
			strategyName = strategy.StrategyName
		}

		res[i] = dtos.BacktestListItem{
			ID:              bt.ID,
			Name:            bt.Name,
			Symbol:          bt.Symbol,
			StrategyName:    strategyName,
			TotalPnL:        bt.TotalPnL,
			TotalPnLPercent: bt.TotalPnLPercent,
			WinRate:         bt.WinRate,
			TotalTrades:     bt.TotalTrades,
			Status:          bt.Status,
			CreatedAt:       bt.CreatedAt,
		}
	}

	return
}

// BacktestGetByID gets backtest by ID with trades, strategy detail, and OHLCV data
func (s *Services) BacktestGetByID(ctx *gin.Context, id uint) (res *dtos.BacktestResponse, err error) {
	bt, err := s.repo.Backtest.FindByID(nil, id)
	if err != nil {
		return nil, fmt.Errorf("backtest not found: %w", err)
	}

	trades, err := s.repo.BacktestTrade.FindByBacktestID(nil, bt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load trades: %w", err)
	}

	// Try to load strategy from JSON snapshot first, fallback to current strategy
	var strategy *dtos.StrategyData
	if bt.StrategyJSON != "" {
		// Unmarshal from JSON snapshot
		strategy = &dtos.StrategyData{}
		json.Unmarshal([]byte(bt.StrategyJSON), strategy)
	} else {
		// Fallback to current strategy from DB
		strategy, _ = s.StrategyGetByID(ctx, bt.StrategyID)
	}

	// Parse equity curve JSON
	var equityCurve []dtos.EquityPoint
	if bt.EquityCurveJSON != "" {
		json.Unmarshal([]byte(bt.EquityCurveJSON), &equityCurve)
	}

	// Fetch OHLCV data from Binance
	var ohlcv []dtos.CandleData
	if strategy != nil && strategy.PrimaryTF != "" {
		ohlcv, err = s.backtestFetchOHLCV(bt.Symbol, strategy.PrimaryTF, bt.StartTime, bt.EndTime)
		if err != nil {
			fmt.Printf("⚠️  [BACKTEST] Failed to fetch OHLCV: %v\n", err)
			// Continue without OHLCV data
		}
	}

	// Convert trades to DTOs
	tradeDTOs := make([]dtos.BacktestTradeDTO, len(trades))
	for i, t := range trades {
		tradeDTOs[i] = s.convertBacktestTradeToDTO(t)
	}

	// Build summary
	summary := dtos.BacktestSummary{
		InitialBalance:   bt.Capital,
		FinalBalance:     bt.Capital + bt.TotalPnL,
		NetProfit:        bt.TotalPnL,
		NetProfitPercent: bt.TotalPnLPercent,
		WinRate:          bt.WinRate,
		TotalTrades:      bt.TotalTrades,
		WinningTrades:    bt.WinningTrades,
		LosingTrades:     bt.LosingTrades,
		ExpiredTrades:    bt.ExpiredTrades,
		CancelledTrades:  bt.CancelledTrades,
		MaxDrawdown:      bt.MaxDrawdownPct,
		ProfitFactor:     bt.ProfitFactor,
		AvgWin:           bt.AvgWin,
		AvgLoss:          bt.AvgLoss,
		LargestWin:       bt.LargestWin,
		LargestLoss:      bt.LargestLoss,
	}

	return &dtos.BacktestResponse{
		ID:           bt.ID,
		Name:         bt.Name,
		Symbol:       bt.Symbol,
		StrategyID:   bt.StrategyID,
		StartTime:    bt.StartTime,
		EndTime:      bt.EndTime,
		Capital:      bt.Capital,
		Summary:      summary,
		EquityCurve:  equityCurve,
		OHLCV:        ohlcv,
		Trades:       tradeDTOs,
		Status:       bt.Status,
		ErrorMessage: bt.ErrorMessage,
		CreatedAt:    bt.CreatedAt,
		CompletedAt:  bt.CompletedAt,
		Strategy:     strategy,
	}, nil
}

// backtestFetchOHLCV fetches OHLCV data from Binance for the specified time range
func (s *Services) backtestFetchOHLCV(symbol, timeframe string, startTime, endTime time.Time) ([]dtos.CandleData, error) {
	klines, err := s.BinanceClient.GetKlinesWithStartTime(symbol, timeframe, 1000, startTime.UnixMilli())
	if err != nil {
		return nil, err
	}

	// Filter by end time and convert to CandleData
	var candles []dtos.CandleData
	endTimeMs := endTime.UnixMilli()

	for _, k := range klines {
		if k.OpenTime > endTimeMs {
			break
		}
		candles = append(candles, dtos.CandleData{
			Timestamp: k.OpenTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		})
	}

	return candles, nil
}

// convertBacktestTradeToDTO converts model.BacktestTrade to DTO with parsed JSON
func (s *Services) convertBacktestTradeToDTO(trade models.BacktestTrade) dtos.BacktestTradeDTO {
	// Parse entries JSON
	var entries []dtos.TradeEntry
	if trade.EntriesJSON != "" {
		json.Unmarshal([]byte(trade.EntriesJSON), &entries)
	}

	dto := dtos.BacktestTradeDTO{
		TradeID:     trade.ID,
		TradeNum:    trade.TradeNum,
		Side:        trade.Side,
		Signal:      trade.Signal,
		Confidence:  trade.Confidence,
		TradingMode: trade.TradingMode,
		Status:      trade.Status,
		Targets: dtos.TradeTargets{
			TPPrice: trade.TakeProfit,
			SLPrice: trade.StopLoss,
			Ratio:   trade.RiskRewardRatio,
		},
		Entries:         entries,
		TotalQty:        trade.TotalQty,
		AvgEntryPrice:   trade.AvgEntryPrice,
		TotalCapital:    trade.TotalCapital,
		PnL:             trade.PnL,
		PnLPercent:      trade.PnLPercent,
		EntryTime:       trade.EntryTime,
		FilledTime:      trade.FilledTime,
		ExitTime:        trade.ExitTime,
		DurationMinutes: trade.DurationMinutes,
		DailyStats: &dtos.DailyStatsSnapshot{
			TradeCount:      trade.DailyTradeCount,
			PnL:             trade.DailyPnL,
			ConsecutiveLoss: trade.ConsecutiveLoss,
		},
	}

	// Add exit info if trade is closed
	if trade.ExitTime != nil {
		dto.Exit = &dtos.TradeExit{
			Reason:    trade.ExitReason,
			Price:     trade.ExitPrice,
			Timestamp: *trade.ExitTime,
		}
	}

	return dto
}

// BacktestDelete deletes a backtest and its trades
func (s *Services) BacktestDelete(ctx *gin.Context, id uint) (res *dtos.BacktestResponse, err error) {
	res, err = s.BacktestGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		if err := s.repo.BacktestTrade.DeleteByBacktestID(tx, id); err != nil {
			return nil, fmt.Errorf("failed to delete trades: %w", err)
		}
		if _, err := s.repo.Backtest.Delete(tx, id); err != nil {
			return nil, fmt.Errorf("failed to delete backtest: %w", err)
		}
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ============================================================================
// Backtest Execution
// ============================================================================

// BacktestCreate creates a new backtest and runs it in the background
func (s *Services) BacktestCreate(ctx *gin.Context, req *dtos.BacktestRequest) (res *dtos.BacktestResponse, err error) {
	fmt.Println()
	fmt.Println("🚀 [BACKTEST] ═══════════════════════════════════════════")
	fmt.Printf("🚀 [BACKTEST] Creating backtest \"%s\" for %s\n", req.Name, req.Symbol)

	// 1. Load strategy
	strategy, err := s.StrategyGetByID(ctx, req.StrategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load strategy: %w", err)
	}

	// 2. Get MMConfig from strategy (already converted in StrategyGetByID)
	mmConfig := strategy.MoneyManagement
	if mmConfig == nil {
		mmConfig = &dtos.MMConfigResponse{
			MIN_CONFIDENCE:         65,
			MAX_DAILY_TRADES:       10,
			MAX_DAILY_LOSS_PERCENT: 0.05,
			MAX_DAILY_LOSS_COUNT:   5,
			RISK_REWARD_RATIO:      2.0,
			RISK_REWARD_TARGET:     3.0,
			RISK_ENTRY_BUFFER:      0.0075,
			MAX_POSITION_SIZE:      0.15,
			LEVERAGE:               5,
			IS_AGRESSIVE:           false,
			ORDER_EXPIRATION_HOURS: 4,
		}
	}

	// Convert to config.MMConfig for worker
	mmConfigConverted := &config.MMConfig{
		MIN_CONFIDENCE:         mmConfig.MIN_CONFIDENCE,
		MAX_DAILY_TRADES:       mmConfig.MAX_DAILY_TRADES,
		MAX_DAILY_LOSS_PERCENT: mmConfig.MAX_DAILY_LOSS_PERCENT,
		MAX_DAILY_LOSS_COUNT:   mmConfig.MAX_DAILY_LOSS_COUNT,
		RISK_REWARD_RATIO:      mmConfig.RISK_REWARD_RATIO,
		RISK_REWARD_TARGET:     mmConfig.RISK_REWARD_TARGET,
		RISK_ENTRY_BUFFER:      mmConfig.RISK_ENTRY_BUFFER,
		MAX_POSITION_SIZE:      mmConfig.MAX_POSITION_SIZE,
		LEVERAGE:               mmConfig.LEVERAGE,
		IS_AGRESSIVE:           mmConfig.IS_AGRESSIVE,
		ORDER_EXPIRATION_HOURS: mmConfig.ORDER_EXPIRATION_HOURS,
	}

	// 4. Serialize strategy to JSON for snapshot
	strategyJSON, _ := json.Marshal(strategy)

	// 5. Calculate time range
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -req.Days)

	// 6. Create backtest with PENDING status
	now := time.Now()
	backtest := &models.Backtest{
		Name:            req.Name,
		Symbol:          req.Symbol,
		StrategyID:      req.StrategyID,
		StartTime:       startTime,
		EndTime:         endTime,
		Capital:         req.Capital,
		StrategyJSON:    string(strategyJSON),
		EquityCurveJSON: "[]", // Empty JSON array as default
		Status:          "PENDING",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err = s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		backtest, err = s.repo.Backtest.Create(tx, backtest)
		if err != nil {
			return nil, fmt.Errorf("failed to create backtest: %w", err)
		}
		return backtest, nil
	})

	if err != nil {
		return nil, err
	}

	fmt.Printf("💾 [BACKTEST] Backtest created with ID: %d (Status: PENDING)\n", backtest.ID)
	fmt.Println("🔄 [BACKTEST] Starting background worker...")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// 5. Run backtest in background
	go s.backtestRunWorker(backtest.ID, req.Days, strategy, mmConfigConverted, req.Symbol, req.Capital, startTime)

	// 6. Return simple response (without OHLCV)
	return &dtos.BacktestResponse{
		ID:         backtest.ID,
		Name:       backtest.Name,
		Symbol:     backtest.Symbol,
		StrategyID: backtest.StrategyID,
		StartTime:  backtest.StartTime,
		EndTime:    backtest.EndTime,
		Capital:    backtest.Capital,
		Status:     backtest.Status,
		CreatedAt:  backtest.CreatedAt,
	}, nil
}
