# 🤖 Auto-Trading Implementation Plan — Binance Integration

> **Status**: Planning — belum diimplementasi
> **Target**: Crypto auto-trading via Binance API (Spot & Futures)
> **Prerequisite**: Signal accuracy sudah dicek dan cukup reliable

---

## Arsitektur

```
┌──────────────────────────────────────────────────────────────┐
│                     app.js (Express)                        │
├──────────────────────────────────────────────────────────────┤
│  GET  /crypto/analyze          ← Read-only (tetap ada)      │
│  POST /crypto/trade/auto       ← NEW: Auto-trade endpoint   │
│  GET  /crypto/trade/status     ← NEW: Cek status bot        │
│  POST /crypto/trade/stop       ← NEW: Kill switch           │
└──────────────┬───────────────────────────────────────────────┘
               │
    ┌──────────▼──────────┐
    │   tradeExecutor.js  │  ← Orchestrator (safety + decision)
    │   (shared/)         │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  binanceTrader.js   │  ← Execute orders via Binance API
    │  (crypto/)          │
    └─────────────────────┘
```

### Flow Diagram

```
POST /crypto/trade/auto { symbol, mode: "live"|"paper" }
  │
  ├─ 1. analyzeSymbol(symbol)        ← Reuse existing analyze logic
  │     ├─ fetchMultiTimeframe
  │     ├─ analyzeBTCMarket
  │     ├─ makeCryptoDecision
  │     ├─ generateSignal
  │     ├─ calculateTPSL
  │     └─ calculateMoneyManagement
  │
  ├─ 2. Safety Checks (tradeExecutor)
  │     ├─ signal !== "WAIT"?
  │     ├─ confidence >= threshold (configurable, default 70)?
  │     ├─ moneyMgmt.isValid?
  │     ├─ capitalStatus.available > 0?
  │     ├─ dailyTradeCount < maxDailyTrades?
  │     ├─ drawdown < maxDrawdown?
  │     └─ isCrash === false?
  │
  ├─ 3. Execute (binanceTrader)
  │     ├─ mode "paper" → simulate only, log result
  │     └─ mode "live"  → Binance API order
  │           ├─ Market/Limit order (entry)
  │           ├─ Stop-Loss order
  │           └─ Take-Profit order (OCO/TP-SL)
  │
  └─ 4. Response
        ├─ analysis (signal, confidence, tpsl)
        ├─ execution (orderId, status, filled price)
        └─ capitalStatus (updated)
```

---

## File-by-File Implementation

### 1. [NEW] `src/crypto/binanceTrader.js`

Modul untuk interact dengan Binance API.

```js
// Dependencies: binance-api-node (npm install binance-api-node)
// Environment: BINANCE_API_KEY, BINANCE_API_SECRET di .env

// Functions:
export async function initClient(testnet = false)
export async function getAccountBalance()
export async function placeSpotOrder({ symbol, side, quantity, type })
export async function placeFuturesOrder({ symbol, side, quantity, leverage, type })
export async function setStopLoss({ symbol, side, stopPrice, quantity })
export async function setTakeProfit({ symbol, side, price, quantity })
export async function cancelAllOrders(symbol)
export async function getOpenOrders(symbol)
export async function getOrderStatus(orderId)
```

**Key considerations:**
- Support both **Testnet** dan **Production** via config flag
- Retry logic untuk network errors
- Rate limiting (max 1200 req/min)
- Error handling untuk insufficient balance, min notional, dll

---

### 2. [NEW] `src/shared/tradeExecutor.js`

Safety layer + orchestrator. Menggabungkan analisis → safety check → execution.

```js
// Functions:
export async function executeTrade({ symbol, mode, capital, leverage })
// → Calls analyze logic, safety checks, then execute

export function checkSafetyRules(analysis, capitalStatus, config)
// → Returns { safe: boolean, reasons: string[] }

export function getTradingStatus()
// → Returns { active: boolean, todayTrades: number, drawdown: number }

export function stopTrading(reason)
// → Kill switch — sets active = false
```

**Safety Rules (configurable in config.js):**

| Rule | Default | Penjelasan |
|---|---|---|
| `minConfidence` | 70 | Minimum confidence score |
| `maxDailyTrades` | 5 | Max trade per hari |
| `maxDrawdownPercent` | 10 | Stop kalau rugi > 10% dari initial |
| `requireValidMM` | true | moneyMgmt.isValid harus true |
| `blockOnCrash` | true | Tidak trade saat market crash |
| `minRiskReward` | 1.5 | Minimum risk/reward ratio |

