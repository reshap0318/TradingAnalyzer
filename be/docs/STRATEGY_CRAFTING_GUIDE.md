# Panduan Meracik Strategi Trading (Strategy Crafting Guide)

Dokumen ini menjelaskan cara membuat dan mengkonfigurasi sebuah **Strategy** dari nol hingga siap dioperasikan oleh Bot maupun Backtester.

---

## 1. Anatomi Sebuah Strategi

Setiap strategi terdiri dari **5 komponen** utama:

```
Strategy
├── Primary Timeframe   (1 value: e.g. "15m")
├── Timeframes[]        (Multi-TF dengan bobot masing-masing)
├── Indicators[]        (Indikator teknikal + peran + bobot + TF target)
├── Money Management[]  (Parameter manajemen risiko)
└── Symbols[]           (Koin yang akan diperdagangkan)
```

---

## 2. Komponen Detail

### 2.1 Primary Timeframe

Timeframe utama yang digunakan sebagai acuan iterasi candle oleh mesin. Tabel di bawah **hanya contoh ilustrasi, bukan patokan baku** — Anda bebas menentukan Primary TF sesuai kebutuhan dan eksperimen Anda sendiri:

| Gaya         | Primary TF | Cocok Untuk                      |
|:-------------|:-----------|:---------------------------------|
| Scalping     | `1m`/`5m`  | Profit cepat, risiko tinggi      |
| Day Trading  | `15m`      | Intraday, seimbang kecepatan     |
| Swing        | `1h`       | Tren menengah, sinyal berkualitas |
| Position     | `4h`/`1d`  | Tren panjang, frekuensi rendah   |

### 2.2 Timeframes (Multi-TF Analysis)

Tentukan **timeframe mana saja** yang akan dianalisis secara paralel. Setiap TF memiliki **Weight** (bobot kontribusi terhadap skor akhir).

**Aturan:**
- Total **weight semua TF harus = 1.0** (100%).
- TF yang lebih besar biasanya berisi pembacaan makro/arah.
- TF yang lebih kecil memberikan presisi *entry*.

**Contoh:**
```
Timeframes:
  1m  → 0.20 (20%)   ← Presisi entry
  5m  → 0.50 (50%)   ← Primary execution
  15m → 0.30 (30%)   ← Konfirmasi arah
```

### 2.3 Indicators (Indikator Teknikal)

Ini adalah **inti dari strategi**. Setiap indikator memiliki 3 atribut kunci:

| Atribut   | Deskripsi                                                |
|:----------|:---------------------------------------------------------|
| **ID**    | Referensi ke master indikator (`m_indicator`)            |
| **Weight**| Bobot kontribusi (tergantung peran, lihat di bawah)      |
| **TF**    | Target timeframe. `nil` = berlaku di semua TF            |

#### 2.3.1 Tiga Peran Indikator (Roles)

Setiap indikator memiliki salah satu dari 3 peran yang menentukan *cara* ia berkontribusi terhadap skor:

##### 🔵 DRIVER — Penentu Arah  
Memberikan skor dasar `-100` hingga `+100`. Driver adalah satu-satunya yang **menentukan arah sinyal** (BUY/SELL).

- **Weight** = persentase kontribusi. Total semua DRIVER weight harus = **1.0** (100%).
- **Indikator**: Moving Average (MA), MACD.
- **Contoh**: `{MA, weight: 0.70}` + `{MACD, weight: 0.30}` = 1.0 ✅

##### 🟡 FILTER — Penjaga Keamanan  
Tidak menyumbang skor. Berfungsi sebagai **pengali hukuman** (*penalty multiplier*).

- Jika FILTER **setuju** dengan DRIVER → skor tetap utuh (`×1.0`).
- Jika FILTER **tidak setuju** → skor dikurangi sesuai bobot (`×0.1` = veto 90%).
- **Weight** = seberapa ketat filter ini. `0.10` = sangat ketat (veto 90%). `0.80` = longgar (veto 20%).
- **Indikator**: RSI, Stochastic, Bollinger Bands, ATR.

##### 🟣 BOOSTER — Pengali Bonus  
Tidak mengubah arah. Hanya **memperkuat skor** jika kondisi terpenuhi.

- Jika BOOSTER **aktif** → skor dikalikan `weight` (misal `×1.50`).
- Jika BOOSTER **diam** → tidak ada efek (`×1.0`).
- **Weight** = faktor pengali bonus. `1.50` = bonus 50%. `2.00` = bonus 100%.
- **Indikator**: Volume, Candle Patterns, Trend Bonus.

