package service

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reshap/trading-bot/internal/clients/binance"
	"github.com/reshap/trading-bot/internal/dtos"
	"github.com/reshap/trading-bot/internal/models"
	"gorm.io/gorm"
)

// TradeMonitorProcessAllActive processes all active trades (called by cron job every 1 minute)
func (s *Services) TradeMonitorProcessAllActive(ctx *gin.Context) ([]dtos.ProcessTradeResult, error) {
	// Get all active trades with entries
	trades, err := s.repo.Trade.FindAllActiveWithEntries(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active trades: %w", err)
	}

	results := make([]dtos.ProcessTradeResult, 0, len(trades))

	// Process each trade
	for i := range trades {
		result, err := s.tradeMonitorProcessTrade(ctx, &trades[i])
		if err != nil {
			// Log error but continue processing other trades
			fmt.Printf("Error processing trade %d (%s): %v\n", trades[i].ID, trades[i].Symbol, err)
			results = append(results, dtos.ProcessTradeResult{
				TradeID: trades[i].ID,
				Symbol:  trades[i].Symbol,
				Status:  "ERROR",
				Message: err.Error(),
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// tradeMonitorProcessTrade processes a single active trade (private function)
// This is the main function called for each trade
func (s *Services) tradeMonitorProcessTrade(ctx *gin.Context, trade *models.Trade) (*dtos.ProcessTradeResult, error) {
	result := &dtos.ProcessTradeResult{
		TradeID: trade.ID,
		Symbol:  trade.Symbol,
	}

	// ========================================================================
	// FASE 0: PERSIAPAN DATA
	// ========================================================================

	// 1. Validasi: Jika status Trade bukan "ACTIVE", langsung RETURN
	if trade.Status != "ACTIVE" {
		result.Status = "SKIPPED"
		result.Message = fmt.Sprintf("Trade status is %s, not ACTIVE", trade.Status)
		return result, nil
	}

	// 2. Tarik Binance: GET All Open Orders untuk symbol ini (cache untuk cek di bawah)
	openOrders, err := s.BinanceClient.GetOpenOrders(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	// Build map for quick lookup: order_id -> order
	orderMap := make(map[int64]*binance.OrderResponse, len(openOrders))
	for i := range openOrders {
		orderMap[openOrders[i].OrderID] = &openOrders[i]
	}

	// ========================================================================
	// FASE 1: CEK TP / SL (Prioritas Utama Pencegah Ghost Order)
	// ========================================================================
	fase1Result, shouldReturn, err := s.tradeMonitorFase1CheckTPSL(ctx, trade, orderMap)
	if err != nil {
		return nil, fmt.Errorf("fase 1 failed: %w", err)
	}

	if fase1Result.TPUpdated {
		result.TPUpdated = true
	}
	if fase1Result.SLUpdated {
		result.SLUpdated = true
	}

	if shouldReturn {
		// TP/SL hit - trade sudah close
		result.Status = trade.Status // TP_HIT atau SL_HIT
		result.Message = "TP/SL hit, trade closed"
		return result, nil
	}

	// ========================================================================
	// FASE 2: SINKRONISASI JARING / ENTRY
	// ========================================================================
	fase2Result, err := s.tradeMonitorFase2SyncEntries(ctx, trade, orderMap)
	if err != nil {
		return nil, fmt.Errorf("fase 2 failed: %w", err)
	}

	result.EntriesSync = fase2Result.UpdatedCount
	if fase2Result.TPUpdated {
		result.TPUpdated = true
	}
	if fase2Result.SLUpdated {
		result.SLUpdated = true
	}

	// ========================================================================
	// FASE 3: NETTING & FINALISASI
	// ========================================================================
	err = s.tradeMonitorFase3Netting(ctx, trade)
	if err != nil {
		return nil, fmt.Errorf("fase 3 failed: %w", err)
	}

	// Reload trade to get final status
	trade, err = s.repo.Trade.FindWithEntries(nil, trade.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload trade: %w", err)
	}

	result.Status = trade.Status
	result.Message = "Trade processed successfully"

	return result, nil
}

// tradeMonitorFase1CheckTPSL handles Fase 1: Cek TP/SL
func (s *Services) tradeMonitorFase1CheckTPSL(
	ctx *gin.Context,
	trade *models.Trade,
	orderMap map[int64]*binance.OrderResponse,
) (*dtos.ProcessTradeResult, bool, error) {
	result := &dtos.ProcessTradeResult{}
	shouldReturn := false

	// Cek DB: Apakah tp_order_id sudah ada isinya?
	if trade.TPOrderID == 0 {
		return result, false, nil
	}

	// Cocokkan ID TP/SL tersebut dengan data cache Binance
	tpOrder, tpExists := orderMap[trade.TPOrderID]
	slOrder, slExists := orderMap[trade.SLOrderID]

	// 🚨 PENTING: GetOpenOrders HANYA mengembalikan order yang NEW/PARTIALLY_FILLED.
	// Jika status TP atau SL sudah "FILLED", order akan hilang dari orderMap!
	// Oleh karena itu, jika tidak ada di orderMap, kita wajib hit GET Order Detail.
	if !tpExists && trade.TPOrderID > 0 {
		tpOrderDetail, err := s.BinanceClient.GetOrder(&binance.GetOrdersRequest{
			Symbol:  trade.Symbol,
			OrderID: trade.TPOrderID,
		})
		if err == nil {
			tpOrder = tpOrderDetail
			tpExists = true
		}
	}

	if !slExists && trade.SLOrderID > 0 {
		slOrderDetail, err := s.BinanceClient.GetOrder(&binance.GetOrdersRequest{
			Symbol:  trade.Symbol,
			OrderID: trade.SLOrderID,
		})
		if err == nil {
			slOrder = slOrderDetail
			slExists = true
		}
	}

	tpFilled := tpExists && tpOrder.Status == "FILLED"
	slFilled := slExists && slOrder.Status == "FILLED"

	// Jika Status Binance = "FILLED"
	if tpFilled || slFilled {
		// 1. Update DB Trade.status = "TP_HIT" (atau "SL_HIT")
		// 2. Update DB Trade.closed_at = waktu sekarang
		// 3. 🚨 CRITICAL: Hit API Binance untuk CANCEL SEMUA order jaring (Entry) yang masih ngantre ("NEW")
		// 4. Update DB status Entry yang di-cancel tadi jadi "CANCELLED"
		// 5. 🛑 RETURN

		err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
			// Determine exit reason and status
			exitStatus := "TP_HIT"
			if slFilled {
				exitStatus = "SL_HIT"
			}

			// Update trade status
			now := time.Now()
			updateTrade := &models.Trade{
				Status:     exitStatus,
				ClosedAt:   &now,
				ExitPrice:  trade.TPPrice, // Will be updated later with actual fill
				ExitReason: func() string {
					if slFilled {
						return "SL_HIT"
					}
					return "TP_HIT"
				}(),
			}
			_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
			if err != nil {
				return fmt.Errorf("failed to update trade status: %w", err)
			}

			// 🚨 CRITICAL: Cancel semua order jaring (Entry) yang masih ngantre ("NEW")
			// Get all entries for this trade
			entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch entries: %w", err)
			}

			// Cancel pending entries
			for _, entry := range entries {
				if entry.Status == "PENDING" || entry.Status == "NEW" {
					// Cancel order on Binance
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: entry.BinanceOrderID,
					})

					// Ignore error if order already filled/cancelled
					if err != nil {
						fmt.Printf("Warning: Failed to cancel entry order %d: %v\n", entry.BinanceOrderID, err)
					}

					// Update DB status Entry jadi "CANCELLED"
					err = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "TP/SL hit")
					if err != nil {
						return fmt.Errorf("failed to update entry status: %w", err)
					}
				}
			}

			return nil
		})

		if err != nil {
			return result, false, err
		}

		// Update trade status in memory
		trade.Status = "TP_HIT"
		if slFilled {
			trade.Status = "SL_HIT"
		}

		result.TPUpdated = tpFilled
		result.SLUpdated = slFilled
		shouldReturn = true
	}

	return result, shouldReturn, nil
}

