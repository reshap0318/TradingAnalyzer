# SignalAnalyze Response Documentation

Dokumentasi lengkap untuk response dari endpoint **SignalAnalyze**.

---

## 📋 Overview

**Endpoint:** `POST /api/v1/signal/analyze`

**Fungsi:** Menganalisis market data dan menghasilkan trading signal dengan trading plan lengkap.

**Response Structure:**
```
SignalAnalyzeResponse
├── symbol (string)
├── primary_timeframe (string)
├── timestamp (time.Time)
├── signal (SignalInfo)
│   ├── valid (bool)
│   ├── signal (string)
│   ├── current_price (float64)
│   └── trading_plan (TradingPlan)
│       ├── mode (string)
│       ├── entries (array)
│       ├── take_profit (float64)
│       ├── stop_loss (float64)
│       ├── risk_reward_ratio (float64)
│       ├── buffer_percent (float64)
│       └── summary (TradingPlanSummary)
└── scoring (ScoringBreakdown)
    ├── totalScore (float64)
    ├── confidence (float64)
    └── breakdown (array)
```

---

## 📊 Response Fields

### **1. Root Level**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `symbol` | string | Symbol yang dianalisis | `"BTCUSDT"` |
| `primary_timeframe` | string | Timeframe utama untuk entry | `"1h"` |
| `timestamp` | time.Time | Waktu analisis | `"2025-03-10T14:30:00Z"` |
| `signal` | SignalInfo | Informasi signal trading | - |
| `scoring` | ScoringBreakdown | Breakdown scoring signal | - |

---

### **2. SignalInfo**

Informasi utama signal trading.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `valid` | bool | Apakah signal valid (confidence >= threshold) | `true` |
| `signal` | string | Tipe signal: `BUY`, `SELL`, `STRONG_BUY`, `STRONG_SELL`, `WAIT` | `"BUY"` |
| `current_price` | float64 | Harga current symbol | `50000.00` |
| `trading_plan` | TradingPlan | Rencana trading lengkap | - |

**Signal Types:**
| Signal | Strength | Capital Allocation |
|--------|----------|-------------------|
| `BUY` | 80% | 80% dari MAX_POSITION_SIZE |
| `SELL` | 80% | 80% dari MAX_POSITION_SIZE |
| `STRONG_BUY` | 100% | 100% dari MAX_POSITION_SIZE |
| `STRONG_SELL` | 100% | 100% dari MAX_POSITION_SIZE |
| `WAIT` | 0% | No trade |

---

### **3. TradingPlan**

Rencana trading lengkap dengan entry, TP, SL, dan risk management.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `mode` | string | Mode trading: `"CONSERVATIVE"` atau `"AGGRESSIVE"` | `"CONSERVATIVE"` |
| `entries` | array | List entry points | `[...]` |
| `take_profit` | float64 | Harga Take Profit | `50750.00` |
| `stop_loss` | float64 | Harga Stop Loss | `48511.00` |
| `risk_reward_ratio` | float64 | Rasio Risk:Reward | `0.51` |
| `buffer_percent` | float64 | Buffer yang digunakan untuk S/R | `1.50` |
| `summary` | TradingPlanSummary | **Pre-calculated summary data** ⭐ | - |

**Trading Modes:**
| Mode | Entries | Strategy |
|------|---------|----------|
| `CONSERVATIVE` | 1 entry | Entry di harga terbaik (support/resistance + buffer) |
| `AGGRESSIVE` | 2 entries | 50% di current price + 50% di pullback |

---

### **4. TradingPlanEntry** (Array)

Setiap entry point dalam trading plan.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `entry_number` | int | Urutan entry (1, 2, ...) | `1` |
| `entry_price` | float64 | Harga entry | `50000.00` |
| `position_size` | string | Ukuran posisi dalam persen | `"100%"` atau `"50%"` |
| `position_value` | float64 | Nilai posisi dalam USDT (modal kamu) | `400.00` |
| `position_qty` | float64 | Jumlah coin yang dibeli | `0.04000000` |