#### 2.3.2 Timeframe Targeting (V3)

Setiap indikator bisa ditargetkan ke TF tertentu:

| TF Value | Artinya                                 |
|:---------|:----------------------------------------|
| `nil`    | Indikator dihitung di **semua** TF      |
| `"15m"`  | Indikator **hanya** dihitung di TF 15m  |
| `"1h"`   | Indikator **hanya** dihitung di TF 1h   |

**Best Practice:**
- **DRIVER di TF besar** → membaca tren makro (e.g. MA di `4h`).
- **FILTER di TF execution** → menyaring entry yang buruk (e.g. RSI di `15m`).
- **BOOSTER di TF kecil** → mengkonfirmasi momentum (e.g. Volume di `5m`).

### 2.4 Money Management

Parameter manajemen risiko yang mengontrol seberapa agresif bot beroperasi:

| Parameter                | Tipe    | Deskripsi                                               | Rentang Umum    |
|:-------------------------|:--------|:--------------------------------------------------------|:----------------|
| `MIN_CONFIDENCE`         | Number  | Minimum skor keyakinan untuk membuka posisi              | 35–65           |
| `MAX_DAILY_TRADES`       | Number  | Maksimum jumlah trade per hari                           | 1–15            |
| `MAX_DAILY_LOSS_PERCENT` | Decimal | Batas rugi harian (% dari saldo)                         | 0.02–0.08       |
| `MAX_DAILY_LOSS_COUNT`   | Number  | Batas loss berturut-turut sebelum berhenti                | 1–5             |
| `RISK_REWARD_RATIO`      | Decimal | Minimum R:R ratio yang diterima                          | 1.5–3.0         |
| `RISK_REWARD_TARGET`     | Decimal | R:R target untuk melewati batas `MAX_DAILY_TRADES`       | 2.5–6.0         |
| `RISK_ENTRY_BUFFER`      | Decimal | Jarak buffer entry dari harga saat ini (%)                | 0.002–0.008     |
| `MAX_POSITION_SIZE`      | Decimal | Ukuran posisi maksimum (% dari saldo)                    | 0.05–0.20       |
| `LEVERAGE`               | Number  | Faktor leverage                                          | 2–10            |
| `IS_AGRESSIVE`           | Boolean | Entry 1 pakai MARKET order (langsung terisi)              | true/false       |
| `ORDER_EXPIRATION_HOURS` | Number  | Berapa jam limit order bertahan sebelum dibatalkan        | 1–48            |

**Hubungan Antar Parameter (hanya contoh ilustrasi, bukan aturan baku):**

```
Contoh Scalping  → MIN_CONFIDENCE rendah, MAX_TRADES tinggi, LEVERAGE tinggi, EXPIRY pendek
Contoh Swing     → MIN_CONFIDENCE tinggi, MAX_TRADES rendah, LEVERAGE rendah, EXPIRY panjang
```

> ⚠️ Angka dan kombinasi di atas hanyalah gambaran umum. Sesuaikan seluruh parameter berdasarkan hasil backtest dan toleransi risiko Anda.

### 2.5 Symbols

Daftar koin yang akan diperdagangkan. Setiap symbol bisa diaktifkan/nonaktifkan secara independen.

**Tips pemilihan koin:**
- **Scalping**: Pilih koin dengan volatilitas tinggi (PEPE, WIF, DOGE).
- **Day Trading**: Koin menengah-besar (SOL, ETH, LINK, AVAX).
- **Swing/Hodl**: Blue-chip stabil (BTC, ETH, BNB).

---

## 3. Daftar Indikator Tersedia

| ID | Nama             | Kode Internal      | Peran Default | Deskripsi Singkat                         |
|:---|:-----------------|:-------------------|:--------------|:------------------------------------------|
| 1  | Moving Average   | `moving_average`   | DRIVER        | SMA/EMA multi-period (20, 50, 200, 12, 26)|
| 2  | MACD             | `macd`             | DRIVER        | Momentum tren (12, 26, 9)                  |
| 3  | RSI              | `rsi`              | FILTER        | Oscillator overbought/oversold (14)        |
| 4  | Stochastic       | `stochastic`       | FILTER        | K/D momentum (14, 3, 3)                   |
| 5  | Bollinger Bands  | `bollinger_bands`  | FILTER        | Volatility bands (20, 2σ)                  |
| 6  | Volume           | `volume`           | BOOSTER       | Volume vs MA comparison (20)               |
| 7  | Candle Patterns  | `candle_patterns`  | BOOSTER       | Pola candlestick reversal/continuation     |
| 8  | ATR              | `atr`              | FILTER        | Average True Range volatilitas (14)        |
| 9  | Trend Bonus      | `trend_bonus`      | BOOSTER       | Reward saat MA+MACD selaras (±20 signal)   |

