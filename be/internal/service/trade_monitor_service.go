package service

import (
	"context"
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
	// FASE 0: PERSIAPAN DATA (Tarik semua data dari Binance di awal)
	// ========================================================================

	result.Logs = append(result.Logs, fmt.Sprintf("Starting evaluation for trade #%d (%s)", trade.ID, trade.Symbol))

	// 0. Tarik Harga Market Terkini (Global GetPrice Optimization)
	// Ini menghemat pemanggilan API berkali-kali di Fase 1 (Fallback/Manual Close) dan Fase 3 (PnL)
	curPrice, err := s.BinanceClient.GetPrice(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current market price: %w", err)
	}
	currentMarketPrice := curPrice.Price
	result.Logs = append(result.Logs, fmt.Sprintf("Current Market Price: %v", currentMarketPrice))

	// 1. Validasi: Jika status Trade bukan "ACTIVE", langsung RETURN
	if trade.Status != "ACTIVE" {
		result.Status = "SKIPPED"
		result.Logs = append(result.Logs, fmt.Sprintf("Trade skipped: Status is %s (not ACTIVE)", trade.Status))
		result.Message = fmt.Sprintf("Trade status is %s, not ACTIVE", trade.Status)
		return result, nil
	}

	// 2.a Tarik Binance: GET All Open Orders untuk symbol ini (cache untuk cek di bawah)
	result.Logs = append(result.Logs, "Fetching open orders from Binance...")
	openOrders, err := s.BinanceClient.GetOpenOrders(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	// Build map for quick lookup: order_id -> order
	result.Logs = append(result.Logs, fmt.Sprintf("Found %d open orders on Binance", len(openOrders)))
	orderMap := make(map[int64]*binance.OrderResponse, len(openOrders))
	for i := range openOrders {
		orderMap[openOrders[i].OrderID] = &openOrders[i]
	}

	// 2.b Tarik Binance: GET All Open Algo Orders untuk symbol ini (cache untuk cek di bawah)
	result.Logs = append(result.Logs, "Fetching open algo orders from Binance...")
	openAlgoOrders, err := s.BinanceClient.GetOpenAlgoOrders(ctx, trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open algo orders: %w", err)
	}

	// Build map for quick lookup: algo_id -> algo order
	result.Logs = append(result.Logs, fmt.Sprintf("Found %d open algo orders on Binance", len(openAlgoOrders)))
	algoOrderMap := make(map[int64]*binance.AlgoOrderResponse, len(openAlgoOrders))
	for i := range openAlgoOrders {
		algoOrderMap[openAlgoOrders[i].AlgoID] = &openAlgoOrders[i]
	}

	// ========================================================================
	// FASE 1: CEK TP / SL (Prioritas Utama Pencegah Ghost Order)
	// ========================================================================
	result.Logs = append(result.Logs, "Phase 1: Checking TP/SL status...")
	fase1Result, shouldReturn, err := s.tradeMonitorFase1CheckTPSL(ctx, trade, orderMap, algoOrderMap, currentMarketPrice, result)
	if err != nil {
		result.Logs = append(result.Logs, fmt.Sprintf("ERROR in Phase 1: %v", err))
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
		result.Logs = append(result.Logs, "Trade closed due to TP/SL hit. Halting further checks.")
		result.Status = trade.Status // TP_HIT atau SL_HIT
		result.Message = "TP/SL hit, trade closed"
		return result, nil
	}

	// ========================================================================
	// FASE 2: SINKRONISASI JARING / ENTRY
	// ========================================================================
	result.Logs = append(result.Logs, "Phase 2: Syncing Entry orders...")
	fase2Result, err := s.tradeMonitorFase2SyncEntries(ctx, trade, orderMap, result)
	if err != nil {
		result.Logs = append(result.Logs, fmt.Sprintf("ERROR in Phase 2: %v", err))
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
	result.Logs = append(result.Logs, "Phase 3: Performing netting and finalizing metrics...")
	err = s.tradeMonitorFase3Netting(ctx, trade, currentMarketPrice, result)
	if err != nil {
		result.Logs = append(result.Logs, fmt.Sprintf("ERROR in Phase 3: %v", err))
		return nil, fmt.Errorf("fase 3 failed: %w", err)
	}

	// Reload trade to get final status
	result.Logs = append(result.Logs, "Reloading trade state from DB...")
	trade, err = s.repo.Trade.FindWithEntries(nil, trade.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload trade: %w", err)
	}

	result.Status = trade.Status
	result.Message = "Trade processed successfully"

	return result, nil
}

// tradeMonitorFase1CheckTPSL handles Fase 1: Pengecekan TP & SL
func (s *Services) tradeMonitorFase1CheckTPSL(
	ctx *gin.Context,
	trade *models.Trade,
	orderMap map[int64]*binance.OrderResponse,
	algoOrderMap map[int64]*binance.AlgoOrderResponse,
	currentMarketPrice float64,
	processResult *dtos.ProcessTradeResult,
) (*dtos.ProcessTradeResult, bool, error) {
	result := &dtos.ProcessTradeResult{}
	shouldReturn := false

	// Hitung total qty dari semua entry yang filled/partially filled menurut Database
	totalFilledQtyDB := 0.0
	for _, e := range trade.Entries {
		if e.Status == "FILLED" || e.Status == "PARTIALLY_FILLED" {
			totalFilledQtyDB += e.FilledQty
		}
	}

	// 1. GERBANG AWAL: Jika DB belum ada entry terisi, skip cek TP/SL
	if totalFilledQtyDB == 0 {
		processResult.Logs = append(processResult.Logs, "No filled entries yet, skipping TP/SL check.")
		return result, false, nil
	}

	// 2. KONDISI FALLBACK: DB punya koin, tapi TP/SL gagal terbuat (ID 0)
	if trade.TPOrderID == 0 {
		processResult.Logs = append(processResult.Logs, "⚠️ FALLBACK CHECK: No TP order ID found in DB. Checking price manually...")

		// Kita hemat API getPrice dengan menggunakan currentMarketPrice dari Fase 0
		hitTP := false
		hitSL := false

		if trade.Side == "BUY" || trade.Side == "STRONG_BUY" {
			if currentMarketPrice >= trade.TPPrice {
				hitTP = true
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("🚨 TP HIT SYSTEM (Dead Algo Status)! Price: %.8f >= TP: %.8f", currentMarketPrice, trade.TPPrice))
			}
			if currentMarketPrice <= trade.SLPrice {
				hitSL = true
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("🚨 SL HIT SYSTEM (Dead Algo Status)! Price: %.8f <= SL: %.8f", currentMarketPrice, trade.SLPrice))
			}
		} else if trade.Side == "SELL" || trade.Side == "STRONG_SELL" {
			if currentMarketPrice <= trade.TPPrice {
				hitTP = true
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("🚨 TP HIT SYSTEM (Dead Algo Status)! Price: %.8f <= TP: %.8f", currentMarketPrice, trade.TPPrice))
			}
			if currentMarketPrice >= trade.SLPrice {
				hitSL = true
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("🚨 SL HIT SYSTEM (Dead Algo Status)! Price: %.8f >= SL: %.8f", currentMarketPrice, trade.SLPrice))
			}
		}

		if hitTP || hitSL {
			processResult.Logs = append(processResult.Logs, "Executing manual FORCE close due to dead Algo ID...")
			exitReason := "TP_HIT_SYSTEM"
			if hitSL {
				exitReason = "SL_HIT_SYSTEM"
			}

			err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
				exitPrice := currentMarketPrice
				pnl := calculatePnL(trade, exitPrice)
				pnlPct := calculatePnLPct(trade, pnl)

				now := time.Now()
				updateTrade := &models.Trade{
					Status:     "CLOSED",
					ClosedAt:   &now,
					ExitPrice:  exitPrice,
					ExitReason: exitReason,
					PnL:        pnl,
					PnLPct:     pnlPct,
				}
				_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
				if err != nil {
					return fmt.Errorf("failed to update trade: %w", err)
				}

				closeSide := binance.OrderSideSell
				if trade.Side == "SELL" {
					closeSide = binance.OrderSideBuy
				}

				_, closeErr := s.BinanceClient.ClosePosition(trade.Symbol, totalFilledQtyDB, closeSide)
				if closeErr != nil {
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("⚠️ Warning: Failed to close Binance position: %v", closeErr))
				} else {
					processResult.Logs = append(processResult.Logs, "✅ Binance position closed successfully.")
				}

				if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel all orders: %v", cancelErr))
				}

				entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
				for _, entry := range entries {
					if entry.Status == "PENDING" || entry.Status == "NEW" {
						_ = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "Manual TP/SL hit")
					}
				}
				return nil
			})
			if err != nil {
				return result, false, err
			}
			trade.Status = "CLOSED"
			result.TPUpdated = hitTP
			result.SLUpdated = hitSL
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Trade closed by system due to TP/SL hit without orders. Exit Reason: %s", exitReason))
			return result, true, nil
		}

		processResult.Logs = append(processResult.Logs, fmt.Sprintf("⚠️ TP/SL not hit yet. Current: %.8f, TP: %.8f, SL: %.8f. Will retry creating orders if new entry filled.", currentMarketPrice, trade.TPPrice, trade.SLPrice))
		return result, false, nil
	}

	// 3. FOKUS UTAMA ALGO: Validasi Status TP/SL yang tertaut ID sah (>0)
	tpAlgoOrder, tpExists := algoOrderMap[trade.TPOrderID]
	slAlgoOrder, slExists := algoOrderMap[trade.SLOrderID]

	if !tpExists {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Algo %d not found in open orders. Fetching manually...", trade.TPOrderID))
		tpOrderDetail, err := s.BinanceClient.GetAlgoOrder(ctx, &binance.GetAlgoOrdersRequest{
			Symbol:  trade.Symbol,
			AlgoID:  trade.TPOrderID,
		})
		if err == nil {
			tpAlgoOrder = tpOrderDetail
			tpExists = true
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Algo %d retrieved, status: %s", trade.TPOrderID, tpAlgoOrder.AlgoStatus))
		}
	} else {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Algo %d is in open orders, status: %s", trade.TPOrderID, tpAlgoOrder.AlgoStatus))
	}

	if !slExists {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Algo %d not found in open orders. Fetching manually...", trade.SLOrderID))
		slOrderDetail, err := s.BinanceClient.GetAlgoOrder(ctx, &binance.GetAlgoOrdersRequest{
			Symbol:  trade.Symbol,
			AlgoID:  trade.SLOrderID,
		})
		if err == nil {
			slAlgoOrder = slOrderDetail
			slExists = true
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Algo %d retrieved, status: %s", trade.SLOrderID, slAlgoOrder.AlgoStatus))
		}
	} else {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Algo %d is in open orders, status: %s", trade.SLOrderID, slAlgoOrder.AlgoStatus))
	}

	tpFilled := tpExists && tpAlgoOrder.AlgoStatus == "FILLED"
	slFilled := slExists && slAlgoOrder.AlgoStatus == "FILLED"

	// Jika Sistem API TP/SL mendeteksi FILLED - Murni Action dari System Close
	if tpFilled || slFilled {
		// Validasi Actual Posisi untuk Mismatch Handling
		position, posErr := s.BinanceClient.GetPosition(trade.Symbol)
		actualQty := 0.0
		if posErr == nil && position != nil {
			actualQty = math.Abs(position.PositionAmt)
		}

		if actualQty > totalFilledQtyDB && actualQty > 0 {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("⚠️ GHOST POSITION DETECTED (System Hit)! DB: %.8f, Binance: %.8f", totalFilledQtyDB, actualQty))

			// Auto sync entries untuk last effort update DB
			_, _ = s.tradeMonitorFase2SyncEntries(ctx, trade, orderMap, processResult)

			dbTotalQty := 0.0
			for _, e := range trade.Entries {
				if e.Status == "FILLED" || e.Status == "PARTIALLY_FILLED" {
					dbTotalQty += e.FilledQty
				}
			}

			exitReason := "TP_HIT_MISMATCH"
			exitPrice := trade.TPPrice
			if slFilled {
				exitReason = "SL_HIT_MISMATCH"
				exitPrice = trade.SLPrice
			}

			err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
				tempTrade := *trade
				tempTrade.TotalQty = actualQty
				pnl := calculatePnL(&tempTrade, exitPrice)
				pnlPct := calculatePnLPct(&tempTrade, pnl)

				now := time.Now()
				updateTrade := &models.Trade{
					Status:     "CLOSED",
					ClosedAt:   &now,
					ExitPrice:  exitPrice,
					ExitReason: exitReason,
					PnL:        pnl,
					PnLPct:     pnlPct,
					TotalQty:   actualQty,
				}
				_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
				if err != nil {
					return err
				}

				closeSide := binance.OrderSideSell
				if trade.Side == "SELL" || trade.Side == "STRONG_SELL" {
					closeSide = binance.OrderSideBuy
				}

				_, _ = s.BinanceClient.ClosePosition(trade.Symbol, actualQty, closeSide)

				// Cancel semua pending entries di Binance dengan 1 API call
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Canceling ALL open orders for symbol %s...", trade.Symbol))
				if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel all orders for %s: %v", trade.Symbol, cancelErr))
				} else {
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("ALL orders for %s canceled on Binance successfully.", trade.Symbol))
				}

				entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch entries: %w", err)
				}
				
				for _, entry := range entries {
					if entry.Status == "PENDING" || entry.Status == "NEW" {
						err = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", exitReason)
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
			trade.Status = "CLOSED"
			trade.TotalQty = actualQty
			result.TPUpdated = tpFilled
			result.SLUpdated = slFilled
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("✅ Trade closed with %s. Actual Qty Hit: %.8f", exitReason, actualQty))
			return result, true, nil
		}

		// Alur Normal: Posisi cocok (sudah 0) = Valid TP/SL Hit
		processResult.Logs = append(processResult.Logs, "Valid TP/SL Hit. Executing Normal System Close...")
		err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
			exitReason := "TP_HIT"
			exitPrice := trade.TPPrice
			if slFilled {
				exitReason = "SL_HIT"
				exitPrice = trade.SLPrice
			}

			pnl := calculatePnL(trade, exitPrice)
			pnlPct := calculatePnLPct(trade, pnl)

			now := time.Now()
			updateTrade := &models.Trade{
				Status:     "CLOSED",
				ClosedAt:   &now,
				ExitPrice:  exitPrice,
				ExitReason: exitReason,
				PnL:        pnl,
				PnLPct:     pnlPct,
			}
			_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
			if err != nil {
				return err
			}

			// 🚨 CRITICAL: Cancel semua order di Binance dengan 1 API call
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Canceling ALL open orders for symbol %s...", trade.Symbol))
			if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel all orders for %s: %v", trade.Symbol, cancelErr))
			} else {
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("ALL orders for %s canceled on Binance successfully.", trade.Symbol))
			}

			// Get all entries for this trade to update DB
			entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch entries: %w", err)
			}

			// Update DB Status
			for _, entry := range entries {
				if entry.Status == "PENDING" || entry.Status == "NEW" {
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
		trade.Status = "CLOSED"
		if slFilled {
			trade.Status = "SL_HIT"
		} else {
			trade.Status = "TP_HIT"
		}
		result.TPUpdated = tpFilled
		result.SLUpdated = slFilled
		return result, true, nil
	}

	// 4. BENTENG TERAKHIR (MANUAL CLOSE USER): System belum Hit (Algo Status msh NEW), tapi fisik koin di Binance tiba-tiba Lenyap (0)!
	position, err := s.BinanceClient.GetPosition(trade.Symbol)
	if err == nil && position.PositionAmt == 0 {
		processResult.Logs = append(processResult.Logs, "🚨 EMERGENCY: Binance TP/SL Algo is NOT filled yet, but actual Binance position is 0! User must have closed manually.")
		// Kita gunakan currentMarketPrice dari Fase 0 sebagai harga exit perkiraan terbaik

		err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
			pnl := calculatePnL(trade, currentMarketPrice)
			pnlPct := calculatePnLPct(trade, pnl)

			now := time.Now()
			updateTrade := &models.Trade{
				Status:     "CLOSED",
				ClosedAt:   &now,
				ExitPrice:  currentMarketPrice,
				ExitReason: "MANUAL_CLOSE",
				PnL:        pnl,
				PnLPct:     pnlPct,
			}
			_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
			if err != nil {
				return err
			}

			// Cancel semua order di Binance dengan 1 API call
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Canceling ALL open orders for symbol %s...", trade.Symbol))
			if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel all orders for %s: %v", trade.Symbol, cancelErr))
			} else {
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("ALL orders for %s canceled on Binance successfully.", trade.Symbol))
			}

			entries, _ := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
			for _, entry := range entries {
				if entry.Status == "PENDING" || entry.Status == "NEW" {
					_ = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "Manual close")
				}
			}

			if trade.TPOrderID > 0 {
				_, _ = s.BinanceClient.CancelAlgoOrder(ctx, &binance.CancelAlgoOrderRequest{Symbol: trade.Symbol, AlgoID: trade.TPOrderID})
			}
			if trade.SLOrderID > 0 {
				_, _ = s.BinanceClient.CancelAlgoOrder(ctx, &binance.CancelAlgoOrderRequest{Symbol: trade.Symbol, AlgoID: trade.SLOrderID})
			}
			return nil
		})

		if err != nil {
			return result, false, err
		}

		processResult.Logs = append(processResult.Logs, "Trade closed manually by user. Halting further checks.")
		trade.Status = "MANUAL_CLOSE"
		result.TPUpdated = false
		result.SLUpdated = false
		return result, true, nil
	}
	return result, shouldReturn, nil
}

