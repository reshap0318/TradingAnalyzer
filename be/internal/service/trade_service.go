package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/config"
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
	mmConfig := s.getConfigMM(strategy)

	// Step 1: Single DB Query for Today's Trades (Ordered from newest to oldest for Consecutive checks)
	todaysTrades, err := s.repo.Trade.TradeFindToday(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch today's trading history: %w", err)
	}

	// Step 2: Extract Metrics from DB Historical Data
	activeCount := 0
	validTradeCountToday := 0
	consecutiveLosses := 0
	countConsecutive := true
	minBalance := 3.0
	var todayPnL float64

	for _, tradeLog := range todaysTrades {
		// Active Check specific to Symbol
		if tradeLog.Symbol == req.Symbol && (tradeLog.Status == "ACTIVE" || tradeLog.Status == "PENDING" || tradeLog.Status == "PARTIAL") {
			activeCount++
		}

		if tradeLog.Status != "CANCELLED" && tradeLog.Status != "REJECTED" {
			validTradeCountToday++
		}

		// Calculate PnL and Consecutive Losses (Only for CLOSED/FINISHED signals)
		// Assume "CLOSED" status implies completed trades with PnL calculation
		if tradeLog.Status == "CLOSED" || tradeLog.Status == "FINISHED" {
			todayPnL += tradeLog.PnL // Adding PnL (Negative for losses, Positive for wins)

			// Consecutive loss logic: look at most recent trades. As soon as we see a win or break-even, stop counting.
			if countConsecutive {
				if tradeLog.PnL < 0 {
					consecutiveLosses++
				} else if tradeLog.PnL > 0 {
					countConsecutive = false
				}
			}
		}
	}

	// VALIDATION 1: LOCAL HARD LIMIT - Active Trade
	if activeCount > 0 {
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
	if mmConfig.MAX_DAILY_LOSS_COUNT > 0 && consecutiveLosses >= int(mmConfig.MAX_DAILY_LOSS_COUNT) {
		return &dtos.TradeResponse{
			Symbol:    req.Symbol,
			Timestamp: time.Now(),
			ExecutionInfo: dtos.ExecutionInfo{
				Executed: false,
				Message:  fmt.Sprintf("HARD LIMIT: Reached max consecutive loss (%d). Cooling down.", consecutiveLosses),
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

	// VALIDATION 2B: LOCAL HARD LIMIT - Daily Loss Pct (Rugi > x%)
	// If todayPnL is heavily negative, compare against total wallet.
	if mmConfig.MAX_DAILY_LOSS_PERCENT > 0 && todayPnL < 0 {
		lossPctDec := float64(mmConfig.MAX_DAILY_LOSS_PERCENT) // ex: 0.05
		if math.Abs(todayPnL) >= (totalWalletUsdt * lossPctDec) {
			return &dtos.TradeResponse{
				Symbol:    req.Symbol,
				Timestamp: time.Now(),
				ExecutionInfo: dtos.ExecutionInfo{
					Executed: false,
					Message:  fmt.Sprintf("HARD LIMIT: Reached max daily loss percentage (Total PnL: %.2f on %.2f Bal). Cooling down.", todayPnL, totalWalletUsdt),
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
	// We build an analyze request using our accurate `tradCapital`
	analyzeReq := &dtos.SignalAnalyzeRequest{
		Symbol:     req.Symbol,
		StrategyID: strategy.ID,
		Capital:    availableUsdt,
	}
	analyzeRes, err := s.SignalAnalyze(ctx, analyzeReq)
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
				Message:  fmt.Sprintf("Signal invalid. Confidence %.2f is lower than MIN_CONFIDENCE (%d)", analyzeRes.Scoring.Confidence, mmConfig.MIN_CONFIDENCE),
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

	// Hitung total capital yang digunakan langsung dari TradingPlan (tanpa modifikasi)
	actualCapitalUsed := 0.0
	for _, entry := range analyzeRes.Signal.TradingPlan.Entries {
		actualCapitalUsed += entry.PositionValue
	}

	// VALIDATION 5: SOFT LIMIT - Daily Trade Count & RR Ratio TARGET Override
	if validTradeCountToday >= int(mmConfig.MAX_DAILY_TRADES) {
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

	// If valid, Proceed to Execution Preparation
	return s.executeBinanceTrade(ctx, req.Symbol, strategy, mmConfig, analyzeRes, actualCapitalUsed)
}

// executeBinanceTrade handles the actual external API calling for order execution
func (s *Services) executeBinanceTrade(
	ctx *gin.Context,
	symbol string,
	strategy *dtos.StrategyData,
	config *config.MMConfig,
	analyzeRes *dtos.SignalAnalyzeResponse,
	capitalUsed float64,
) (*dtos.TradeResponse, error) {
	tpPlan := analyzeRes.Signal.TradingPlan
	currentPrice := analyzeRes.Signal.CurentPrice
	side := binance.OrderSideBuy // default BUY

	if analyzeRes.Signal.Signal == "SELL" || analyzeRes.Signal.Signal == "STRONG_SELL" {
		side = binance.OrderSideSell
	}

	// 1. Get Symbol Filter / Exchange Info from Binance
	symbolInfo, err := s.BinanceClient.GetSymbolInfo(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch symbol info %s: %w", symbol, err)
	}

	// 2. Setup Position Side (Optional, handling Dual-Mode if configured)
	// Some accounts use one-way mode, some use hedge-mode. Assuming default for now.

	// 3. Setup Margin Type (via Redis Cache wrapper)
	// We call SetMarginMode. E.g 1 for ISOLATED. (Assuming 1 is ISOLATED, 0/2 is CROSS based on your binance code)
	_, err = s.BinanceClient.SetMarginMode(&binance.MarginModeRequest{
		Symbol:     symbol,
		MarginMode: 1, // ISOLATED (Hardcoded for safest approach, or parameterize later)
	})
	if err != nil {
		// Ignore error if it says "No need to change margin type" (usually returns -4046 error code)
		// But fail if really an error. For now we continue.
	}

	// 4. Setup Leverage
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
		totalFilledQty += orderResponse.OrigQuantity

		// Approx entry price matching (in reality you wait for Fill/Update hook, but here we assume LIMIT price config)
		ePrice := orderResponse.Price
		if orderType == binance.OrderTypeMarket {
			ePrice = currentPrice // rough estimation if API response doesn't give AvgPrice directly during Market open
		}
		if orderResponse.AveragePrice > 0 {
			ePrice = orderResponse.AveragePrice
		}
		avgEntryPriceSum += (ePrice * orderResponse.OrigQuantity)
	}

	// 6. Execute TP/SL Orders (Only if there is Filled/Pending Qty)
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

		// Place Take Profit Market
		tpReq := &binance.PlaceOrderRequest{
			Symbol:    symbol,
			Side:      closeSide,
			Type:      binance.OrderTypeTakeProfitMarket,
			StopPrice: tpAdjusted,
			// ClosePosition in go-binance uses ReduceOnly or TimeInForce
			// Since PlaceOrderRequest struct doesn't have ClosePosition field right now, we use ReduceOnly
			ReduceOnly: true,
		}
		tpResp, err := s.BinanceClient.PlaceOrder(tpReq)
		if err == nil {
			tpOrderID = tpResp.OrderID
		}

		// Place Stop Loss Market
		slReq := &binance.PlaceOrderRequest{
			Symbol:     symbol,
			Side:       closeSide,
			Type:       binance.OrderTypeStopMarket,
			StopPrice:  slAdjusted,
			ReduceOnly: true,
		}
		slResp, err := s.BinanceClient.PlaceOrder(slReq)
		if err == nil {
			slOrderID = slResp.OrderID
		}
	}

	// 7. Save To Database using Transaction
	avgEntryPrice := 0.0
	if totalFilledQty > 0 {
		avgEntryPrice = avgEntryPriceSum / totalFilledQty
	}

	err = s.saveTradeRecord(ctx, symbol, side, tpPlan, analyzeRes, capitalUsed, float64(config.LEVERAGE), executedOrders, tpOrderID, slOrderID, avgEntryPrice, totalFilledQty)
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

// saveTradeRecord saves the completed transaction to your DB
func (s *Services) saveTradeRecord(
	ctx context.Context,
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

		// Save Parent Trade
		parentTrade := &models.Trade{
			Symbol:          symbol,
			Interval:        tpPlan.Mode,
			Side:            string(side),
			Confidence:      analyzeRes.Scoring.Confidence,
			TotalScore:      analyzeRes.Scoring.TotalScore,
			RawSignal:       nil, // Assign JSON map properly later
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
			if eo.Status == "NEW" {
				entryStat = "PENDING" // In limit orders
			} else if eo.Status == "FILLED" {
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
