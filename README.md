# 📊 Trading Analyzer API

Analisis teknikal multi-timeframe untuk **Saham Indonesia** dan **Crypto** dengan sinyal BUY / SELL / WAIT otomatis.

## Fitur Utama

- **7 Indikator Teknikal** — MA, RSI, MACD, Bollinger Bands, Stochastic, Volume, ATR
- **4 Timeframe** — 15m, 1h, 4h, 1D
- **Sentimen Pasar** — IHSG (saham) atau BTC Market (crypto)
- **TP & SL Otomatis** — Berdasarkan support/resistance
- **Money Management** — Rekomendasi posisi & kalkulasi risiko
- **Capital Management** — Tracking modal tersedia (initial − allocated + realized PnL)
- **Signal Log** — Evaluasi performa sinyal otomatis (win rate, profit factor, PnL dollar)
- **Position Tracker** — Catat trade manual & hitung PnL
- **Futures Calculator** — Leverage, liquidation price, ROE (crypto only)

---

## Quick Start

```bash
npm install
npm start
```

Server berjalan di `http://localhost:3000`

---

## API Endpoints

### 📈 Analisis

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/saham/analyze?symbol=BBCA&capital=10000000` | Analisis lengkap saham (capital dalam **Rupiah**, default 10jt) |
| GET | `/saham/signal?symbol=BBCA` | Sinyal cepat (tanpa money mgmt) |
| GET | `/saham/raw?symbol=BBCA` | Data OHLCV mentah |
| GET | `/crypto/analyze?symbol=BTCUSDT&capital=50&leverage=10` | Analisis lengkap crypto + futures (capital dalam **USD**, default 50) |
| GET | `/crypto/raw?symbol=BTCUSDT` | Data OHLCV mentah |

### 📊 Signal Log (otomatis)

Setiap sinyal BUY/SELL otomatis dicatat dan dievaluasi. Maksimal 1 sinyal per simbol — sinyal baru yang searah diabaikan, sinyal berlawanan menutup yang lama.

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/saham/signals/summary?capital=10000000` | Win rate, PnL, capital status (saham) |
| GET | `/saham/signals/history` | Riwayat sinyal saham |
| GET | `/crypto/signals/summary?capital=50` | Win rate, PnL, capital status (crypto) |
| GET | `/crypto/signals/history` | Riwayat sinyal crypto |

### 📍 Position Tracker (manual)

Catat trade yang benar-benar kamu ambil — terpisah dari signal log.

| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/saham/position/open` | Buka posisi saham |
| POST | `/saham/position/close` | Tutup posisi saham |
| GET | `/saham/positions` | Posisi saham aktif |
| GET | `/saham/positions/history` | Trade saham yang sudah ditutup |
| GET | `/saham/positions/summary` | Ringkasan performa saham |
| POST | `/crypto/position/open` | Buka posisi crypto |
| POST | `/crypto/position/close` | Tutup posisi crypto |
| GET | `/crypto/positions` | Posisi crypto aktif |
| GET | `/crypto/positions/history` | Trade crypto yang sudah ditutup |
| GET | `/crypto/positions/summary` | Ringkasan performa crypto |

### Lainnya

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/health` | Cek server |

---

## Struktur Folder

```
src/
├── app.js              # Entry point & API endpoints
├── config.js           # Konfigurasi (SAHAM / CRYPTO / SHARED)
├── saham/              # Modul khusus saham
│   ├── decisionEngine.js
│   ├── timeframeManager.js
│   ├── yahooFinance.js
│   └── ihsgAnalyzer.js
├── crypto/             # Modul khusus crypto
│   ├── cryptoDecisionEngine.js
│   ├── btcMarketAnalyzer.js
│   ├── futuresCalculator.js
│   └── binanceData.js
└── shared/             # Modul yang dipakai keduanya
    ├── signalGenerator.js
    ├── tpslCalculator.js
    ├── moneyManagement.js
    ├── signalLogger.js
    ├── positionTracker.js
    └── indicators/     # Semua indikator teknikal
```

---

## Konfigurasi

Edit `src/config.js` untuk menyesuaikan:

| Section | Isi |
|---|---|
| `SAHAM` | Timeframe weights, thresholds, default capital (10jt IDR) |
| `CRYPTO` | Timeframe weights, thresholds, default capital (50 USD) |
| `CRYPTO.FUTURES` | Default leverage, max leverage, fee rate, funding rate |
| `INDICATORS` | Parameter indikator teknikal (berlaku untuk keduanya) |

---

## Dokumentasi Teknis

Detail lengkap response, field, signal log, position tracker, dan futures ada di:

👉 **[TECHNICAL.md](TECHNICAL.md)**