// tradeMonitorFase2SyncEntries handles Fase 2: Sinkronisasi Entry
func (s *Services) tradeMonitorFase2SyncEntries(
	ctx *gin.Context,
	trade *models.Trade,
	orderMap map[int64]*binance.OrderResponse,
) (*dtos.ProcessTradeResult, error) {
	result := &dtos.ProcessTradeResult{}
	updatedCount := 0
	tpUpdated := false
	slUpdated := false

	// Looping untuk setiap Entry yang ada di dalam Trade ini
	for i := range trade.Entries {
		entry := &trade.Entries[i]

		// Cocokkan ID Entry dengan data cache Binance
		binanceOrder, exists := orderMap[entry.BinanceOrderID]

		if !exists {
			// Order tidak ditemukan di Binance - mungkin sudah filled dan hilang dari open orders
			// atau cancelled. Cek status terakhir.
			if entry.Status == "PENDING" || entry.Status == "NEW" || entry.Status == "PARTIALLY_FILLED" {
				// Order tidak ada di open orders dan status belum final - kemungkinan sudah filled
				// atau cancelled. Kita perlu fetch order details secara spesifik.
				orderDetail, err := s.BinanceClient.GetOrder(&binance.GetOrdersRequest{
					Symbol:  trade.Symbol,
					OrderID: entry.BinanceOrderID,
				})

				if err != nil {
					// Order tidak ditemukan/gagal di-fetch - anggap cancelled HANYA jika belum punya barang
					if entry.Status == "PENDING" || entry.Status == "NEW" {
						entry.Status = "CANCELLED"
						updatedCount++
					}
					continue
				}

				binanceOrder = orderDetail
				exists = true
			} else {
				// Entry sudah filled/cancelled sebelumnya secara permanen, skip
				continue
			}
		}

		if !exists {
			continue
		}

		// ========================================================================
		// KONDISI A: Status Binance = "NEW" (Masih Ngantre)
		// ========================================================================
		if binanceOrder.Status == "NEW" {
			// Cek usia order vs hour_expired_config
			orderAge := time.Since(entry.CreatedAt)
			expirationHours := s.cfg.MM.ORDER_EXPIRATION_HOURS
			expirationDuration := time.Duration(expirationHours) * time.Hour

			if orderAge > expirationDuration {
				// Expired - Cancel order
				_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
					Symbol:  trade.Symbol,
					OrderID: entry.BinanceOrderID,
				})

				if err != nil {
					fmt.Printf("Warning: Failed to cancel expired order %d: %v\n", entry.BinanceOrderID, err)
				}

				// Update DB status = "CANCELLED"
				err = s.repo.TradeEntry.UpdateStatus(nil, entry.ID, "CANCELLED", "Expired")
				if err != nil {
					return result, fmt.Errorf("failed to update entry status: %w", err)
				}

				entry.Status = "CANCELLED"
				updatedCount++
			}
			// Jika belum expired, abaikan
		}

		// ========================================================================
		// KONDISI B: Status Binance = "PARTIALLY_FILLED"
		// ========================================================================
		if binanceOrder.Status == "PARTIALLY_FILLED" {
			// Update DB status = "PARTIALLY_FILLED"
			// Update DB filled_qty sesuai data Binance yang baru
			_, err := s.repo.TradeEntry.UpdateFilled(nil, entry.ID,
				binanceOrder.AveragePrice,
				binanceOrder.ExecutedQuantity,
				binanceOrder.OrderID,
				"PARTIALLY_FILLED",
			)
			if err != nil {
				return result, fmt.Errorf("failed to update partially filled entry: %w", err)
			}

			entry.Status = "PARTIALLY_FILLED"
			entry.FilledQty = binanceOrder.ExecutedQuantity
			entry.FilledPrice = binanceOrder.AveragePrice
			updatedCount++
		}

		// ========================================================================
		// KONDISI C: Status Binance = "FILLED" (Dapet Barang)
		// ========================================================================
		if binanceOrder.Status == "FILLED" {
			// Update DB status = "FILLED"
			// Update DB filled_qty, filled_price, dan filled_at
			now := time.Now()
			_, err := s.repo.TradeEntry.UpdateFilled(nil, entry.ID,
				binanceOrder.AveragePrice,
				binanceOrder.ExecutedQuantity,
				binanceOrder.OrderID,
				"FILLED",
			)
			if err != nil {
				return result, fmt.Errorf("failed to update filled entry: %w", err)
			}

			entry.Status = "FILLED"
			entry.FilledQty = binanceOrder.ExecutedQuantity
			entry.FilledPrice = binanceOrder.AveragePrice
			entry.FilledAt = &now
			updatedCount++

			// ================================================================
			// [Logic Pasang/Update TP & SL]
			// ================================================================

			// Hitung total qty dari semua entry yang filled/partially filled
			totalFilledQty := 0.0
			for _, e := range trade.Entries {
				if e.Status == "FILLED" || e.Status == "PARTIALLY_FILLED" {
					totalFilledQty += e.FilledQty
				}
			}

			// Skenario 1: Jika DB belum punya tp_order_id (Ini Entry Pertama)
			if trade.TPOrderID == 0 && totalFilledQty > 0 {
				// Hit API Binance CREATE order TP & SL sesuai qty yang didapat
				tpOrderID, slOrderID, err := s.tradeMonitorCreateTPOrder(trade, totalFilledQty)
				if err != nil {
					return result, fmt.Errorf("failed to create TP/SL orders: %w", err)
				}

				// Simpan ID TP/SL barunya ke DB
				updateTrade := &models.Trade{
					TPOrderID:   tpOrderID,
					SLOrderID:   slOrderID,
					TotalQty:    totalFilledQty,
					CapitalUsed: totalFilledQty * entry.FilledPrice, // Approximate
				}
				_, err = s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, updateTrade)
				if err != nil {
					return result, fmt.Errorf("failed to update trade with TP/SL IDs: %w", err)
				}

				trade.TPOrderID = tpOrderID
				trade.SLOrderID = slOrderID
				tpUpdated = true
				slUpdated = true
			} else if totalFilledQty > 0 {
				// Skenario 2: Ini Entry Averaging (DB sudah punya tp_order_id)
				// 1. Hitung TotalQtyBaru (Jumlah semua koin dari entry yang Filled/Partially)
				// 2. Hit API Binance CANCEL order TP/SL lama (Gratis)
				// 3. Hit API Binance CREATE order TP/SL baru dengan TotalQtyBaru (Harga target TP/SL tetap)
				// 4. Update DB timpa ID TP/SL lama dengan ID yang baru

				// Cancel TP/SL lama
				if trade.TPOrderID > 0 {
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: trade.TPOrderID,
					})
					if err != nil {
						fmt.Printf("Warning: Failed to cancel old TP order %d: %v\n", trade.TPOrderID, err)
					}
				}

				if trade.SLOrderID > 0 {
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: trade.SLOrderID,
					})
					if err != nil {
						fmt.Printf("Warning: Failed to cancel old SL order %d: %v\n", trade.SLOrderID, err)
					}
				}

				// Create TP/SL baru
				tpOrderID, slOrderID, err := s.tradeMonitorCreateTPOrder(trade, totalFilledQty)
				if err != nil {
					// 🚨 JIKA GAGAL CREATE, TRADE TIDAK PUNYA TP/SL DI BINANCE SAMA SEKALI
					// Kita harus mengosongkan TPOrderID & SLOrderID di Database agar
					// di putaran cron berikutnya bot ini men-trigger "Skenario 1" dan mencoba membuat TP/SL ulang.
					fallbackUpdate := &models.Trade{
						TPOrderID: 0,
						SLOrderID: 0,
					}
					// Update DB dengan error handling yang aman
					s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, fallbackUpdate)

					trade.TPOrderID = 0
					trade.SLOrderID = 0

					return result, fmt.Errorf("failed to create new TP/SL orders (IDs cleared in DB for safety): %w", err)
				}

				// Update DB timpa ID TP/SL lama dengan ID yang baru
				updateTrade := &models.Trade{
					TPOrderID: tpOrderID,
					SLOrderID: slOrderID,
					TotalQty:  totalFilledQty,
				}
				_, err = s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, updateTrade)
				if err != nil {
					return result, fmt.Errorf("failed to update trade with new TP/SL IDs: %w", err)
				}

				trade.TPOrderID = tpOrderID
				trade.SLOrderID = slOrderID
				tpUpdated = true
				slUpdated = true
			}
		}
	}

	result.UpdatedCount = updatedCount
	result.TPUpdated = tpUpdated
	result.SLUpdated = slUpdated

	return result, nil
}