**Contoh (Aggressive Mode - 2 entries):**
```json
"entries": [
  {
    "entry_number": 1,
    "entry_price": 50000.00,
    "position_size": "50%",
    "position_value": 200.00,
    "position_qty": 0.02000000
  },
  {
    "entry_number": 2,
    "entry_price": 49990.00,
    "position_size": "50%",
    "position_value": 200.00,
    "position_qty": 0.02000400
  }
]
```

---

### **5. TradingPlanSummary** ⭐

**Pre-calculated summary data** untuk memudahkan analisis tanpa perlu recalculate.

#### **5.1 Position Info**

| Field | Type | Description | Rumus | Example |
|-------|------|-------------|-------|---------|
| `total_entries` | int | Jumlah total entry | - | `1` |
| `total_position_value` | float64 | **Total modal yang dipakai (USDT)** | Sum dari semua `position_value` | `400.00` |
| `total_position_qty` | float64 | Total quantity coin | Sum dari semua `position_qty` | `0.04000000` |
| `avg_entry_price` | float64 | Harga entry rata-rata (weighted) | `(Σ value×price) / Σ value` | `50000.00` |

**Contoh Perhitungan:**
```
Balance: $1,000
MAX_POSITION_SIZE: 50%
Signal: BUY (80% strength)

total_position_value = $1,000 × 0.5 × 0.8 = $400
```

---

#### **5.2 Risk Info** ⚠️

| Field | Type | Description | Rumus | Example |
|-------|------|-------------|-------|---------|
| `max_risk_usdt` | float64 | **Kerugian maksimal dalam USDT** jika SL hit | `(avg_entry - SL) × qty` (BUY) | `60.00` |
| `max_risk_percent` | float64 | Risk sebagai % dari **position value** | `(max_risk_usdt / position_value) × 100` | `3.00` |
| `risk_from_capital` | float64 | **Risk sebenarnya** sebagai % dari **modal yang dipakai** | `(max_risk_usdt / total_position_value) × 100` | `15.00` |

**Perbedaan Penting:**
- `max_risk_percent` = Jarak SL dari entry (**tidak termasuk leverage**)
- `risk_from_capital` = Kerugian sebenarnya ke portfolio kamu (**sudah termasuk leverage**)

**Contoh Perhitungan:**
```
Entry: $50,000
SL: $48,511
Distance: 3%

Position Value (dengan leverage 5x): $2,000
Modal kamu: $400

max_risk_percent = $60 / $2,000 × 100 = 3%
risk_from_capital = $60 / $400 × 100 = 15% ⭐
```

**Arti `risk_from_capital = 15%`:**
> Jika SL hit, kamu akan kehilangan **15% dari modal $400** yang dipakai untuk trade ini (= $60).

---

#### **5.3 Profit Info** 💰

| Field | Type | Description | Rumus | Example |
|-------|------|-------------|-------|---------|
| `target_profit_usdt` | float64 | **Profit yang diharapkan dalam USDT** jika TP hit | `(TP - avg_entry) × qty` (BUY) | `30.00` |
| `target_profit_percent` | float64 | Profit sebagai % dari **position value** | `(target_profit_usdt / position_value) × 100` | `1.50` |
| `profit_from_capital` | float64 | **Profit sebenarnya** sebagai % dari **modal yang dipakai** | `(target_profit_usdt / total_position_value) × 100` | `7.50` |

**Perbedaan Penting:**
- `target_profit_percent` = Jarak TP dari entry (**tidak termasuk leverage**)
- `profit_from_capital` = Return sebenarnya ke portfolio kamu (**sudah termasuk leverage**)

**Contoh Perhitungan:**
```
Entry: $50,000
TP: $50,750
Distance: 1.5%

Position Value (dengan leverage 5x): $2,000
Modal kamu: $400

target_profit_percent = $30 / $2,000 × 100 = 1.5%
profit_from_capital = $30 / $400 × 100 = 7.5% ⭐
```

**Arti `profit_from_capital = 7.5%`:**
> Jika TP hit, kamu akan mendapatkan **7.5% dari modal $400** yang dipakai untuk trade ini (= $30).

---

#### **5.4 Leverage Info**

| Field | Type | Description | Rumus | Example |
|-------|------|-------------|-------|---------|
| `effective_leverage` | float64 | **Leverage aktual** yang digunakan | `position_value / total_position_value` | `5.00` |

