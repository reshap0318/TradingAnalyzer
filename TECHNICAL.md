# 📖 Penjelasan Teknis — Trading Analyzer API

Dokumen ini menjelaskan detail teknis dari response API, logika signal log, position tracker, dan futures calculator.

---

## Penjelasan Response

Berikut penjelasan lengkap setiap field dalam response `/saham/analyze` dan `/crypto/analyze`.

### Contoh Response

```json
{
  "symbol": "BBCA.JK",
  "timestamp": "17/02/2026, 00.30.00",
  "currentPrice": 8625,
  "change": -75,
  "changePercent": -0.86,
  "volume": 12500000,
  "trade_plan": { ... },
  "market_sentiment": { ... },
  "timeframes": { ... },
  "timeframeAlignment": "BULLISH_ALIGNED",
  "reasoning": [ ... ],
  "warnings": [ ... ],
  "moneyManagement": { ... },
  "patterns": { ... },
  "indicators": { ... }
}
```

---

### 📌 Info Dasar

| Field | Penjelasan |
|---|---|
| `symbol` | Kode saham/crypto yang dianalisis (contoh: `BBCA.JK`, `BTCUSDT`) |
| `timestamp` | Waktu analisis dilakukan (WIB) |
| `currentPrice` | Harga terakhir saat analisis |
| `change` | Selisih harga dari penutupan sebelumnya (dalam rupiah/USD) |
| `changePercent` | Perubahan harga dalam persen. Misal `-0.86` artinya turun 0.86% |
| `volume` | Jumlah saham/coin yang diperdagangkan hari ini |

---

### 📋 Trade Plan — Rencana Trading

**Ini adalah bagian terpenting** — berisi rekomendasi apakah sebaiknya beli, jual, atau tunggu.

```json
"trade_plan": {
  "valid": true,
  "signal": "BUY",
  "strength": "STRONG",
  "confidence": 72.5,
  "score": 72.5,
  "entry": { "low": 8600, "high": 8650 },
  "tp1": { "price": 8850, "percent": 2.6, "reason": "R1 Resistance (Pivot)" },
  "tp2": { "price": 9050, "percent": 4.9, "reason": "R2 Resistance (Pivot)" },
  "tp3": { "price": 9300, "percent": 7.8, "reason": "R3 Resistance (Fib)" },
  "sl":  { "price": 8400, "percent": -2.6, "reason": "S1 Support (Pivot)" },
  "riskReward": { "tp1": 1.5, "tp2": 3.1, "tp3": 5.0 }
}
```

| Field | Penjelasan |
|---|---|
| `valid` | `true` = sinyal layak dieksekusi, `false` = jangan ambil posisi |
| `signal` | **BUY** = beli, **SELL** = jual, **WAIT** = tunggu dulu |
| `strength` | Kekuatan sinyal: `STRONG` (kuat), `MODERATE` (sedang), `WEAK` (lemah) |
| `confidence` | Tingkat keyakinan dalam persen (0-100). Semakin tinggi semakin yakin |
| `score` | Skor mentah dari semua indikator. Positif = cenderung beli, negatif = cenderung jual |
| `entry` | **Zona masuk** — rentang harga ideal untuk membuka posisi |
| `tp1` | **Target Profit 1** — target terdekat. `percent` = berapa % keuntungan dari harga masuk |
| `tp2` | **Target Profit 2** — target menengah |
| `tp3` | **Target Profit 3** — target terjauh (paling optimis) |
| `sl` | **Stop Loss** — harga dimana kamu harus keluar untuk membatasi kerugian |
| `riskReward` | Rasio keuntungan vs kerugian. Misal `1.5` artinya potensi untung 1.5x lipat dari risiko rugi |

> **Tips**: Idealnya `riskReward.tp1` minimal **1.5** — artinya potensi untung minimal 1.5x dari potensi rugi.

---

### 🌍 Market Sentiment — Kondisi Pasar

Menunjukkan kondisi pasar secara keseluruhan (IHSG untuk saham, BTC untuk crypto).

```json
"market_sentiment": {
  "index": "IHSG",
  "trend": "BULLISH",
  "strength": 65,
  "change1d": 0.82,
  "change5d": 2.15,
  "isCrash": false,
  "correlation": 0.75,
  "details": ["EMA20 > EMA50", "Above SMA200", "RSI: 58.3"]
}
```