// tradeMonitorCreateTPOrder creates TP and SL orders for a trade
func (s *Services) tradeMonitorCreateTPOrder(trade *models.Trade, totalQty float64) (int64, int64, error) {
	// Get symbol info for precision adjustment
	symbolInfo, err := s.BinanceClient.GetSymbolInfo(trade.Symbol)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get symbol info: %w", err)
	}

	// Determine close side (opposite of entry side)
	closeSide := binance.OrderSideSell
	if trade.Side == "SELL" || trade.Side == "STRONG_SELL" {
		closeSide = binance.OrderSideBuy
	}

	// Adjust prices to precision
	tpAdjusted := binance.AdjustPricePrecision(trade.TPPrice, symbolInfo.TickSize)
	slAdjusted := binance.AdjustPricePrecision(trade.SLPrice, symbolInfo.TickSize)
	qtyAdjusted := binance.AdjustQuantityPrecision(totalQty, symbolInfo.StepSize)

	// Place Take Profit Market
	tpReq := &binance.PlaceOrderRequest{
		Symbol:     trade.Symbol,
		Side:       closeSide,
		Type:       binance.OrderTypeTakeProfitMarket,
		StopPrice:  tpAdjusted,
		ReduceOnly: true,
		Quantity:   qtyAdjusted,
	}

	tpResp, err := s.BinanceClient.PlaceOrder(tpReq)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to place TP order: %w", err)
	}

	// Place Stop Loss Market
	slReq := &binance.PlaceOrderRequest{
		Symbol:     trade.Symbol,
		Side:       closeSide,
		Type:       binance.OrderTypeStopMarket,
		StopPrice:  slAdjusted,
		ReduceOnly: true,
		Quantity:   qtyAdjusted,
	}

	slResp, err := s.BinanceClient.PlaceOrder(slReq)
	if err != nil {
		// TP sudah dibuat, tapi SL gagal - kita tetap return TP
		fmt.Printf("Warning: SL order failed but TP created: %v\n", err)
		return tpResp.OrderID, 0, nil
	}

	return tpResp.OrderID, slResp.OrderID, nil
}