---

### 3. [MODIFY] `src/config.js`

Tambah section `AUTO_TRADING`:

```js
AUTO_TRADING: {
  ENABLED: false,                    // Master switch
  MODE: "paper",                     // "paper" | "live"
  USE_TESTNET: true,                 // Binance Testnet
  SAFETY: {
    MIN_CONFIDENCE: 70,
    MAX_DAILY_TRADES: 5,
    MAX_DRAWDOWN_PERCENT: 10,
    MIN_RISK_REWARD: 1.5,
    BLOCK_ON_CRASH: true,
  },
  ORDER_TYPE: "MARKET",              // "MARKET" | "LIMIT"
  DEFAULT_LEVERAGE: 5,
  AUTO_SET_TP_SL: true,              // Auto place TP/SL orders
}
```

---

### 4. [NEW] `.env`

```
BINANCE_API_KEY=your_api_key_here
BINANCE_API_SECRET=your_api_secret_here
```

> ⚠️ Tambahkan `.env` ke `.gitignore`!

---

### 5. [MODIFY] `src/app.js`

Tambah 3 endpoint baru:

```js
// Auto-trade endpoint
POST /crypto/trade/auto
  Body: { symbol: "BTCUSDT", mode: "paper"|"live" }
  Response: { analysis, execution, capitalStatus }

// Trading status
GET /crypto/trade/status
  Response: { active, mode, todayTrades, drawdown, lastTrade }

// Kill switch
POST /crypto/trade/stop
  Body: { reason: "manual stop" }
  Response: { stopped: true }
```

---

### 6. [MODIFY] `src/app.js` — Refactor analyze logic

Extract core analyze logic ke fungsi terpisah agar bisa dipanggil internal:

```js
// Before: semua logic di dalam app.get("/crypto/analyze", ...)
// After:  extract ke async function analyzeSymbol(symbol, capital)

export async function analyzeSymbol(symbol, capital) {
  // ... existing analyze logic ...
  return { quote, decision, signalResult, tpsl, moneyMgmt, btcMarket };
}

// Endpoint tetap pakai fungsi ini:
app.get("/crypto/analyze", async (req, res) => {
  const result = await analyzeSymbol(symbol, capital);
  res.json(formatResponse(result));
});
```

---

## Dependencies

```bash
npm install binance-api-node dotenv
```

| Package | Fungsi |
|---|---|
| `binance-api-node` | Official Binance API wrapper |
| `dotenv` | Load API keys dari .env |

---

## Fase Implementasi

### Fase 1: Paper Trading (minggu pertama)
1. Buat `binanceTrader.js` (connect ke Testnet)
2. Buat `tradeExecutor.js` (safety rules)
3. Buat endpoint `/crypto/trade/auto` (mode paper only)
4. Test: pastikan paper trades di-log dengan benar

### Fase 2: Testnet Live (minggu kedua)
1. Connect ke Binance Testnet
2. Execute real orders di Testnet
3. Verify: order tereksekusi, TP/SL terpasang, balance berubah
4. Monitor selama beberapa hari

### Fase 3: Production (setelah yakin)
1. Switch ke Production API keys
2. Mulai dengan capital kecil
3. Monitor closely

---

## Verification Plan

### Unit Tests
- `tests/binanceTrader.test.js` — Mock API calls
- `tests/tradeExecutor.test.js` — Safety rules, edge cases

### Integration Tests
- Paper trade → cek signal_log updated correctly
- Testnet → cek order actually placed
- Kill switch → verify trading stops immediately

### Manual Testing
1. `POST /crypto/trade/auto` dengan mode "paper" → verify response
2. Check signal_log.json → entry harus ada `allocatedAmount` dan `executionDetails`
3. Run beberapa kali → verify max daily trades limit works
4. Simulate crash condition → verify trade blocked

---

## Risiko & Mitigasi

| Risiko | Mitigasi |
|---|---|
| API key leak | `.env` + `.gitignore`, IP whitelist di Binance |
| Over-trading | `maxDailyTrades` limit |
| Large loss | `maxDrawdownPercent` kill switch |
| Network error saat order | Retry logic + order status verification |
| Price slippage | Limit order option, min R:R check |
| Binance downtime | Graceful error handling, skip trade |