| Field | Penjelasan |
|---|---|
| `index` | Indeks yang dijadikan acuan: `IHSG` (saham) atau `BTC Market` (crypto) |
| `trend` | Arah pasar: `BULLISH` (naik), `BEARISH` (turun), `NEUTRAL` (sideways) |
| `strength` | Seberapa kuat tren pasar (0-100) |
| `change1d` | Perubahan indeks dalam 1 hari terakhir (%). Misal `0.82` = naik 0.82% |
| `change5d` / `change7d` | Perubahan dalam 5 hari (saham) atau 7 hari (crypto) |
| `isCrash` | `true` = pasar sedang **crash** (turun tajam). Hindari trading saat ini! |
| `correlation` | *(Saham only)* Korelasi harga saham dengan IHSG. `0.75` = cukup berkorelasi |
| `details` | Penjelasan teknikal kondisi pasar |

> **Tips**: Jika `isCrash = true`, sebaiknya **jangan trading** sampai pasar stabil.

---

### ⏰ Timeframes — Analisis Multi Waktu

Sistem menganalisis dari 4 sudut pandang waktu yang berbeda.

```json
"timeframes": {
  "15m": { "trend": "BULLISH", "signal": 45 },
  "1h":  { "trend": "BULLISH", "signal": 55 },
  "4h":  { "trend": "NEUTRAL", "signal": 10 },
  "1D":  { "trend": "BEARISH", "signal": -25 }
}
```

| Timeframe | Artinya |
|---|---|
| `15m` | Tren dalam 15 menit terakhir (sangat jangka pendek) |
| `1h` | Tren dalam 1 jam terakhir |
| `4h` | Tren dalam 4 jam terakhir |
| `1D` | Tren harian (gambaran besar) |

| Field | Penjelasan |
|---|---|
| `trend` | Arah tren di timeframe tersebut |
| `signal` | Skor sinyal (-100 s/d +100). Positif = bullish, negatif = bearish |

**`timeframeAlignment`** — Jika semua timeframe searah:
- `BULLISH_ALIGNED` = semua timeframe menunjukkan naik → sinyal beli lebih kuat
- `BEARISH_ALIGNED` = semua timeframe menunjukkan turun → sinyal jual lebih kuat
- `MIXED` = timeframe tidak sejalan → hati-hati, pasar belum jelas arahnya

---

### 📝 Reasoning & Warnings

```json
"reasoning": [
  "✓ MACD: Bullish crossover (Buy signal)",
  "✓ RSI: 55 - Bullish territory",
  "✗ Volume: Low volume, weak confirmation"
],
"warnings": [
  "Mendekati batas maksimal kerugian",
  "Risk/Reward ratio kurang dari 1.5"
]
```

| Field | Penjelasan |
|---|---|
| `reasoning` | Daftar alasan teknikal kenapa sinyal tersebut muncul. Tanda `✓` = mendukung sinyal, `✗` = berlawanan |
| `warnings` | Peringatan risiko yang perlu diperhatikan sebelum trading |

---

### 💰 Money Management — Manajemen Uang

Rekomendasi berapa banyak yang sebaiknya diinvestasikan.

```json
"moneyManagement": {
  "isValid": true,
  "signal": "BUY",
  "recommendation": {
    "lots": 5,
    "totalShares": 500,
    "positionValue": 4312500,
    "maxLossAmount": 62500,
    "maxLossPercent": 0.63
  },
  "analysis": {
    "riskPerShare": 125,
    "riskPerSharePercent": 1.45,
    "riskRewardRatio": 1.5
  },
  "potentialProfit": {
    "atTP1": 112500,
    "atTP2": 237500,
    "atTP3": 375000
  },
  "trailingStop": {
    "activationPrice": 8812,
    "distance": 62
  }
}
```

| Field | Penjelasan |
|---|---|
| **recommendation** | |
| `lots` | Jumlah lot yang disarankan (1 lot = 100 lembar saham, crypto = unit) |
| `positionValue` | Total nilai posisi dalam Rupiah / USD |
| `maxLossAmount` | Kerugian maksimal jika kena stop loss |
| `maxLossPercent` | Kerugian maksimal sebagai % dari total modal |
| **analysis** | |
| `riskPerShare` | Risiko per lembar saham (selisih harga masuk dan stop loss) |
| `riskRewardRatio` | Rasio potensi untung vs rugi |
| **potentialProfit** | |
| `atTP1` / `atTP2` / `atTP3` | Potensi keuntungan di masing-masing target |
| **trailingStop** | |
| `activationPrice` | Harga di mana trailing stop aktif (mengunci keuntungan) |
| `distance` | Jarak trailing stop dari harga tertinggi |