// tradeMonitorFase3Netting handles Fase 3: Netting & Finalisasi
func (s *Services) tradeMonitorFase3Netting(ctx *gin.Context, trade *models.Trade) error {
	// ========================================================================
	// FASE 3: NETTING & FINALISASI
	// ========================================================================

	// 1. Refresh Data: Tarik ulang data TradeEntry dari DB untuk dapet state paling update
	entries, err := s.repo.TradeEntry.FindByTradeID(nil, trade.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch entries: %w", err)
	}

	// 2. Kalkulasi Induk (Update ke tabel Trade)
	var totalQty float64
	var capitalUsed float64

	for _, entry := range entries {
		if entry.Status == "FILLED" || entry.Status == "PARTIALLY_FILLED" {
			totalQty += entry.FilledQty
			capitalUsed += entry.FilledQty * entry.FilledPrice
		}
	}

	// Calculate average entry price
	avgEntryPrice := 0.0
	if totalQty > 0 {
		avgEntryPrice = capitalUsed / totalQty
	}

	// Update Trade dengan kalkulasi
	updateTrade := &models.Trade{
		TotalQty:      totalQty,
		CapitalUsed:   capitalUsed,
		AvgEntryPrice: avgEntryPrice,
	}
	_, err = s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, updateTrade)
	if err != nil {
		return fmt.Errorf("failed to update trade calculations: %w", err)
	}

	// 3. Evaluasi Dead Signal
	// Cek apakah SEMUA Entry berstatus "CANCELLED" atau "REJECTED"
	allCancelledOrRejected := true
	hasAnyFilled := false

	for _, entry := range entries {
		if entry.Status != "CANCELLED" && entry.Status != "REJECTED" {
			allCancelledOrRejected = false
		}
		if entry.Status == "FILLED" || entry.Status == "PARTIALLY_FILLED" {
			hasAnyFilled = true
		}
	}

	// Jika YA (semua cancelled/rejected) dan tidak ada yang filled: Update DB Trade.status = "CANCELLED"
	if allCancelledOrRejected && !hasAnyFilled {
		now := time.Now()
		updateTrade := &models.Trade{
			Status:     "CANCELLED",
			ClosedAt:   &now,
			ExitReason: "DEAD_SIGNAL",
		}
		_, err = s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, updateTrade)
		if err != nil {
			return fmt.Errorf("failed to update trade as dead signal: %w", err)
		}
	}

	return nil
}