**Arti:**
- Menunjukkan berapa kali modal kamu di-leverage
- Bisa berbeda dari config leverage jika ada limitasi exchange
- Jika `effective_leverage = 5x`, berarti modal $400 jadi position $2,000

---

### **6. ScoringBreakdown**

Breakdown scoring signal dari berbagai timeframe dan indicator.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `totalScore` | float64 | Total score dari semua timeframe | `0.75` |
| `confidence` | float64 | Confidence level (0-100) | `75.00` |
| `breakdown` | array | Breakdown per timeframe | `[...]` |

---

### **7. TimeframeSignalData** (Array dalam breakdown)

Signal breakdown per timeframe.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timeframe` | string | Nama timeframe | `"1h"` |
| `trend` | string | Trend kategori dari threshold | `"BULLISH"` |
| `rawSignal` | float64 | Raw signal value sebelum threshold | `0.65` |
| `weight` | float64 | Weight timeframe ini | `0.5` |
| `contribution` | float64 | Kontribusi ke total score | `0.325` |
| `indicator` | array | Breakdown indicator | `[...]` |

---

### **8. IndicatorBreakdown** (Array dalam indicator)

Detail per indicator dalam timeframe.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `name` | string | Nama indicator | `"RSI"` |
| `rawSignal` | int | Raw signal: `1` (bullish), `-1` (bearish), `0` (neutral) | `1` |
| `weight` | float64 | Weight indicator ini | `0.3` |
| `contribution` | float64 | Kontribusi ke timeframe signal | `0.30` |
| `details` | array | Detail strings dari indicator | `["RSI: 65.5 (Bullish)"]` |
| `value` | interface{} | Value utama indicator | `65.5` |
| `zone` | string | Zone indicator (jika ada) | `"Bullish"` |

---

## 📝 Contoh Response Lengkap

```json
{
  "symbol": "BTCUSDT",
  "primary_timeframe": "1h",
  "timestamp": "2025-03-10T14:30:00Z",
  "signal": {
    "valid": true,
    "signal": "BUY",
    "current_price": 50000.00,
    "trading_plan": {
      "mode": "CONSERVATIVE",
      "entries": [
        {
          "entry_number": 1,
          "entry_price": 49990.00,
          "position_size": "100%",
          "position_value": 400.00,
          "position_qty": 0.04000800
        }
      ],
      "take_profit": 50750.00,
      "stop_loss": 48511.00,
      "risk_reward_ratio": 0.51,
      "buffer_percent": 1.50,
      "summary": {
        "total_entries": 1,
        "total_position_value": 400.00,
        "total_position_qty": 0.04000800,
        "avg_entry_price": 49990.00,
        
        "max_risk_usdt": 59.16,
        "max_risk_percent": 2.96,
        "risk_from_capital": 14.79,
        
        "target_profit_usdt": 30.40,
        "target_profit_percent": 1.52,
        "profit_from_capital": 7.60,
        
        "effective_leverage": 5.00
      }
    }
  },
  "scoring": {
    "totalScore": 0.75,
    "confidence": 75.00,
    "breakdown": [
      {
        "timeframe": "1h",
        "trend": "BULLISH",
        "rawSignal": 0.65,
        "weight": 0.6,
        "contribution": 0.39,
        "indicator": [
          {
            "name": "RSI",
            "rawSignal": 1,
            "weight": 0.3,
            "contribution": 0.30,
            "details": ["RSI: 65.5 (Bullish)"],
            "value": 65.5,
            "zone": "Bullish"
          },
          {
            "name": "MACD",
            "rawSignal": 1,
            "weight": 0.4,
            "contribution": 0.40,
            "details": ["MACD: Bullish Crossover"],
            "value": 125.3,
            "zone": "Bullish"
          }
        ]
      },
      {
        "timeframe": "4h",
        "trend": "BULLISH",
        "rawSignal": 0.80,
        "weight": 0.4,
        "contribution": 0.32,
        "indicator": [
          {
            "name": "RSI",
            "rawSignal": 1,
            "weight": 0.3,
            "contribution": 0.30,
            "details": ["RSI: 70.2 (Bullish)"],
            "value": 70.2,
            "zone": "Bullish"
          }
        ]
      }
    ]
  }
}
```

---

## 🎯 Cara Menggunakan Response

### **1. Quick Assessment (Cepat)**

```javascript
const response = await analyzeSignal();