---

### 📊 Indicators — Indikator Teknikal

| Indikator | Apa artinya? |
|---|---|
| **MA** (Moving Average) | Tren harga rata-rata. Jika harga di atas MA = tren naik |
| **RSI** (Relative Strength Index) | Mengukur apakah harga sudah terlalu mahal (>70) atau terlalu murah (<30) |
| **MACD** | Mengukur momentum. Crossover ke atas = sinyal beli, ke bawah = sinyal jual |
| **BB** (Bollinger Bands) | Mengukur volatilitas. Harga di batas atas = mungkin akan turun |
| **Stoch** (Stochastic) | Mirip RSI, tapi lebih sensitif. >80 = overbought, <20 = oversold |
| **Volume** | Apakah volume perdagangan mendukung pergerakan harga |

---

### 🕯️ Patterns — Pola Candlestick

```json
"patterns": {
  "1D": [
    { "date": "17/02/2026", "patterns": ["Bullish Engulfing", "Hammer"] }
  ]
}
```

| Pola | Artinya |
|---|---|
| **Bullish Engulfing** | Candle hijau besar "menelan" candle merah sebelumnya → potensi naik |
| **Hammer** | Sumbu bawah panjang → pembeli mulai masuk → potensi naik |
| **Morning Star** | Pola 3 candle yang menandakan pembalikan dari turun ke naik |
| **Bearish Engulfing** | Candle merah besar "menelan" candle hijau → potensi turun |
| **Shooting Star** | Sumbu atas panjang → penjual mulai masuk → potensi turun |
| **Doji** | Badan candle sangat kecil → pasar ragu-ragu, bisa berbalik arah |
| **Marubozu** | Candle tanpa sumbu → tren sangat kuat ke satu arah |

---

## 📊 Signal Log — Detail

### Alur Data

```
Jam 01:00   Cron hit /crypto/analyze → Sinyal SELL @ 67900
            → Auto-log: outcome = PENDING, tp1 = 66500, sl = 69000

Jam 02:00   Cron hit lagi (sinyal masih SELL) → Skip (sudah ada PENDING SELL)
            → updateOutcomes: cek harga 67200 → tp1 & sl belum kena → tetap PENDING

Jam 05:00   Cron hit lagi → Harga sekarang 66400
            → tp1 (66500) KENA! → outcome = TP1_HIT, pnl = +2.06%

Jam 49:00   Jika 48 jam lewat tanpa TP/SL → outcome = EXPIRED
```

### Aturan Dedup (1 PENDING per simbol)

| Hit | Sinyal | Yang Terjadi |
|---|---|---|
| Hit 1 | BUY | ✅ Log baru → PENDING |
| Hit 2 | BUY | ⏭️ Skip — sudah ada PENDING BUY untuk simbol ini |
| Hit 3 | SELL | 🔄 PENDING BUY ditutup → `SIGNAL_REVERSED` + Log SELL baru |
| Hit 4 | WAIT | ⏭️ WAIT tidak dicatat, PENDING SELL tetap jalan |

### Daftar Outcome

| Status | Artinya |
|---|---|
| `PENDING` | Sinyal baru, menunggu TP/SL tercapai |
| `TP1_HIT` | Target profit 1 tercapai ✅ |
| `TP2_HIT` | Target profit 2 tercapai ✅✅ |
| `TP3_HIT` | Target profit 3 tercapai ✅✅✅ |
| `SL_HIT` | Stop loss kena ❌ |
| `SIGNAL_REVERSED` | Sinyal berbalik (BUY→SELL / SELL→BUY) 🔄 |
| `EXPIRED` | 48 jam (crypto) / 10 hari (saham) lewat tanpa TP/SL ⏰ |

### Contoh Summary Response

```json
{
  "totalSignals": 47,
  "pending": 3,
  "completed": 44,
  "wins": 28,
  "losses": 12,
  "reversed": 4,
  "expired": 4,
  "winRate": 70,
  "totalPnlPercent": 15.23,
  "avgWinPercent": 2.1,
  "avgLossPercent": -1.5,
  "profitFactor": 2.45
}
```

