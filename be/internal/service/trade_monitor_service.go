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

	result.Logs = append(result.Logs, fmt.Sprintf("Starting evaluation for trade #%d (%s)", trade.ID, trade.Symbol))

	// 1. Validasi: Jika status Trade bukan "ACTIVE", langsung RETURN
	if trade.Status != "ACTIVE" {
		result.Status = "SKIPPED"
		result.Logs = append(result.Logs, fmt.Sprintf("Trade skipped: Status is %s (not ACTIVE)", trade.Status))
		result.Message = fmt.Sprintf("Trade status is %s, not ACTIVE", trade.Status)
		return result, nil
	}

	// 2. Tarik Binance: GET All Open Orders untuk symbol ini (cache untuk cek di bawah)
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

	// ========================================================================
	// FASE 1: CEK TP / SL (Prioritas Utama Pencegah Ghost Order)
	// ========================================================================
	result.Logs = append(result.Logs, "Phase 1: Checking TP/SL status...")
	fase1Result, shouldReturn, err := s.tradeMonitorFase1CheckTPSL(ctx, trade, orderMap, result)
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
	err = s.tradeMonitorFase3Netting(ctx, trade, result)
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

// tradeMonitorFase1CheckTPSL handles Fase 1: Cek TP/SL
func (s *Services) tradeMonitorFase1CheckTPSL(
	ctx *gin.Context,
	trade *models.Trade,
	orderMap map[int64]*binance.OrderResponse,
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

	// ========================================================================
	// CEK POSISI: Validasi apakah user melakukan Manual Close di Binance
	// ========================================================================
	if totalFilledQtyDB > 0 {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("Checking actual Binance position size. DB thinks we have %v coins.", totalFilledQtyDB))
		position, err := s.BinanceClient.GetPosition(trade.Symbol)
		if err != nil {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to fetch position from Binance: %v", err))
		} else {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Actual Binance PositionAmt is %v", position.PositionAmt))

			// Jika di DB kita punya barang, tapi di Binance PositionAmt == 0,
			// artinya user telah MENUTUP posisi ini secara manual lewat aplikasi!
			if position.PositionAmt == 0 {
				processResult.Logs = append(processResult.Logs, "🚨 EMERGENCY: Binance position is 0 but DB has coins! User must have closed manually.")
				curPrice, _ := s.BinanceClient.GetPrice(trade.Symbol)
				err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
					// Hitung PnL untuk dicatat di DB
					pnl := calculatePnL(trade, curPrice.Price)
					pnlPct := calculatePnLPct(trade, pnl)
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Close Position With Market Price %.2f, pnl %.3f", curPrice.Price, pnl))
					// 1. Update trade status
					now := time.Now()
					updateTrade := &models.Trade{
						Status:     "CLOSED",
						ClosedAt:   &now,
						ExitPrice:  curPrice.Price, // Approximate closing price
						ExitReason: "MANUAL_CLOSE",
						PnL:        pnl,
						PnLPct:     pnlPct,
					}
					_, err := s.repo.Trade.Update(tx, &models.Trade{ID: trade.ID}, updateTrade)
					if err != nil {
						return fmt.Errorf("failed to update trade status to CLOSED (Manual): %w", err)
					}

					// 2. Cancel semua order jaring (Entry) yang masih ngantre ("NEW")
					entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
					if err != nil {
						return fmt.Errorf("failed to fetch entries: %w", err)
					}

					for _, entry := range entries {
						if entry.Status == "PENDING" || entry.Status == "NEW" {
							processResult.Logs = append(processResult.Logs, fmt.Sprintf("Canceling pending Entry Order %d...", entry.BinanceOrderID))
							_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
								Symbol:  trade.Symbol,
								OrderID: entry.BinanceOrderID,
							})

							if err != nil {
								processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Cancel entry order %d failed: %v", entry.BinanceOrderID, err))
							} else {
								processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d canceled on Binance successfully.", entry.BinanceOrderID))
							}

							err = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "Manual close")
							if err != nil {
								return fmt.Errorf("failed to update entry status: %w", err)
							}
						}
					}

					// 3. Cancel TP/SL if they still exist (often closed automatically by Binance, but just to be sure)
					if trade.TPOrderID > 0 {
						s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{Symbol: trade.Symbol, OrderID: trade.TPOrderID})
					}
					if trade.SLOrderID > 0 {
						s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{Symbol: trade.Symbol, OrderID: trade.SLOrderID})
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
		}
	}

	// Cek DB: Apakah tp_order_id sudah ada isinya?
	if trade.TPOrderID == 0 {
		processResult.Logs = append(processResult.Logs, "No TP order ID found in DB.")
		
		// 🆕 FALLBACK: Cek apakah ada entry yang sudah filled tapi tidak punya TP/SL
		hasFilledEntry := false
		totalFilledQty := 0.0
		
		for _, e := range trade.Entries {
			if e.Status == "FILLED" || e.Status == "PARTIALLY_FILLED" {
				hasFilledEntry = true
				totalFilledQty += e.FilledQty
			}
		}
		
		if hasFilledEntry && totalFilledQty > 0 {
			processResult.Logs = append(processResult.Logs, 
				fmt.Sprintf("⚠️ WARNING: Trade has %.8f filled qty but NO TP/SL! Checking price manually...", totalFilledQty))
			
			// 🆕 Cek harga sekarang vs TP/SL price di DB
			curPrice, err := s.BinanceClient.GetPrice(trade.Symbol)
			if err != nil {
				return result, false, fmt.Errorf("failed to get price: %w", err)
			}
			
			// Cek apakah harga sudah hit TP atau SL
			hitTP := false
			hitSL := false
			
			// Note: trade.Side menggunakan "BUY" untuk LONG dan "SELL" untuk SHORT
			if trade.Side == "BUY" {
				// LONG: TP di atas, SL di bawah
				if curPrice.Price >= trade.TPPrice {
					hitTP = true
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("🚨 TP HIT SYSTEM! Price: %.8f >= TP: %.8f", curPrice.Price, trade.TPPrice))
				}
				if curPrice.Price <= trade.SLPrice {
					hitSL = true
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("🚨 SL HIT SYSTEM! Price: %.8f <= SL: %.8f", curPrice.Price, trade.SLPrice))
				}
			} else if trade.Side == "SELL" {
				// SHORT: TP di bawah, SL di atas
				if curPrice.Price <= trade.TPPrice {
					hitTP = true
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("🚨 TP HIT SYSTEM! Price: %.8f <= TP: %.8f", curPrice.Price, trade.TPPrice))
				}
				if curPrice.Price >= trade.SLPrice {
					hitSL = true
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("🚨 SL HIT SYSTEM! Price: %.8f >= SL: %.8f", curPrice.Price, trade.SLPrice))
				}
			}
			
			// Jika TP/SL hit, langsung close trade
			if hitTP || hitSL {
				processResult.Logs = append(processResult.Logs, "Executing manual TP/SL close...")

				exitReason := "TP_HIT_SYSTEM"
				if hitSL {
					exitReason = "SL_HIT_SYSTEM"
				}

				err := s.repo.TxManager.WithinTransaction(func(tx *gorm.DB) error {
					// Gunakan current price sebagai exit price (real market price saat close)
					// Ini konsisten dengan manual close behavior di Binance
					exitPrice := curPrice.Price

					// Hitung PnL
					pnl := calculatePnL(trade, exitPrice)
					pnlPct := calculatePnLPct(trade, pnl)

					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("Close Position With Price %.8f, PnL: %.8f (%.2f%%)", exitPrice, pnl, pnlPct))

					// Update trade status
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

					// 🚨 CRITICAL: Close actual Binance position first (market order opposite direction)
					// This is the REAL close - without this, position stays open on Binance!
					closeSide := binance.OrderSideSell // BUY position => close with SELL
					if trade.Side == "SELL" {
						closeSide = binance.OrderSideBuy // SELL position => close with BUY
					}
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("Closing Binance position for %s (Qty: %v, Side: %s)...", trade.Symbol, totalFilledQty, closeSide))

					_, closeErr := s.BinanceClient.ClosePosition(trade.Symbol, totalFilledQty, closeSide)
					if closeErr != nil {
						processResult.Logs = append(processResult.Logs,
							fmt.Sprintf("⚠️ Warning: Failed to close Binance position for %s: %v", trade.Symbol, closeErr))
					} else {
						processResult.Logs = append(processResult.Logs,
							fmt.Sprintf("✅ Binance position for %s closed successfully.", trade.Symbol))
					}

					// Cancel semua pending entry orders untuk symbol ini (1 API call untuk semua orders)
					processResult.Logs = append(processResult.Logs,
						fmt.Sprintf("Canceling ALL open orders for symbol %s...", trade.Symbol))

					if cancelErr := s.BinanceClient.CancelAllOrders(trade.Symbol); cancelErr != nil {
						processResult.Logs = append(processResult.Logs,
							fmt.Sprintf("Warning: Failed to cancel all orders for %s: %v", trade.Symbol, cancelErr))
					} else {
						processResult.Logs = append(processResult.Logs,
							fmt.Sprintf("ALL orders for %s canceled on Binance successfully.", trade.Symbol))
					}

					// Update semua entry status jadi CANCELLED di DB
					entries, err := s.repo.TradeEntry.FindByTradeID(tx, trade.ID)
					if err != nil {
						return fmt.Errorf("failed to fetch entries: %w", err)
					}

					for _, entry := range entries {
						if entry.Status == "PENDING" || entry.Status == "NEW" {
							err = s.repo.TradeEntry.UpdateStatus(tx, entry.ID, "CANCELLED", "Manual TP/SL hit")
							if err != nil {
								return fmt.Errorf("failed to update entry: %w", err)
							}
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
				shouldReturn = true
				processResult.Logs = append(processResult.Logs,
					fmt.Sprintf("Trade closed by system due to TP/SL hit without orders. Exit Reason: %s", exitReason))
				return result, true, nil
			}
			
			// Jika belum hit, log warning dan lanjutkan (akan retry create TP/SL di Fase 2 jika ada entry baru)
			processResult.Logs = append(processResult.Logs, 
				fmt.Sprintf("⚠️ TP/SL not hit yet. Current: %.8f, TP: %.8f, SL: %.8f. Will retry creating orders if new entry filled.", 
					curPrice.Price, trade.TPPrice, trade.SLPrice))
		} else {
			processResult.Logs = append(processResult.Logs, "No filled entries yet, skipping TP/SL check.")
		}
		
		return result, false, nil
	}

	// Cocokkan ID TP/SL tersebut dengan data cache Binance
	tpOrder, tpExists := orderMap[trade.TPOrderID]
	slOrder, slExists := orderMap[trade.SLOrderID]

	// 🚨 PENTING: GetOpenOrders HANYA mengembalikan order yang NEW/PARTIALLY_FILLED.
	// Jika status TP atau SL sudah "FILLED", order akan hilang dari orderMap!
	// Oleh karena itu, jika tidak ada di orderMap, kita wajib hit GET Order Detail.
	if !tpExists && trade.TPOrderID > 0 {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Order %d not found in open orders. Fetching manually...", trade.TPOrderID))
		tpOrderDetail, err := s.BinanceClient.GetOrder(&binance.GetOrdersRequest{
			Symbol:  trade.Symbol,
			OrderID: trade.TPOrderID,
		})
		if err == nil {
			tpOrder = tpOrderDetail
			tpExists = true
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Order %d retrieved, status: %s", trade.TPOrderID, tpOrder.Status))
		} else {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Failed to fetch TP Order %d: %v", trade.TPOrderID, err))
		}
	} else if tpExists {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("TP Order %d is in open orders, status: %s", trade.TPOrderID, tpOrder.Status))
	}

	if !slExists && trade.SLOrderID > 0 {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Order %d not found in open orders. Fetching manually...", trade.SLOrderID))
		slOrderDetail, err := s.BinanceClient.GetOrder(&binance.GetOrdersRequest{
			Symbol:  trade.Symbol,
			OrderID: trade.SLOrderID,
		})
		if err == nil {
			slOrder = slOrderDetail
			slExists = true
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Order %d retrieved, status: %s", trade.SLOrderID, slOrder.Status))
		} else {
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Failed to fetch SL Order %d: %v", trade.SLOrderID, err))
		}
	} else if slExists {
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("SL Order %d is in open orders, status: %s", trade.SLOrderID, slOrder.Status))
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
			exitReason := "TP_HIT"
			exitPrice := trade.TPPrice
			if slFilled {
				exitReason = "SL_HIT"
				exitPrice = trade.SLPrice
			}
			processResult.Logs = append(processResult.Logs, fmt.Sprintf("Exit condition met: %s. Updating trade status in DB...", exitReason))

			// Hitung PnL
			pnl := calculatePnL(trade, exitPrice)
			pnlPct := calculatePnLPct(trade, pnl)

			// Update trade status to CLOSED
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
				return fmt.Errorf("failed to update trade status to CLOSED (%s): %w", exitReason, err)
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
					processResult.Logs = append(processResult.Logs, fmt.Sprintf("Canceling pending Entry Order %d...", entry.BinanceOrderID))

					// Cancel order on Binance
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: entry.BinanceOrderID,
					})

					// Ignore error if order already filled/cancelled
					if err != nil {
						fmt.Printf("Warning: Failed to cancel entry order %d: %v\n", entry.BinanceOrderID, err)
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Cancel entry order %d failed: %v", entry.BinanceOrderID, err))
					} else {
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Entry Order %d canceled on Binance successfully.", entry.BinanceOrderID))
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
				processResult.Logs = append(processResult.Logs, fmt.Sprintf("First entry filled (Total Qty: %v). Creating original TP/SL...", totalFilledQty))
				// Hit API Binance CREATE order TP & SL sesuai qty yang didapat
				tpOrderID, slOrderID, err := s.tradeMonitorCreateTPOrder(trade, totalFilledQty, processResult)
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

				processResult.Logs = append(processResult.Logs, fmt.Sprintf("Averaging entry filled. Canceling old TP/SL to replace with Total Qty: %v", totalFilledQty))

				// Cancel TP/SL lama
				if trade.TPOrderID > 0 {
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: trade.TPOrderID,
					})
					if err != nil {
						fmt.Printf("Warning: Failed to cancel old TP order %d: %v\n", trade.TPOrderID, err)
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel old TP Order %d: %v", trade.TPOrderID, err))
					} else {
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Old TP Order %d canceled successfully.", trade.TPOrderID))
					}
				}

				if trade.SLOrderID > 0 {
					_, err := s.BinanceClient.CancelOrder(&binance.CancelOrderRequest{
						Symbol:  trade.Symbol,
						OrderID: trade.SLOrderID,
					})
					if err != nil {
						fmt.Printf("Warning: Failed to cancel old SL order %d: %v\n", trade.SLOrderID, err)
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Warning: Failed to cancel old SL Order %d: %v", trade.SLOrderID, err))
					} else {
						processResult.Logs = append(processResult.Logs, fmt.Sprintf("Old SL Order %d canceled successfully.", trade.SLOrderID))
					}
				}

				// Create TP/SL baru
				tpOrderID, slOrderID, err := s.tradeMonitorCreateTPOrder(trade, totalFilledQty, processResult)
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
func (s *Services) tradeMonitorCreateTPOrder(trade *models.Trade, totalQty float64, processResult *dtos.ProcessTradeResult) (int64, int64, error) {
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
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("Requesting new TP Order (Price: %v, Qty: %v, Side: %s)...", tpAdjusted, qtyAdjusted, closeSide))
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
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("FAILED setting Take Profit: %v", err))
		return 0, 0, fmt.Errorf("failed to place TP order: %w", err)
	}
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("SUCCESS setting Take Profit (OrderID: %d)", tpResp.OrderID))

	// Place Stop Loss Market
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("Requesting new SL Order (Price: %v, Qty: %v, Side: %s)...", slAdjusted, qtyAdjusted, closeSide))
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
		processResult.Logs = append(processResult.Logs, fmt.Sprintf("FAILED setting Stop Loss: %v", err))
		// TP sudah dibuat, tapi SL gagal - kita tetap return TP
		fmt.Printf("Warning: SL order failed but TP created: %v\n", err)
		return tpResp.OrderID, 0, nil
	}
	processResult.Logs = append(processResult.Logs, fmt.Sprintf("SUCCESS setting Stop Loss (OrderID: %d)", slResp.OrderID))

	return tpResp.OrderID, slResp.OrderID, nil
}

// tradeMonitorFase3Netting handles Fase 3: Netting & Finalisasi
func (s *Services) tradeMonitorFase3Netting(ctx *gin.Context, trade *models.Trade, processResult *dtos.ProcessTradeResult) error {
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
		processResult.Logs = append(processResult.Logs, "Zero entries filled and all entries cancelled/rejected. Marking trade as DEAD_SIGNAL (CANCELLED).")
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

	// Run Fase 3 Netting to get final TotalQty & AvgEntryPrice
	err = s.tradeMonitorFase3Netting(ctx, trade, result)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate netting: %w", err)
	}

	// Reload trade to get freshly netted qty
	trade, err = s.repo.Trade.FindWithEntries(nil, tradeID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload netted trade: %w", err)
	}

	// Fetch current market price (needed for exitPrice calculation)
	curPrice, err := s.BinanceClient.GetPrice(trade.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market price: %w", err)
	}
	exitPrice := curPrice.Price

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