// tradeMonitorFase2SyncEntries handles Fase 2: Sinkronisasi Entry
func (s *Services) tradeMonitorFase2SyncEntries(
	ctx *gin.Context,
	trade *models.Trade,
	orderMap map[int64]*binance.OrderResponse,
	processResult *dtos.ProcessTradeResult,
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
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Failed to fetch detail for Entry Order %d (%v).", entry.BinanceOrderID, err))
					// Order tidak ditemukan/gagal di-fetch - anggap cancelled HANYA jika belum punya barang
					if entry.Status == "PENDING" || entry.Status == "NEW" {
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d is missing and was pending. Marking as CANCELLED.", entry.BinanceOrderID))
						entry.Status = "CANCELLED"
						updatedCount++
					}
					continue
				}

				binanceOrder = orderDetail
				exists = true
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry order %d recovered manually. Status: %s", entry.BinanceOrderID, binanceOrder.Status))
			} else {
				// Entry sudah filled/cancelled sebelumnya secara permanen, skip
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry order %d already has final status %s, skipping.", entry.BinanceOrderID, entry.Status))
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
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d expired (Age: %v). Canceling...", entry.BinanceOrderID, orderAge))
				// Expired - Cancel order
				_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
					Symbol:  trade.Symbol,
					OrderID: entry.BinanceOrderID,
				})

				if err != nil {
					fmt.Printf("Warning: Failed to cancel expired order %d: %v\n", entry.BinanceOrderID, err)
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel expired Entry Order %d: %v", entry.BinanceOrderID, err))
				} else {
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Expired Entry Order %d canceled on Binance.", entry.BinanceOrderID))
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
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d is PARTIALLY_FILLED. Updating DB...", entry.BinanceOrderID))
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
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d is FILLED! Updating DB...", entry.BinanceOrderID))
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
			// [Logic Pasang TP & SL Algo]
			// ================================================================

			// Hitung total qty dari semua entry yang filled/partially filled
			totalFilledQty := 0.0
			for _, e := range trade.Entries {
				if e.Status == "FILLED" || e.Status == "PARTIALLY_FILLED" {
					totalFilledQty += e.FilledQty
				}
			}

			// Karena Algo Order memakai closePosition=true, kita TIDAK PERLU 
			// melakukan "Cancel & Replace" setiap kali averaging hit!
			// Kita hanya perlu membuatnya SEKALI: saat entry pertama kena, 
			// atau saat Fallback (ID = 0 karena gagal dibuat sebelumnya).
			if trade.TPOrderID == 0 && totalFilledQty > 0 {
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entries filled but TP/SL Algo missing (ID: 0). Creating Algo TP/SL..."))
				
				tpOrderID, slOrderID, err := s.tradeMonitorCreateAlgoTPOrder(ctx, trade, processResult)
				if err != nil {
					// 🚨 JIKA GAGAL CREATE, TRADE TIDAK PUNYA TP/SL DI BINANCE SAMA SEKALI
					// Kita mengosongkan TPOrderID & SLOrderID di Database agar
					// di putaran cron berikutnya bot ini men-trigger fallback kembali.
					fallbackUpdate := &models.Trade{
						TPOrderID: 0,
						SLOrderID: 0,
					}
					s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, fallbackUpdate)

					trade.TPOrderID = 0
					trade.SLOrderID = 0

					return result, fmt.Errorf("failed to create TP/SL Algo orders: %w", err)
				}

				// Simpan ID TP/SL Algo barunya ke DB
				updateTrade := &models.Trade{
					TPOrderID:   tpOrderID,
					SLOrderID:   slOrderID,
					TotalQty:    totalFilledQty, // Approximate total qty for visual reference
					CapitalUsed: totalFilledQty * entry.FilledPrice,
				}
				_, err = s.repo.Trade.Update(nil, &models.Trade{ID: trade.ID}, updateTrade)
				if err != nil {
					return result, fmt.Errorf("failed to update trade with TP/SL Algo IDs: %w", err)
				}

				trade.TPOrderID = tpOrderID
				trade.SLOrderID = slOrderID
				tpUpdated = true
				slUpdated = true
			} else if totalFilledQty > 0 {
				// Averaging hit, TP/SL sudah ada (> 0). 
				// Biarkan saja, `closePosition=true` akan mengatur semuanya!
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Averaging entry filled. Existing TP/SL Algo (%d, %d) will automatically adapt using closePosition=true.", trade.TPOrderID, trade.SLOrderID))
			}
		}
	}

	result.UpdatedCount = updatedCount
	result.TPUpdated = tpUpdated
	result.SLUpdated = slUpdated

	return result, nil
}

