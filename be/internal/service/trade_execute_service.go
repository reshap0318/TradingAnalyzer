package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// TradeExecute coordinates the automated trading process
func (s *Services) TradeExecute(ctx *gin.Context, req *dtos.TradeRequest) (*dtos.TradeResponse, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// 1. Get Strategy & Build MM Config
	var strategy *dtos.StrategyData
	var err error
	if req.StrategyID > 0 {
		strategy, err = s.StrategyGetByID(ctx, req.StrategyID)
	} else {
		strategy, err = s.StrategyGetActive(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active strategy: %w", err)
	}
	mmConfig := strategy.MoneyManagement
	minBalance := 3.0

	symStat := s.tradeExecuteTodayStat(req.Symbol)

	// VALIDATION 1: LOCAL HARD LIMIT - Active Trade
	if symStat.Active > 0 {
		return &dtos.TradeResponse{
			Symbol:    req.Symbol,
			Timestamp: time.Now(),
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  fmt.Sprintf("HARD LIMIT: Symbol %s already has an active trade. Only 1 active trade permitted.", req.Symbol),
			},
		}, nil
	}

	// VALIDATION 2A: LOCAL HARD LIMIT - Consecutive Loss
	if mmConfig.MAX_DAILY_LOSS_COUNT > 0 && int(symStat.SLHits) >= int(mmConfig.MAX_DAILY_LOSS_COUNT) {
		return &dtos.TradeResponse{
			Symbol:    req.Symbol,
			Timestamp: time.Now(),
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  fmt.Sprintf("HARD LIMIT: Reached max consecutive loss (%d). Cooling down.", symStat.SLHits),
			},
		}, nil
	}

	// VALIDATION 3: Binance API & Position Sizing
	balanceInfo, err := s.BinanceClient.GetBalance("USDT")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch USDT balance from Binance Futures: %w", err)
	}

	availableUsdt := balanceInfo.AvailableBalance
	totalWalletUsdt := balanceInfo.WalletBalance // or MarginBalance depending on exact accounting preference for PnL

	// Reserve 2% for trading fees (taker/maker fees, funding rates)
	feesReservePercent := 0.02
	availableUsdtWithFeesReserve := availableUsdt * (1 - feesReservePercent)

	// VALIDATION 2B: LOCAL HARD LIMIT - Daily Loss Pct (Rugi > x%)
	// If todayPnL is heavily negative, compare against total wallet.
	if mmConfig.MAX_DAILY_LOSS_PERCENT > 0 && symStat.PnL < 0 {
		lossPctDec := float64(mmConfig.MAX_DAILY_LOSS_PERCENT) // ex: 0.05
		if math.Abs(symStat.PnL) >= (totalWalletUsdt * lossPctDec) {
			return &dtos.TradeResponse{
				Symbol:    req.Symbol,
				Timestamp: time.Now(),
				ExecutionInfo: dtos.ExecutionInfo{
					Executed: false,
					Message:  fmt.Sprintf("HARD LIMIT: Reached max daily loss percentage (Total PnL: %.2f on %.2f Bal). Cooling down.", symStat.PnL, totalWalletUsdt),
				},
			}, nil
		}
	}

	// Validasi dasar: Minimal saldo Binance utuh harus ada minimal untuk bisa testing/trading
	if availableUsdt < minBalance {
		return &dtos.TradeResponse{
			Symbol:    req.Symbol,
			Timestamp: time.Now(),
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  fmt.Sprintf("Insufficient total wallet balance. Requires at least %.1f USDT, but got %.2f USDT", minBalance, availableUsdt),
			},
		}, nil
	}

	// VALIDATION 4: Signal Analyze & Confidence
	// We build an analyze request using available balance AFTER fees reserve (98% of available)
	analyzeReq := &dtos.SignalAnalyzeRequest{
		Symbol:     req.Symbol,
		StrategyID: strategy.ID,
		Capital:    availableUsdtWithFeesReserve,
	}
	
	// Analyze and save signal to database (saveSignal=true)
	analyzeRes, _, err := s.SignalAnalyzeAndSave(ctx, analyzeReq, true)
	if err != nil {
		return nil, fmt.Errorf("signal analysis failed: %w", err)
	}

	if !analyzeRes.Signal.Valid {
		return &dtos.TradeResponse{
			Symbol:           req.Symbol,
			PrimaryTimeframe: analyzeRes.PrimaryTimeframe,
			Timestamp:        time.Now(),
			Signal:           analyzeRes.Signal,
			Scoring:          analyzeRes.Scoring,
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  fmt.Sprintf("Signal invalid: Confidence(%.2f) under threshold(%d)", analyzeRes.Scoring.Confidence, mmConfig.MIN_CONFIDENCE),
			},
		}, nil
	}

	if analyzeRes.Signal.Signal == "WAIT" {
		return &dtos.TradeResponse{
			Symbol:           req.Symbol,
			PrimaryTimeframe: analyzeRes.PrimaryTimeframe,
			Timestamp:        time.Now(),
			Signal:           analyzeRes.Signal,
			Scoring:          analyzeRes.Scoring,
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  "Signal result is WAIT. No trade taken.",
			},
		}, nil
	}

	// VALIDATION 4.5: HARD LIMIT - Minimum Risk/Reward Ratio
	if analyzeRes.Signal.TradingPlan.RiskRewardRatio < float64(mmConfig.RISK_REWARD_RATIO) {
		return &dtos.TradeResponse{
			Symbol:           req.Symbol,
			PrimaryTimeframe: analyzeRes.PrimaryTimeframe,
			Timestamp:        time.Now(),
			Signal:           analyzeRes.Signal,
			Scoring:          analyzeRes.Scoring,
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message: fmt.Sprintf("HARD LIMIT: Rejected because R:R %.2f is lower than minimum allowed %.2f.",
					analyzeRes.Signal.TradingPlan.RiskRewardRatio, mmConfig.RISK_REWARD_RATIO),
			},
		}, nil
	}

	// VALIDATION 5: SOFT LIMIT - Daily Trade Count & RR Ratio TARGET Override
	if symStat.Count >= mmConfig.MAX_DAILY_TRADES {
		isExcellentSetup := analyzeRes.Signal.TradingPlan.RiskRewardRatio >= float64(mmConfig.RISK_REWARD_TARGET)
		if !isExcellentSetup {
			return &dtos.TradeResponse{
				Symbol:           req.Symbol,
				PrimaryTimeframe: analyzeRes.PrimaryTimeframe,
				Timestamp:        time.Now(),
				Signal:           analyzeRes.Signal,
				Scoring:          analyzeRes.Scoring,
				ExecutionInfo: dtos.ExecutionInfo{
					Executed: false,
					Message: fmt.Sprintf("SOFT LIMIT: Max daily trades (%d) reached. Rejected because R:R %.2f is lower than TARGET %.2f for exception.",
						mmConfig.MAX_DAILY_TRADES, analyzeRes.Signal.TradingPlan.RiskRewardRatio, mmConfig.RISK_REWARD_TARGET),
				},
			}, nil
		}
	}

	// Get actual capital used from pre-calculated summary (no need to recalculate)
	actualCapitalUsed := analyzeRes.Signal.TradingPlan.Summary.CapitalUsed

	// FINAL VALIDATION: Safeguard before execution
	if analyzeRes.Signal.TradingPlan.Summary == nil {
		return nil, fmt.Errorf("trading plan summary is nil, cannot execute trade")
	}
	if actualCapitalUsed <= 0 {
		return nil, fmt.Errorf("invalid capital used: %.2f USDT", actualCapitalUsed)
	}

	// If valid, Proceed to Execution Preparation
	return s.tradeExecuteBinance(ctx, req.Symbol, mmConfig, analyzeRes, actualCapitalUsed)
}

