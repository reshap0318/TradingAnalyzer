# Migration Seeders

Folder ini berisi function-function untuk seed data default ke database.

## 📁 Structure

```
seeders/
├── threshold_seeder.go    # Seed signal thresholds (STRONG_BUY, BUY, WAIT, SELL, STRONG_SELL)
├── timeframe_seeder.go    # Seed timeframes (1m, 5m, 15m, 1h, 4h, 1d, etc.)
├── config_seeder.go       # Seed system configurations (MM, Binance settings)
├── indicator_seeder.go    # Seed technical indicators (MA, MACD, RSI, etc.)
├── strategy_seeder.go     # Seed strategy with timeframes, indicators, and money management
└── watchlist_seeder.go    # Seed watchlist symbols (BTCUSDT, ETHUSDT, etc.)
```

## 🔧 Usage

### Seed All Data

```bash
cd be
go run cmd/migration/main.go seed
```

### Refresh (Drop + Migrate + Seed)

```bash
go run cmd/migration/main.go refresh
```

### Individual Seeder (for development)

```go
import "github.com/reshap/trading-bot/cmd/migration/seeders"

// In your code
seeders.SeedThreshold(db)
seeders.SeedTimeframes(db)
// ... etc
```

## 📊 Seeded Data

### 1. Thresholds (5 records)

Signal score thresholds untuk categorization:

| Category | Min | Max | Action | Color |
|----------|-----|-----|--------|-------|
| STRONG_BUY | 70 | 100 | BUY | green |
| BUY | 25 | 70 | BUY | light-green |
| WAIT | -25 | 25 | WAIT | gray |
| SELL | -70 | -25 | SELL | red |
| STRONG_SELL | -100 | -70 | SELL | dark-red |

### 2. Timeframes (16 records)

All standard trading timeframes:

- **Scalping:** 1m, 3m, 5m
- **Short-term:** 15m, 30m
- **Day Trading:** 1h, 2h, 4h, 6h, 8h, 12h
- **Swing Trading:** 1d, 3d
- **Position Trading:** 1w, 1M

### 3. Configs (13 records)

**Money Management:**
- `MIN_CONFIDENCE` = 45 (%)
- `MAX_DAILY_TRADES` = 10
- `MAX_DAILY_LOSS_PERCENT` = 0.05 (5%)
- `MAX_DAILY_LOSS_COUNT` = 7
- `MAX_POSITION_SIZE` = 0.15 (15%)
- `RISK_ENTRY_BUFFER` = 0.0075 (0.75%)
- `MAX_RISK_PER_TRADE` = 0.04 (4%)
- `RISK_REWARD_RATIO` = 1.5
- `RISK_REWARD_TARGET` = 3.0
- `LEVERAGE` = 5x
- `IS_AGRESSIVE` = false
- `ORDER_EXPIRATION_HOURS` = 4

**Binance:**
- `BINANCE_TESTNET` = true

### 4. Indicators (9 records)

**Core Indicators (65% weight):**
- Moving Average (30%)
- MACD (22%)
- RSI (13%)

**Secondary Indicators (25%):**
- Stochastic (10%)
- Bollinger Bands (10%)
- Volume (5%)

**Special Indicators (10%):**
- Candle Patterns (4%)
- ATR (2%)
- Trend Bonus (4%)

### 5. Strategy (1 record + relations)

**Strategy: "Scalping Conservative"**

**Timeframes (4):**
- 5m, 15m (primary)
- 1h, 4h (trend confirmation)

**Indicators (9):**
- All indicators are linked to strategy

**Money Management:**
- Inherits from config defaults
- Can be customized per strategy

### 6. Watchlist (15 records)

Major crypto assets:

| Symbol | Name |
|--------|------|
| BTCUSDT | Bitcoin |
| ETHUSDT | Ethereum |
| BNBUSDT | Binance Coin |
| SOLUSDT | Solana |
| XRPUSDT | Ripple |
| ADAUSDT | Cardano |
| DOGEUSDT | Dogecoin |
| TRXUSDT | Tron |
| LINKUSDT | Chainlink |
| AVAXUSDT | Avalanche |
| SUIUSDT | Sui |
| LTCUSDT | Litecoin |
| NEARUSDT | Near Protocol |
| UNIUSDT | Uniswap |
| FETUSDT | Fetch.ai |

## 🔄 Idempotency

Semua seeder bersifat **idempotent**, artinya:
- ✅ Aman dijalankan berkali-kali
- ✅ Tidak akan create duplicate data
- ✅ Akan update data yang sudah ada jika perlu

**Logic:**
```go
// Check if exists
result := db.Where(...).First(&existing)

if result.Error == gorm.ErrRecordNotFound {
    // Create new
    db.Create(&model)
} else {
    // Update if needed (for some seeders)
    db.Save(&existing)
}
```

## 🛠️ Development Tips

### Add New Seeder

1. Create new file: `seeders/<name>_seeder.go`
2. Export function: `func Seed<Name>(db *gorm.DB)`
3. Call in `main.go`:

```go
func runSeed(dsn string) {
    // ...
    seeders.SeedNewSeeder(db)
    // ...
}
```

### Modify Existing Seeder

1. Edit the seeder file
2. Run `refresh` to recreate:
   ```bash
   go run cmd/migration/main.go refresh
   ```

### Debug Seeder

Enable SQL logging:

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Change from Silent
})
```

## 📝 Best Practices

1. **Always use Where().First()** to check existence
2. **Log errors but don't fail** - continue seeding other data
3. **Print progress** - use `fmt.Printf` for user feedback
4. **Respect foreign keys** - seed in correct order:
   - Master data first (Threshold, Timeframe, Config, Indicators)
   - Then Strategy (depends on Master)
   - Finally Watchlist (independent)
5. **Keep data minimal** - only seed defaults, not production data

## ⚠️ Important Notes

- **DO NOT seed production data** - this is for development/testing only
- **Config values** can be overridden via UI/database after initial seed
- **Strategy** is created once, then customized by user
- **Indicators** will be updated on re-seed (to get latest params/weights)

---

**Last Updated:** 2026-03-21
**Author:** TradingAnalyzer Team
