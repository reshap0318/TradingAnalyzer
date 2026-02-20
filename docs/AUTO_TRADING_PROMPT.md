# Prompt untuk Melanjutkan Auto-Trading Implementation

Gunakan prompt di bawah ini di conversation baru untuk melanjutkan implementasi auto-trading.

---

## Prompt

```
Saya ingin melanjutkan implementasi auto-trading untuk aplikasi Trading Analyzer saya. 

Ada implementation plan detail di file `docs/AUTO_TRADING_PLAN.md` — tolong baca terlebih dahulu.

Ringkasan aplikasi saat ini:
- Express.js API untuk analisis teknikal saham & crypto
- 7 indikator teknikal, multi-timeframe analysis (15m, 1h, 4h, 1D)
- Signal generation (BUY/SELL/WAIT) dengan confidence scoring
- Money management + capital tracking (available = initial - allocated + realized PnL)
- Signal logger otomatis dengan PnL tracking (dollar & percent)
- Futures calculator (leverage, liquidation, ROE)
- Data dari Binance API (sudah ada di src/crypto/binanceData.js)

Yang ingin saya tambahkan:
1. Modul `binanceTrader.js` untuk execute order via Binance API
2. Modul `tradeExecutor.js` sebagai safety layer + orchestrator
3. Endpoint POST /crypto/trade/auto (paper & live mode)
4. Endpoint GET /crypto/trade/status dan POST /crypto/trade/stop (kill switch)
5. Refactor analyze logic di app.js ke fungsi terpisah agar bisa dipanggil internal

Mulai dari Fase 1 (Paper Trading) dulu. Baca docs/AUTO_TRADING_PLAN.md untuk detail lengkap termasuk safety rules, arsitektur, dan verification plan.
```

---

## Catatan
- Pastikan signal accuracy sudah dicek sebelum implementasi
- Mulai dengan paper trading, jangan langsung live
- API keys Binance diperlukan (bisa pakai Testnet dulu)