// tradeMonitorCreateAlgoTPOrder creates conditional Algo TP and SL orders for a trade
func (s *Services) tradeMonitorCreateAlgoTPOrder(ctx context.Context, trade *models.Trade, processResult *dtos.ProcessTradeResult) (int64, int64, error) {
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

	// Place Take Profit Market (Algo Order)
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("Requesting new Algo TP Order (TriggerPrice: %v, Side: %s, ClosePosition: true)...", tpAdjusted, closeSide))
	
	tpReq := &binance.PlaceAlgoOrderRequest{
		Symbol:        trade.Symbol,
		Side:          closeSide,
		Type:          binance.OrderTypeTakeProfitMarket,
		TriggerPrice:  tpAdjusted,
		ClosePosition: true,
	}

	tpResp, err := s.BinanceClient.PlaceAlgoOrder(ctx, tpReq)
	if err != nil {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("FAILED setting Algo Take Profit: %v", err))
		return 0, 0, fmt.Errorf("failed to place Algo TP order: %w", err)
	}
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("SUCCESS setting Algo Take Profit (AlgoID: %d)", tpResp.AlgoID))

	// Place Stop Loss Market (Algo Order)
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("Requesting new Algo SL Order (TriggerPrice: %v, Side: %s, ClosePosition: true)...", slAdjusted, closeSide))
	
	slReq := &binance.PlaceAlgoOrderRequest{
		Symbol:        trade.Symbol,
		Side:          closeSide,
		Type:          binance.OrderTypeStopMarket,
		TriggerPrice:  slAdjusted,
		ClosePosition: true,
	}

	slResp, err := s.BinanceClient.PlaceAlgoOrder(ctx, slReq)
	if err != nil {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("FAILED setting Algo Stop Loss: %v", err))
		
		// Rollback TP if SL fails to keep state consistent
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("Rolling back orphaned Algo TP %d...", tpResp.AlgoID))
		s.BinanceClient.CancelAlgoOrder(ctx, &binance.CancelAlgoOrderRequest{Symbol: trade.Symbol, AlgoID: tpResp.AlgoID})
		
		return 0, 0, fmt.Errorf("failed to place Algo SL order: %w", err)
	}
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("SUCCESS setting Algo Stop Loss (AlgoID: %d)", slResp.AlgoID))

	return tpResp.AlgoID, slResp.AlgoID, nil
}