---

## 4. Formula Perhitungan Skor

Untuk setiap Timeframe, mesin menghitung skor akhir dengan 3 tahap:

```
Tahap 1: DRIVER SCORE
  driverScore = Σ (driver.signal × driver.weight)
  → Contoh: MA(+80)×0.70 + MACD(+60)×0.30 = 56 + 18 = 74

Tahap 2: FILTER MULTIPLIER
  Jika filter TIDAK SETUJU dengan arah driver:
    filterMultiplier *= 1.0 - (disagreement × (1.0 - filter.weight))
  → Contoh: RSI berlawanan (weight=0.10) → multiplier = 0.10 (veto 90%)

Tahap 3: BOOSTER MULTIPLIER
  Jika booster SETUJU dengan arah driver:
    boosterMultiplier *= booster.weight
  → Contoh: Volume mendukung (weight=1.50) → multiplier = 1.50

SKOR AKHIR TF = driverScore × filterMultiplier × boosterMultiplier
```

Skor akhir global dihitung dari rata-rata tertimbang semua TF:
```
totalScore = Σ (skor_TF × tf_weight)
```

---

## 5. Contoh Strategi: "Momentum Breakout Hunter"

Strategi ini mendeteksi **breakout momentum** pada timeframe 15m dengan konfirmasi tren dari 1h dan presisi entry dari 5m.

### 5.1 Konfigurasi

```yaml
Name:       "Momentum Breakout Hunter"
Primary TF: "15m"

# Multi-Timeframe Analysis
Timeframes:
  - 5m:  0.20   # 20% — Presisi entry (candle patterns)
  - 15m: 0.50   # 50% — Zona utama (Volume + BB squeeze)
  - 1h:  0.30   # 30% — Konfirmasi tren makro (MA + MACD)

# Indicator Weights
Indicators:
  # ─── DRIVER (di 1h: membaca tren besar) ───
  - MA   → weight: 0.70, TF: "1h"    # MA dominan di 1h
  - MACD → weight: 0.30, TF: "1h"    # MACD momentum di 1h

  # ─── FILTER (di 15m: cegah masuk di zona berbahaya) ───
  - RSI  → weight: 0.20, TF: "15m"   # Veto 80% jika overbought
  - BB   → weight: 0.30, TF: "15m"   # Veto 70% jika di luar band

  # ─── BOOSTER (di 5m & 15m: konfirmasi momentum) ───
  - Volume         → weight: 1.50, TF: "15m"  # 50% bonus jika volume meledak
  - Candle Pattern → weight: 1.30, TF: "5m"   # 30% bonus jika pola reversal

# Money Management
Money Management:
  MIN_CONFIDENCE:         45
  MAX_DAILY_TRADES:       6
  MAX_DAILY_LOSS_PERCENT: 0.04    # Max 4% rugi per hari
  MAX_DAILY_LOSS_COUNT:   3       # Berhenti setelah 3x loss berturut
  RISK_REWARD_RATIO:      2.0     # Minimum R:R 2:1
  RISK_REWARD_TARGET:     3.5     # R:R target untuk bonus trade
  RISK_ENTRY_BUFFER:      0.003   # Buffer entry 0.3%
  MAX_POSITION_SIZE:      0.10    # Max 10% saldo per posisi
  LEVERAGE:               5
  IS_AGRESSIVE:           false   # LIMIT order, sabar menunggu
  ORDER_EXPIRATION_HOURS: 3       # Limit order kadaluarsa 3 jam

# Symbols
Symbols:
  - ETHUSDT
  - SOLUSDT
  - AVAXUSDT
  - LINKUSDT
```

### 5.2 Representasi Seeder (Go Code)