// Helper function untuk menghitung weighted average price
func calculateWeightedAvgPrice(entries []models.TradeEntry) float64 {
	totalValue := 0.0
	totalQty := 0.0

	for _, entry := range entries {
		if entry.Status == "FILLED" || entry.Status == "PARTIALLY_FILLED" {
			totalValue += entry.FilledPrice * entry.FilledQty
			totalQty += entry.FilledQty
		}
	}

	if totalQty == 0 {
		return 0
	}

	return totalValue / totalQty
}

// Helper function untuk menghitung PnL
func calculatePnL(trade *models.Trade, currentPrice float64) float64 {
	if trade.AvgEntryPrice == 0 || trade.TotalQty == 0 {
		return 0
	}

	// PnL = (Current Price - Entry Price) * Qty untuk LONG
	// PnL = (Entry Price - Current Price) * Qty untuk SHORT
	if trade.Side == "BUY" || trade.Side == "STRONG_BUY" {
		return (currentPrice - trade.AvgEntryPrice) * trade.TotalQty
	}
	return (trade.AvgEntryPrice - currentPrice) * trade.TotalQty
}

// Helper function untuk menghitung PnL percentage
func calculatePnLPct(trade *models.Trade, pnl float64) float64 {
	if trade.CapitalUsed == 0 {
		return 0
	}
	return (pnl / trade.CapitalUsed) * 100
}

// Helper function untuk rounding float dengan presisi
func roundToPrecision(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}

// TradeMonitorProcessSingle processes a single trade by ID (public function for controller)
// Fetches trade details from DB and calls tradeMonitorProcessTrade
func (s *Services) TradeMonitorProcessSingle(ctx *gin.Context, req *dtos.TradeMonitorRequest) (*dtos.ProcessTradeResult, error) {
	// Get trade with entries from DB
	trade, err := s.repo.Trade.FindWithEntries(nil, req.TradeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trade: %w", err)
	}

	// Call private function to process the trade
	return s.tradeMonitorProcessTrade(ctx, trade)
}