// tradeMonitorFase3Netting handles Fase 3: Netting & Finalisasi
func (s *Services) tradeMonitorFase3Netting(ctx *gin.Context, trade *models.Trade, currentMarketPrice float64, processResult *dtos.ProcessTradeResult) error {
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
	var totalValue float64 // Used for Volume Weighted Average Price

	for _, entry := range entries {
		if entry.Status == "FILLED" || entry.Status == "PARTIALLY_FILLED" {
			totalQty += entry.FilledQty
			totalValue += entry.FilledQty * entry.FilledPrice

			// Capital Used is real margin amount: Position Value (Price * Qty) / Leverage
			capitalUsed += (entry.FilledQty * entry.FilledPrice) / float64(trade.Leverage)
		}
	}

	// Calculate Average Entry Price based on Volume Weighted Average
	avgEntryPrice := 0.0
	if totalQty > 0 {
		avgEntryPrice = totalValue / totalQty
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
		processResult.Logs = append(processResult.Logs, "Zero entries filled and all entries cancelled/rejected. Marking trade as DEAD_SIGNAL (CANCELLED).")
		
		// 🚨 SAPU JAGAT BACKUP PLAN (Pencegah Zombie Limit Order 🧟)
		// Jika Fase 2 gagal cancel akibat network down, paksa sapu bersih di sini sebelum menutup Trade DB.
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("Executing CancelAllOrders for symbol %s as DEAD_SIGNAL precaution...", trade.Symbol))
		if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: CancelAllOrders failed during DEAD_SIGNAL execution: %v", cancelErr))
		} else {
			processResult.Logs = append(processResult.Logs, "All remaining orders wiped clean from Binance.")
		}

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
	} else {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("Netting updated. Current average entry price: %f, Total Qty: %f, Capital Used: %f", avgEntryPrice, totalQty, capitalUsed))
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
	if trade.CapitalUsed == 0 || trade.Leverage == 0 {
		return 0
	}

	// CapitalUsed in DB now correctly represents the actual margin amount
	// (not the leveraged total position value)
	actualCapital := trade.CapitalUsed

	if actualCapital == 0 {
		return 0
	}

	return (pnl / actualCapital) * 100
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