```go
{
    Name:       "Momentum Breakout Hunter",
    PrimaryTF:  "15m",
    Timeframes: []tfSeed{{"5m", 0.20}, {"15m", 0.50}, {"1h", 0.30}},
    Indicators: []indSeed{
        // DRIVER di 1h (tren makro)
        {1, 0.70, tf("1h")},   // MA Driver
        {2, 0.30, tf("1h")},   // MACD Driver

        // FILTER di 15m (penjaga entry)
        {3, 0.20, tf("15m")},  // RSI Filter (veto 80%)
        {5, 0.30, tf("15m")},  // BB Filter (veto 70%)

        // BOOSTER di 5m & 15m (konfirmasi momentum)
        {6, 1.50, tf("15m")},  // Volume Booster (+50%)
        {7, 1.30, tf("5m")},   // Candle Booster (+30%)
    },
    MoneyMgmt: []mmSeed{
        {"MIN_CONFIDENCE", "45"}, {"MAX_DAILY_TRADES", "6"},
        {"MAX_DAILY_LOSS_PERCENT", "0.04"}, {"MAX_DAILY_LOSS_COUNT", "3"},
        {"RISK_REWARD_RATIO", "2.0"}, {"RISK_REWARD_TARGET", "3.5"},
        {"RISK_ENTRY_BUFFER", "0.003"}, {"MAX_POSITION_SIZE", "0.10"},
        {"LEVERAGE", "5"}, {"IS_AGRESSIVE", "false"},
        {"ORDER_EXPIRATION_HOURS", "3"},
    },
    Symbols: []symSeed{
        {"ETHUSDT"}, {"SOLUSDT"}, {"AVAXUSDT"}, {"LINKUSDT"},
    },
},
```

### 5.3 Simulasi Skor

Misalkan pada candle tertentu, kondisi pasar di setiap TF:

```
─── TF 1h (Driver Zone) ───
  MA  signal: +80  × weight 0.70 = +56.0
  MACD signal: +60 × weight 0.30 = +18.0
  driverScore = 74.0 (orientasi BULLISH ↑)

─── TF 15m (Filter + Booster Zone) ───
  Driver: inherited score = 74.0
  RSI  : Overbought (SELL signal, disagreement!)
         → penalty = 1.0 - (1.0 × (1.0 - 0.20)) = 0.20
         → filterMultiplier = 0.20
  BB   : Neutral (no disagreement)
         → filterMultiplier tetap = 0.20
  Volume: Konfirmasi BULL
         → boosterMultiplier = 1.50

  Skor TF 15m = 74.0 × 0.20 × 1.50 = 22.2

─── TF 5m (Booster Zone) ───
  Driver: sama = 74.0
  Candle: Bullish Engulfing detected
         → boosterMultiplier = 1.30

  Skor TF 5m = 74.0 × 1.0 × 1.30 = 96.2

─── FINAL ───
  totalScore = (96.2 × 0.20) + (22.2 × 0.50) + (74.0 × 0.30)
             = 19.24 + 11.10 + 22.20
             = 52.54

  Confidence = 52.54% → melebihi MIN_CONFIDENCE (45%) → ✅ VALID SIGNAL
```

> **Catatan**: Dalam contoh di atas, RSI yang overbought secara agresif menekan skor di TF 15m (dari 74 → 22). Ini adalah mekanisme perlindungan yang diinginkan: bot tetap membuka posisi karena TF 5m dan 1h masih kuat, tapi dengan keyakinan yang lebih rendah (52% vs 74%).

---

## 6. Tips Meracik Strategi

### ✅ DO (Lakukan)
1. **Selalu punya minimal 1 DRIVER** — tanpa driver, skor akan selalu 0.
2. **Tempatkan DRIVER di TF lebih besar** dari primary TF — membaca tren makro.
3. **Gunakan FILTER hemat** — terlalu banyak filter membuat bot terlalu takut masuk.
4. **Backtest sebelum live** — gunakan fitur Backtest untuk menguji strategi pada data historis.
5. **Mulai dengan leverage rendah** — naikkan secara bertahap setelah terbukti profit.

### ❌ DON'T (Hindari)
1. **Jangan taruh semua indikator di semua TF** (`nil`) — terlalu banyak konflik sinyal.
2. **Jangan set Filter weight terlalu rendah** (< 0.10) — veto 90%+ membuat filter terlalu garang.
3. **Jangan gunakan Booster tanpa Filter** — potensi over-leveraging di zona berbahaya.
4. **Jangan set `MAX_POSITION_SIZE` terlalu besar** di leverage tinggi — risiko likuidasi.

---

## 7. Proses Operasional

```
1. Racik Strategi    → Tentukan TF, Indikator, MM, Symbols
2. Backtest          → Jalankan simulasi pada data historis (e.g. 30–90 hari)
3. Evaluasi Metrik   → Cek Win Rate, Profit Factor, Max Drawdown
4. Optimasi          → Sesuaikan bobot dan parameter berdasarkan hasil backtest
5. Aktivasi Bot      → Set strategy sebagai Active, mulai live trading
6. Monitor           → Pantau trade aktif, pastikan bot berjalan sesuai harapan
```

---

*Dokumen ini dibuat: 2026-03-23. Versi: V3 (dengan per-TF indicator targeting)*