// Cek apakah signal valid
if (!response.signal.valid) {
  console.log("Signal tidak valid, skip trade");
  return;
}

// Cek signal type
if (response.signal.signal === "WAIT") {
  console.log("Menunggu setup yang lebih baik");
  return;
}

// Cek risk dari modal
const riskPercent = response.signal.trading_plan.summary.risk_from_capital;
if (riskPercent > 20) {
  console.log("Risk terlalu tinggi!");
  return;
}

// Cek R:R
const rr = response.signal.trading_plan.risk_reward_ratio;
if (rr < 0.5) {
  console.log("R:R terlalu kecil");
  return;
}

// All checks passed, ready to trade!
console.log("Signal OK, siap untuk execute!");
```

---

### **2. Display ke User (UI)**

```
📊 BTCUSDT - BUY Signal
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Signal Valid | Confidence: 75%
💰 Current Price: $50,000

📋 Trading Plan
├─ Mode: CONSERVATIVE
├─ Entry: $49,990
├─ Take Profit: $50,750 (+$30.40 / +7.60%)
├─ Stop Loss: $48,511 (-$59.16 / -14.79%)
├─ Risk/Reward: 0.51
└─ Leverage: 5x

💵 Position
├─ Capital Used: $400.00
├─ Position Size: 0.04000800 BTC
└─ Effective Value: $2,000.00

⏱️ Timeframes
├─ 1h: BULLISH (65% weight)
│  ├─ RSI: 65.5 (Bullish)
│  └─ MACD: Bullish Crossover
└─ 4h: BULLISH (40% weight)
   └─ RSI: 70.2 (Bullish)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Ready to Execute Trade
```

---

### **3. Risk Management Check**

```javascript
const summary = response.signal.trading_plan.summary;

// Risk checks
const checks = {
  riskOk: summary.risk_from_capital <= 20,      // Max 20% risk per trade
  profitOk: summary.profit_from_capital >= 5,   // Min 5% profit
  rrOk: response.signal.trading_plan.risk_reward_ratio >= 0.5,
  leverageOk: summary.effective_leverage <= 10, // Max 10x leverage
};

const allPassed = Object.values(checks).every(v => v);

if (allPassed) {
  executeTrade();
} else {
  console.log("Risk checks failed:", checks);
}
```

---

## ⚠️ Important Notes

### **1. `risk_from_capital` vs `max_risk_percent`**

```
❌ SALAH: Menggunakan max_risk_percent untuk risk assessment
   max_risk_percent = 3% → "Oh risk cuma 3%, aman!"
   (Padahal leverage 5x, risk sebenarnya 15%)

✅ BENAR: Gunakan risk_from_capital
   risk_from_capital = 15% → "Risk 15% dari modal, perlu dipertimbangkan"
```

### **2. Signal Strength**

```
BUY/SELL (80% strength):
  - Capital digunakan = 80% dari MAX_POSITION_SIZE
  - Untuk trade dengan confidence medium

STRONG_BUY/STRONG_SELL (100% strength):
  - Capital digunakan = 100% dari MAX_POSITION_SIZE
  - Untuk trade dengan confidence tinggi
```

### **3. Summary adalah Pre-calculated**

```
✅ GOOD: Langsung akses summary
  const risk = response.signal.trading_plan.summary.risk_from_capital;

❌ BAD: Recalculate manual
  let totalValue = 0;
  for (const entry of entries) {
    totalValue += entry.position_value;
  }
```

---

## 📚 Related Documentation

- [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API endpoints lengkap
- [SIGNAL_BREAKDOWN.md](./SIGNAL_BREAKDOWN.md) - Cara kerja signal calculation
- [CODING_RULES.md](./CODING_RULES.md) - Coding conventions

---

## 📞 Support

Untuk pertanyaan atau issue terkait SignalAnalyze response, silakan buat issue di repository atau hubungi development team.