// TradeManualClose allows a user to manually close an active trade via the app
func (s *Services) TradeManualClose(ctx *gin.Context, tradeID uint) (*dtos.ProcessTradeResult, error) {
	result := &dtos.ProcessTradeResult{
		TradeID: tradeID,
	}

	// 1. Fetch trade with entries
	trade, err := s.repo.Trade.FindWithEntries(nil, tradeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trade: %w", err)
	}

	result.Symbol = trade.Symbol
	result.Logs = append(result.Logs, fmt.Sprintf("Starting Manual Close for trade #%d (%s)", trade.ID, trade.Symbol))

	// 2. Validate status
	if trade.Status != "ACTIVE" {
		result.Status = "SKIPPED"
		result.Message = fmt.Sprintf("Trade status is %s, not ACTIVE", trade.Status)
		return result, nil
	}

	// 3. Sync Status to ensure we have the correct Filled Qty
	result.Logs = append(result.Logs, "Syncing exact filled quantities from Binance open orders...")
	openOrders, err := s.BinanceClient.GetOpenOrders(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders for sync: %w", err)
	}
	orderMap := make(map[int64]*binance.OrderResponse, len(openOrders))
	for i := range openOrders {
		orderMap[openOrders[i].OrderID] = &openOrders[i]
	}

	// Run Fase 2 sync to ensure DB accurately reflects Binance
	syncResult, err := s.tradeMonitorFase2SyncEntries(ctx, trade, orderMap, result)
	if err != nil {
		result.Logs = append(result.Logs, fmt.Sprintf("Warning: Failed to sync entries: %v", err))
		// We can still proceed to market close even if sync failed partially, but let's record the error
	} else {
		result.EntriesSync = syncResult.UpdatedCount
	}

	// Fetch current market price (needed for exitPrice calculation and Fase 3 parameter)
	curPrice, err := s.BinanceClient.GetPrice(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market price: %w", err)
	}
	exitPrice := curPrice.Price

	// Run Fase 3 Netting to get final TotalQty & AvgEntryPrice
	err = s.tradeMonitorFase3Netting(ctx, trade, exitPrice, result)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate netting: %w", err)
	}

	// Reload trade to get freshly netted qty
	trade, err = s.repo.Trade.FindWithEntries(nil, tradeID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload netted trade: %w", err)
	}

	// 4. Execute Market Close FIRST (close physical position before cancelling orders)
	if trade.TotalQty > 0 {
		closeSide := binance.OrderSideSell
		if trade.Side == "SELL" || trade.Side == "STRONG_SELL" {
			closeSide = binance.OrderSideBuy
		}

		result.Logs = append(result.Logs, fmt.Sprintf("Executing Market %s for TotalQty: %v to close position...", closeSide, trade.TotalQty))

		symbolInfo, err := s.BinanceClient.GetSymbolInfo(trade.Symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get symbol info: %w", err)
		}

		qtyAdjusted := binance.AdjustQuantityPrecision(trade.TotalQty, symbolInfo.StepSize)

		closeReq := &binance.PlaceOrderRequest{
			Symbol:     trade.Symbol,
			Side:       closeSide,
			Type:       binance.OrderTypeMarket,
			ReduceOnly: true,
			Quantity:   qtyAdjusted,
		}

		closeResp, err := s.BinanceClient.PlaceOrder(closeReq)
		if err != nil {
			result.Logs = append(result.Logs, fmt.Sprintf("FAILED executing Market Close: %v", err))
			return nil, fmt.Errorf("failed to place market close order: %w", err)
		}

		result.Logs = append(result.Logs, fmt.Sprintf("SUCCESS Market Close. OrderID: %d", closeResp.OrderID))
	} else {
		result.Logs = append(result.Logs, "Total Qty is 0, no market close order needed.")
	}

	// 5. Cancel remaining pending orders (after position is already closed)
	result.Logs = append(result.Logs, fmt.Sprintf("Canceling ALL open orders for symbol %s...", trade.Symbol))
	if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
		result.Logs = append(result.Logs, fmt.Sprintf("Warning: Failed to cancel all open orders: %v", cancelErr))
	} else {
		result.Logs = append(result.Logs, "All open orders canceled successfully.")
	}

	// 6. Update Database Record
	err = s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
		pnl := calculatePnL(trade, exitPrice)
		pnlPct := calculatePnLPct(trade, pnl)

		now := time.Now()
		updateTrade := &models.Trade{
			Status:     "CLOSED",
			ClosedAt:   &now,
			ExitPrice:  exitPrice,
			ExitReason: "MANUAL_CLOSE_BY_USER",
			PnL:        pnl,
			PnLPct:     pnlPct,
			TPOrderID:  0, // Cleared out
			SLOrderID:  0, // Cleared out
		}

		_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
		if err != nil {
			return fmt.Errorf("failed to update trade status: %w", err)
		}

		entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch entries for update: %w", err)
		}

		for _, entry := range entries {
			if entry.Status == "PENDING" || entry.Status == "NEW" {
				err = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "Manual close via App")
				if err != nil {
					return fmt.Errorf("failed to cancel entry status in DB: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to commit db updates: %w", err)
	}

	result.Status = "CLOSED"
	result.Message = "Trade closed manually by user."

	return result, nil
}
