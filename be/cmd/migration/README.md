# Migration

Database migration tool using GORM's AutoMigrate.

## Setup

### Prerequisites

- Go 1.25+
- MySQL database
- `.env` file with database configuration

### Environment Variables

Make sure your `.env` file contains:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=trading_bot
```

## Usage

### Run Migrations (Up)

Create or update tables based on models:

```bash
go run cmd/migration/main.go up
```

This will:
- Create tables that don't exist
- Update existing table schema if models have changed
- Add missing columns/indexes

### Seed Default Data

Insert default data into tables:

```bash
go run cmd/migration/main.go seed
```

This will:
- Insert default threshold configurations
- Insert default strategies with weights
- Skip existing records (idempotent - safe to run multiple times)

### Rollback Migrations (Down)

Drop all migrated tables:

```bash
go run cmd/migration/main.go down
```

⚠️ **Warning:** This will delete all data in the migrated tables!

## Adding New Models

To add a new model to migrations:

1. Create the model in `internal/models/`
2. Add the model to the `AutoMigrate` call in `cmd/migration/main.go`:

```go
err := db.AutoMigrate(
    &models.Threshold{},
    &models.YourNewModel{},  // Add your new model here
)
```

## Strategy Seeder - Signal Generation

The strategy seeder inserts indicator strategies with weights. Here's how signals are generated:

### How Signal is Calculated

#### 1. Each Indicator Returns Raw Signal (-100 to 100)

Example:
- RSI oversold → +30 (bullish)
- MACD bearish cross → -40 (bearish)
- MA bullish trend → +85 (strong bullish)
- Trend Bonus (MA+MACD aligned) → ±20 (trend confirmation)

#### 2. Raw Signal is Multiplied by Weight

```
contribution = rawSignal × weight
```

Example:
- RSI: +30 × 0.15 = +4.5
- MACD: -40 × 0.18 = -7.2
- MA: +85 × 0.22 = +18.7

#### 3. All Contributions are Summed

```
finalSignal = Σ(contributions)
```

**Example Calculation:**

| Indicator       | Raw Signal | Weight | Contribution |
|-----------------|------------|--------|--------------|
| MA              | +85        | 0.22   | +18.7        |
| MACD            | -40        | 0.18   | -7.2         |
| RSI             | +30        | 0.15   | +4.5         |
| Stochastic      | -20        | 0.10   | -2.0         |
| Bollinger Bands | +15        | 0.10   | +1.5         |
| Volume          | 0          | 0.05   | 0            |
| Candle Patterns | +12        | 0.08   | +0.96        |
| ATR             | 0          | 0.04   | 0            |
| Trend Bonus     | +20        | 0.08   | +1.6         |
| **TOTAL**       |            | **1.00** | **+18.06**  |

#### 4. Final Signal is Matched Against Threshold

| Category     | Range       | Action | Color       |
|--------------|-------------|--------|-------------|
| STRONG_BUY   | 70 to 100   | BUY    | green       |
| BUY          | 45 to 70    | BUY    | light-green |
| WAIT         | -45 to 45   | WAIT   | gray        |
| SELL         | -70 to -45  | SELL   | light-red   |
| STRONG_SELL  | -100 to -70 | SELL   | red         |

**Example:** +18.06 → **WAIT** (within -45 to 45 range)

### Special Notes

#### Trend Bonus
- Returns fixed ±20 signal when MA and MACD are aligned
- Weight: 0.08 (8%)
- Contribution: ±20 × 0.08 = ±1.6 to final signal
- Purpose: Trend confirmation, not primary signal
- Why ±20? Prevents trend from dominating the signal

#### Candle Patterns
- Now analyzes last 5 candles (with history)
- Each pattern has specific score:
  - Bullish Engulfing: +12
  - Morning Star: +15
  - Hammer: +8
  - Bearish Engulfing: -12
  - Evening Star: -15
  - etc.
- Weight: 0.08 (8%)
- Patterns accumulate (multiple patterns can be detected)

#### Stochastic Neutralization
- In strong uptrend: Overbought penalty is ignored (signal = 0)
- In strong downtrend: Oversold bonus is ignored (signal = 0)
- Prevents false reversal signals in trending markets

### Weight Distribution

**Total Weight: 1.00 (100%)**

- **Core Indicators (55%)**: MA (22%), MACD (18%), RSI (15%)
- **Secondary Indicators (25%)**: Stochastic (10%), BB (10%), Volume (5%)
- **Special Indicators (20%)**: Candle Patterns (8%), ATR (4%), Trend Bonus (8%)

---

## How It Works

This migration tool uses GORM's `AutoMigrate` feature which:

- Automatically creates tables for your models
- Adds new columns as you add fields to models
- Creates indexes and unique constraints
- Does **not** delete columns (to prevent data loss)

For more complex migrations (renaming columns, custom SQL), consider using raw SQL migrations.

## Troubleshooting

### Connection Issues

Ensure:
- MySQL server is running
- Database exists (create with `CREATE DATABASE trading_bot;`)
- Credentials in `.env` are correct

### Permission Denied

Make sure the database user has privileges to create/alter tables:

```sql
GRANT ALL PRIVILEGES ON trading_bot.* TO 'root'@'localhost';
FLUSH PRIVILEGES;
```