> `profitFactor` > 1.5 dan `winRate` > 60% menandakan script cukup reliable.

---

## 📍 Position Tracker — Detail

### Alur Pemakaian

```
1. Lihat sinyal BUY dari /crypto/analyze
2. Eksekusi trade di exchange
3. Panggil POST /crypto/position/open   → catat posisi
4. Saat mau keluar, panggil POST /crypto/position/close  → hitung PnL
```

### Contoh: Buka Posisi

```json
POST /crypto/position/open
{
  "symbol": "BTCUSDT",
  "side": "LONG",
  "entryPrice": 67500,
  "quantity": 0.01,
  "sl": 66000,
  "tp1": 69000,
  "notes": "Sinyal dari analyze tadi pagi"
}
```

### Contoh: Tutup Posisi

```json
POST /crypto/position/close
{
  "symbol": "BTCUSDT",
  "exitPrice": 68000,
  "reason": "TP1"
}
// Response: { pnlPercent: +0.74%, pnlTotal: 5 USD }
```

---

## 🚀 Futures — Detail (Crypto Only)

Jika sinyal BUY atau SELL muncul, response crypto otomatis menyertakan field `futures` — kalkulasi leverage, liquidation price, dan ROE per target.

### Cara Pakai

Tambahkan parameter `leverage` ke URL:
```
GET /crypto/analyze?symbol=BTCUSDT&capital=1000&leverage=10
```

Default leverage dari config (5x) jika tidak diisi.

### Contoh Response

```json
"futures": {
  "side": "SHORT",
  "leverage": 10,
  "margin": { "required": 100, "maintenanceRate": "0.5%" },
  "position": { "notionalValue": 1000, "quantity": 0.01466, "entryPrice": 68182 },
  "liquidation": {
    "price": 74659,
    "distancePercent": 9.5,
    "slBeyondLiquidation": false
  },
  "risk": {
    "slPrice": 68410,
    "slPercent": 0.33,
    "slLoss": 3.35,
    "slLossOfCapital": "0.33%",
    "effectiveRisk": "3.35% ROE"
  },
  "targets": {
    "tp1": { "price": 67840, "pnl": 5.02, "roe": 5.02 },
    "tp2": { "price": 67497, "pnl": 10.04, "roe": 10.04 },
    "tp3": { "price": 67041, "pnl": 16.73, "roe": 16.73 }
  },
  "fees": { "openFee": 0.4, "closeFee": 0.4, "totalFees": 0.8, "fundingPer8h": 0.1 },
  "warnings": []
}
```

### Penjelasan Field

| Field | Penjelasan |
|---|---|
| `leverage` | Pengali posisi. 10x = posisi 10x lebih besar dari margin |
| `margin.required` | Uang yang dikunci sebagai jaminan |
| `position.notionalValue` | Nilai total posisi (margin × leverage) |
| `liquidation.price` | **Harga liquidasi** — posisi otomatis ditutup dan margin hangus |
| `liquidation.distancePercent` | Jarak dari entry ke liquidation (%) |
| `liquidation.slBeyondLiquidation` | `true` = **BAHAYA**: SL melewati harga liquidasi |
| `risk.effectiveRisk` | Risiko sebenarnya (slPercent × leverage) |
| `targets.tp1.roe` | **Return on Equity** — profit sebagai % dari margin |
| `fees.totalFees` | Total biaya buka + tutup posisi |
| `fees.fundingPer8h` | Estimasi biaya funding rate setiap 8 jam |

> **Tips**: Jika `slBeyondLiquidation = true`, posisi akan di-liquidate SEBELUM stop loss. Kurangi leverage!

---

## Perbedaan Saham vs Crypto

| Aspek | Saham | Crypto |
|---|---|---|
| Mata uang / Capital | **Rupiah** (IDR) | **Dolar** (USD) |
| Timeframe utama | 1D (harian) | 1H (per jam) |
| Sentimen pasar | IHSG | BTC Market |
| Sinyal valid | BUY saja (SELL muncul tapi tidak dicatat ke log) | BUY dan SELL |
| Max stop loss | 6% | 8% |
| Presisi harga | 0 desimal | Hingga 8 desimal |
| Satuan | Lot (1 lot = 100 lembar) | Unit (bisa pecahan) |
| Signal expiry | 10 hari | 48 jam |
| Futures | ❌ Tidak tersedia | ✅ Tersedia |