// tradeExecuteTodayStat calculates trading statistics for today
func (s *Services) tradeExecuteTodayStat(symbol string) dtos.TradeDayStat {
	stat := dtos.TradeDayStat{}
	countConsecutive := true

	// Step 1: Single DB Query for Today's Trades (Ordered from newest to oldest for Consecutive checks)
	todaysTrades, err := s.repo.Trade.TradeFindToday(nil)
	if err != nil {
		return stat
	}

	for _, tradeLog := range todaysTrades {
		// Active Check specific to Symbol
		if tradeLog.Symbol == symbol && (tradeLog.Status == "ACTIVE" || tradeLog.Status == "PENDING" || tradeLog.Status == "PARTIAL") {
			stat.Active++
		}

		if tradeLog.Status != "CANCELLED" && tradeLog.Status != "REJECTED" {
			stat.Count++
		}

		// Calculate PnL and Consecutive Losses (Only for CLOSED/FINISHED signals)
		if tradeLog.Status == "CLOSED" || tradeLog.Status == "FINISHED" {
			stat.PnL += tradeLog.PnL // Adding PnL (Negative for losses, Positive for wins)

			if tradeLog.PnL < 0 {
				stat.SLHits++
				stat.TotalProfit += tradeLog.PnL
			} else if tradeLog.PnL > 0 {
				stat.TPHits++
				stat.TotalLoss += tradeLog.PnL
			}

			// Consecutive loss logic: look at most recent trades. As soon as we see a win or break-even, stop counting.
			if countConsecutive {
				if tradeLog.PnL < 0 {
					stat.ConsecutiveLossess++
				} else if tradeLog.PnL > 0 {
					countConsecutive = false
				}
			}
		}
	}

	return stat
}

