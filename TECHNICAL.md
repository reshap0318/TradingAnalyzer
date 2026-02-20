# 📖 Penjelasan Teknis — Trading Analyzer API

Dokumen ini menjelaskan detail teknis dari response API, logika signal log, position tracker, dan futures calculator.

---

## Penjelasan Response

Berikut penjelasan lengkap setiap field dalam response `/saham/analyze` dan `/crypto/analyze`.

### Contoh Response

```json
{
  "symbol": "BBCA.JK",
  "interval": "15m",
  "timestamp": "17/02/2026, 00.30.00",
  "capitalStatus": { ... },
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
  "entry": { "low": 8600, "high": 8650 },
  "tp": { "price": 9050, "percent": 4.9, "reason": "Risk 1.5x / Res" },
  "sl": { "price": 8400, "percent": -2.6, "reason": "S1 Support (Pivot)" },
  "riskReward": { "tp": 1.5 }
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
| `tp` | **Target Profit** — target taking profit. `percent` = berapa % keuntungan dari harga masuk |
| `sl` | **Stop Loss** — harga dimana kamu harus keluar untuk membatasi kerugian |
| `riskReward` | Rasio keuntungan vs kerugian. Misal `1.5` artinya potensi untung 1.5x lipat dari risiko rugi |

> **Tips**: Idealnya `riskReward.tp` minimal **1.5** — artinya potensi untung minimal 1.5x dari potensi rugi.

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

> [!NOTE]
> `isValid` bernilai `true` jika signal BUY (saham & crypto) atau SELL (crypto only) dan TP1 > SL. Saham hanya BUY karena tidak bisa short.

```json
"moneyManagement": {
  "isValid": true,
  "signal": "BUY",
  "totalLot": 5,
  "priceLot": {
    "satuan": 8625,
    "totalBelanja": 4312500
  },
  "tpPercent": 4.9,
  "slPercent": 2.6,
  "riskRewardRatio": 1.5,
  "maksimalKerugian": 112500,
  "potensiKeuntungan": 211312,
  "warnings": [
    "Mendekati batas maksimal kerugian"
  ]
}
```

| Field | Penjelasan |
|---|---|
| `totalLot` | Jumlah lot yang disarankan (1 lot = 100 lembar saham, crypto = unit) |
| `priceLot.satuan` | Harga beli per lembar aset |
| `priceLot.totalBelanja` | Rekomendasi total nilai uang yang di-belanjakan (position value) |
| `tpPercent` | Keuntungan kotor (%) bilamana TP tercapai |
| `slPercent` | Kerugian kotor (%) bilamana SL tercapai |
| `riskRewardRatio` | Rasio ganjaran banding resiko (idealnya >= 1.5) |
| `maksimalKerugian` | Kerugian nilai *fiat* modal maksimal bila kena SL |
| `potensiKeuntungan` | Keuntungan nilai *fiat* uang ekspektasi jika menyentuh target profit |

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

```text
1. User menembakkan POST /crypto/signals/log dengan payload sinyal BUY @ 67900
   → System insert ke active_trades: outcome = PENDING, tp = 69000, sl = 66000

2. Jam 02:00: Trigger `updateOutcomes` dipicu oleh sembarang call /analyze lain 
   → cek harga saat ini 67200 → tp & sl belum kena → status tetap PENDING

3. Jam 05:00: Trigger `updateOutcomes` dipicu lagi → Harga sekarang 69100
   → tp (69000) KENA! → outcome = TP_HIT, pnl = +1.62%

4. Jam 49:00: Jika 48 jam lewat tanpa TP/SL tersentuh → outcome = EXPIRED
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
| `TP_HIT` | Target profit tercapai ✅ |
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
  "totalPnlDollar": 7.62,
  "avgWinPercent": 2.1,
  "avgWinDollar": 1.05,
  "avgLossPercent": -1.5,
  "avgLossDollar": -0.75,
  "profitFactor": 2.45,
  "capitalStatus": {
    "initialCapital": 50,
    "allocated": 12.5,
    "realizedPnl": 7.62,
    "available": 45.12,
    "openPositions": 3
  }
}
```

| Field | Penjelasan |
|---|---|
| `totalPnlDollar` | Total keuntungan/kerugian dalam mata uang (USD/IDR) |
| `avgWinDollar` | Rata-rata keuntungan per win dalam mata uang |
| `avgLossDollar` | Rata-rata kerugian per loss dalam mata uang |
| `capitalStatus.initialCapital` | Modal awal (dari config / query `?capital=`) |
| `capitalStatus.allocated` | Modal yang sedang dialokasikan untuk trade PENDING |
| `capitalStatus.realizedPnl` | Total PnL dari trade yang sudah selesai |
| `capitalStatus.available` | Modal tersedia = initial − allocated + realizedPnl |
| `capitalStatus.openPositions` | Jumlah signal PENDING aktif |

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
  "tp": 69000,
  "notes": "Sinyal dari analyze tadi pagi"
}
```

### Contoh: Tutup Posisi

```json
POST /crypto/position/close
{
  "symbol": "BTCUSDT",
  "exitPrice": 68000,
  "reason": "TP"
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
    "tp": { "price": 67840, "pnl": 5.02, "roe": 5.02 }
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
| `targets.tp.roe` | **Return on Equity** — profit sebagai % dari margin |
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