// tradeExecuteBinance handles the actual external API calling for order execution
func (s *Services) tradeExecuteBinance(
	ctx *gin.Context,
	symbol string,
	config *dtos.MMConfigResponse,
	analyzeRes *dtos.SignalAnalyzeResponse,
	capitalUsed float64,
) (*dtos.TradeResponse, error) {
	tpPlan := analyzeRes.Signal.TradingPlan
	side := binance.OrderSideBuy // default BUY

	if analyzeRes.Signal.Signal == "SELL" || analyzeRes.Signal.Signal == "STRONG_SELL" {
		side = binance.OrderSideSell
	}

	// 1. Get Symbol Filter / Exchange Info from Binance
	symbolInfo, err := s.BinanceClient.GetSymbolInfo(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch symbol info %s: %w", symbol, err)
	}

	// 2. Setup Margin Type (via Redis Cache wrapper)
	// Set to ISOLATED margin mode for safer risk management
	_, err = s.BinanceClient.SetMarginMode(&binance.MarginModeRequest{
		Symbol:     symbol,
		MarginMode: 1, // 1 = ISOLATED, 2 = CROSSED
	})
	if err != nil {
		// Check if error is "No need to change margin type" (Binance error code -4046)
		// If so, we can safely ignore it and continue
		errMsg := err.Error()
		if !strings.Contains(errMsg, "-4046") && !strings.Contains(errMsg, "No need to change margin type") {
			return nil, fmt.Errorf("failed to set margin mode to ISOLATED: %w", err)
		}
		// Error ignored successfully
	}

	// 3. Setup Leverage
	_, err = s.BinanceClient.SetLeverage(&binance.LeverageRequest{
		Symbol:   symbol,
		Leverage: int(config.LEVERAGE),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set leverage: %w", err)
	}

	// 5. Execute Order Entry Loop
	var executedOrders []dtos.OrderInfo
	var totalFilledQty float64
	var avgEntryPriceSum float64 // Will need accurate sum based on order quantities

	for _, entry := range tpPlan.Entries {
		adjustedPrice := binance.AdjustPricePrecision(entry.EntryPrice, symbolInfo.TickSize)
		adjustedQty := binance.AdjustQuantityPrecision(entry.PositionQty, symbolInfo.StepSize)

		if adjustedQty <= 0 {
			continue // Skip if qty is too small
		}

		// First entry uses Market typically if it's close to current price, but the TradingPlan creates Limit values.
		// For conservative we use Limit. For aggressive, 1st is usually MARKET (or very close Limit), 2nd is limit.
		orderType := binance.OrderTypeLimit
		if entry.EntryNumber == 1 && tpPlan.Mode == "AGGRESSIVE" {
			orderType = binance.OrderTypeMarket
		}

		reqOrder := &binance.PlaceOrderRequest{
			Symbol:   symbol,
			Side:     side,
			Type:     orderType,
			Quantity: adjustedQty,
		}

		if orderType == binance.OrderTypeLimit {
			reqOrder.Price = adjustedPrice
			reqOrder.TimeInForce = "GTC"
		}

		orderResponse, err := s.BinanceClient.PlaceOrder(reqOrder)
		if err != nil {
			// If first order fails, we should abort
			if entry.EntryNumber == 1 {
				return nil, fmt.Errorf("failed to place entry order #%d: %w", entry.EntryNumber, err)
			}
			continue
		}

		executedMsg := dtos.OrderInfo{
			EntryNumber:    entry.EntryNumber,
			BinanceOrderID: orderResponse.OrderID,
			Price:          orderResponse.Price,
			Quantity:       orderResponse.OrigQuantity,
			Type:           orderResponse.Type,
			Status:         orderResponse.Status,
		}
		executedOrders = append(executedOrders, executedMsg)

		// Check if order is FILLED (immediately for market orders, or check status)
		// For LIMIT orders, status might be NEW (pending) or FILLED (if instant match)
		isFilled := orderResponse.Status == "FILLED" || orderResponse.Status == "PARTIALLY_FILLED"

		if isFilled && orderResponse.ExecutedQuantity > 0 {
			filledQty := orderResponse.ExecutedQuantity
			filledPrice := orderResponse.AveragePrice
			if filledPrice == 0 {
				filledPrice = orderResponse.Price
			}

			totalFilledQty += filledQty
			avgEntryPriceSum += (filledPrice * filledQty)
		}
	}

	// 6. Execute TP/SL Orders ONLY if there are FILLED entries
	var tpOrderID, slOrderID int64
	if totalFilledQty > 0 {
		var closeSide binance.OrderSide
		if side == binance.OrderSideBuy {
			closeSide = binance.OrderSideSell
		} else {
			closeSide = binance.OrderSideBuy
		}

		tpAdjusted := binance.AdjustPricePrecision(tpPlan.TakeProfit, symbolInfo.TickSize)
		slAdjusted := binance.AdjustPricePrecision(tpPlan.StopLoss, symbolInfo.TickSize)

		// Place Take Profit Market via Algo API
		tpReq := &binance.PlaceAlgoOrderRequest{
			Symbol:        symbol,
			Side:          closeSide,
			Type:          binance.OrderTypeTakeProfitMarket,
			TriggerPrice:  tpAdjusted,
			ClosePosition: true,
		}
		tpResp, err := s.BinanceClient.PlaceAlgoOrder(ctx, tpReq)
		if err == nil && tpResp.AlgoID > 0 {
			tpOrderID = tpResp.AlgoID // Capture actual AlgoID from response
		} else if err != nil {
			fmt.Printf("Warning: Failed to place TP for %s: %v\n", symbol, err)
		}

		// Place Stop Loss Market via Algo API
		slReq := &binance.PlaceAlgoOrderRequest{
			Symbol:        symbol,
			Side:          closeSide,
			Type:          binance.OrderTypeStopMarket,
			TriggerPrice:  slAdjusted,
			ClosePosition: true,
		}
		slResp, err := s.BinanceClient.PlaceAlgoOrder(ctx, slReq)
		if err == nil && slResp.AlgoID > 0 {
			slOrderID = slResp.AlgoID // Capture actual AlgoID from response
		} else if err != nil {
			fmt.Printf("Warning: Failed to place SL for %s: %v\n", symbol, err)
		}
	} else {
		// No entries filled yet - TP/SL will be placed later by background job or manual check
		// This is expected for LIMIT orders that are still PENDING
		fmt.Printf("Info: No entries filled yet for %s. TP/SL will be placed after fill.\n", symbol)
	}

	// 7. Save To Database using Transaction
	avgEntryPrice := 0.0
	if totalFilledQty > 0 {
		avgEntryPrice = avgEntryPriceSum / totalFilledQty
	}

	err = s.tradeExecuteSaveRecord(symbol, side, tpPlan, analyzeRes, capitalUsed, float64(config.LEVERAGE), executedOrders, tpOrderID, slOrderID, avgEntryPrice, totalFilledQty)
	if err != nil {
		fmt.Printf("Warning: Trade executed but DB tracking failed: %v", err)
	}

	return &dtos.TradeResponse{
		Symbol:           symbol,
		PrimaryTimeframe: analyzeRes.PrimaryTimeframe,
		Timestamp:        time.Now(),
		Signal:           analyzeRes.Signal,
		Scoring:          analyzeRes.Scoring,
		ExecutionInfo: dtos.ExecutionInfo{
			Executed:    true,
			Message:     "Trade successfully executed",
			MarginType:  "ISOLATED",
			Leverage:    int(config.LEVERAGE),
			CapitalUsed: capitalUsed,
			Orders:      executedOrders,
			TPOrderID:   tpOrderID,
			SLOrderID:   slOrderID,
		},
	}, nil
}

// tradeExecuteSaveRecord saves the completed transaction to your DB
func (s *Services) tradeExecuteSaveRecord(
	symbol string,
	side binance.OrderSide,
	tpPlan *dtos.TradingPlan,
	analyzeRes *dtos.SignalAnalyzeResponse,
	capitalUsed float64,
	leverage float64,
	executedOrders []dtos.OrderInfo,
	tpOrderID, slOrderID int64,
	avgEntryPrice float64,
	totalQty float64,
) error {
	_, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
		// Calculate Risk Reward explicitly
		rrRatio := tpPlan.RiskRewardRatio
		if rrRatio == 0 {
			diffTP := math.Abs(tpPlan.TakeProfit - avgEntryPrice)
			diffSL := math.Abs(avgEntryPrice - tpPlan.StopLoss)
			if diffSL > 0 {
				rrRatio = float64(diffTP / diffSL)
			}
		}

		// Convert SignalAnalyzeResponse to JSONMap for RawSignal
		rawSignal := s.convertAnalyzeResToJSONMap(analyzeRes)

		// Save Parent Trade
		parentTrade := &models.Trade{
			Symbol:          symbol,
			Interval:        analyzeRes.PrimaryTimeframe,
			Side:            string(side),
			Confidence:      analyzeRes.Scoring.Confidence,
			TotalScore:      analyzeRes.Scoring.TotalScore,
			RawSignal:       rawSignal,
			IsAggressive:    tpPlan.Mode == "AGGRESSIVE",
			TPPrice:         tpPlan.TakeProfit,
			SLPrice:         tpPlan.StopLoss,
			RiskRewardRatio: rrRatio,
			AvgEntryPrice:   avgEntryPrice,
			Leverage:        int(leverage),
			CapitalUsed:     capitalUsed,
			TotalQty:        totalQty,
			Status:          "ACTIVE",
			TPOrderID:       tpOrderID,
			SLOrderID:       slOrderID,
		}

		savedTrade, err := s.repo.Trade.Create(tx, parentTrade)
		if err != nil {
			return nil, err
		}

		// Save Trade Entries
		for _, eo := range executedOrders {
			// Find original planned position sizing info
			posSizeStr := ""
			posValue := 0.0
			for _, planned := range tpPlan.Entries {
				if planned.EntryNumber == eo.EntryNumber {
					posSizeStr = planned.PositionSize
					posValue = planned.PositionValue
					break
				}
			}

			entryStat := "PENDING"
			if eo.Status == "FILLED" {
				entryStat = "FILLED"
			}

			entry := &models.TradeEntry{
				TradeID:        savedTrade.ID,
				EntryNumber:    eo.EntryNumber,
				EntryPrice:     eo.Price,
				EntryType:      eo.Type, // LIMIT / MARKET
				PositionSize:   posSizeStr,
				PositionValue:  posValue,
				PositionQty:    eo.Quantity,
				BinanceOrderID: eo.BinanceOrderID,
				BinanceStatus:  eo.Status,
				Status:         entryStat,
			}

			if entryStat == "FILLED" {
				now := time.Now()
				entry.FilledPrice = eo.Price
				entry.FilledQty = eo.Quantity
				entry.FilledAt = &now
			}

			_, err = s.repo.TradeEntry.Create(tx, entry)
			if err != nil {
				return nil, err
			}
		}
		return savedTrade, nil
	})

	return err
}

// convertAnalyzeResToJSONMap converts SignalAnalyzeResponse to JSONMap for database storage
func (s *Services) convertAnalyzeResToJSONMap(analyzeRes *dtos.SignalAnalyzeResponse) models.JSONMap {
	// Build final JSON map containing the whole analysis result
	// The struct is marshaled then unmarshaled to generic map type
	b, err := json.Marshal(analyzeRes)
	if err != nil {
		fmt.Printf("Warning: failed to marshal analyze result: %v\n", err)
		return models.JSONMap{}
	}

	var jsonMap models.JSONMap
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		fmt.Printf("Warning: failed to unmarshal analyze result to JSONMap: %v\n", err)
		return models.JSONMap{}
	}

	return jsonMap
}
