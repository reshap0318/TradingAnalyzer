# Binance Futures Client

Client untuk integrasi dengan **Binance Futures API** menggunakan library `github.com/adshao/go-binance/v2/futures`.

## 📋 Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Initialization](#initialization)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Error Handling](#error-handling)
- [Rate Limits](#rate-limits)

---

## ✨ Features

### Market Data
- ✅ Get current price
- ✅ Get kline/candlestick data
- ✅ Get 24hr ticker statistics
- ✅ Get all prices

### Account Management
- ✅ Get account information
- ✅ Get balance by asset
- ✅ Get position by symbol

### Order Management
- ✅ Place market/limit/stop orders
- ✅ Cancel single order
- ✅ Cancel all open orders
- ✅ Get order details
- ✅ Get open orders

### Position Management
- ✅ Set leverage (1-125x)
- ✅ Set margin mode (Isolated/Crossed)
- ✅ Set position side (Hedge Mode)
- ✅ Get all positions
- ✅ Close single position
- ✅ Close all positions

### Health Check
- ✅ API connectivity test
- ✅ Server time
- ✅ Exchange information

---

## 📦 Installation

Library sudah terinstall di `go.mod`:

```bash
go get github.com/adshao/go-binance/v2
```

---

## ⚙️ Configuration

Tambahkan konfigurasi di `.env`:

```env
# Binance Futures Configuration
TESTNET_API_KEY=your_testnet_api_key
TESTNET_SECRET_KEY=your_testnet_secret_key
BINANCE_API_KEY=your_mainnet_api_key
BINANCE_SECRET_KEY=your_mainnet_secret_key
```

Update konfigurasi di `internal/config/Config.go`:

```go
type BinanceConfig struct {
    IsTestnet  bool
    APIKey     string
    SecretKey  string
    Timeout    int        // default: 30
    MaxRetries int        // default: 3
    RetryDelay int        // default: 1 (seconds)
}
```

---

## 🚀 Initialization

### Initialize di Dependency Injection Container

```go
// internal/di/container.go
import "github.com/reshap/trading-bot/internal/clients/binance"

func NewContainer(cfg *config.Config) (*Container, error) {
    // ... existing code

    // Initialize Binance Futures client
    binanceClient := binance.NewClient(&cfg.BINANCE)

    return &Container{
        // ... existing fields
        BinanceClient: binanceClient,
    }, nil
}
```

---

## 💡 Usage Examples

### 1. Get Current Price

```go
package service

import (
    "github.com/reshap/trading-bot/internal/clients/binance"
)

func (s *Services) GetCryptoPrice(symbol string) (float64, error) {
    priceInfo, err := s.BinanceClient.GetPrice(symbol)
    if err != nil {
        return 0, err
    }
    
    return priceInfo.Price, nil
}

// Usage
price, err := s.GetCryptoPrice("BTCUSDT")
```

### 2. Place Market Order

```go
func (s *Services) PlaceMarketOrder(symbol string, side binance.OrderSide, quantity float64) (*binance.OrderResponse, error) {
    req := &binance.PlaceOrderRequest{
        Symbol:   symbol,
        Side:     side,
        Type:     binance.OrderTypeMarket,
        Quantity: quantity,
    }
    
    return s.BinanceClient.PlaceOrder(req)
}

// Usage - Buy 0.001 BTC
order, err := s.PlaceMarketOrder("BTCUSDT", binance.OrderSideBuy, 0.001)
```

### 3. Place Limit Order

```go
func (s *Services) PlaceLimitOrder(symbol string, side binance.OrderSide, quantity, price float64) (*binance.OrderResponse, error) {
    req := &binance.PlaceOrderRequest{
        Symbol:      symbol,
        Side:        side,
        Type:        binance.OrderTypeLimit,
        Quantity:    quantity,
        Price:       price,
        TimeInForce: binance.TimeInForceGTC,
    }

    return s.BinanceClient.PlaceOrder(req)
}

// Usage - Buy 0.001 BTC at $50,000
order, err := s.PlaceLimitOrder("BTCUSDT", binance.OrderSideBuy, 0.001, 50000)
```

### 3a. Place Order with Auto Precision Adjustment

Function ini otomatis menyesuaikan quantity dan price dengan presisi exchange:

```go
func (s *Services) PlaceOrderWithPrecision(symbol string, side binance.OrderSide, quantity, price float64) (*binance.OrderResponse, error) {
    req := &binance.PlaceOrderRequest{
        Symbol:      symbol,
        Side:        side,
        Type:        binance.OrderTypeLimit,
        Quantity:    quantity,  // Akan disesuaikan otomatis dengan step size
        Price:       price,     // Akan disesuaikan otomatis dengan tick size
        TimeInForce: binance.TimeInForceGTC,
    }

    // PlaceOrder otomatis adjust precision menggunakan:
    // - AdjustQuantityPrecision(qty, stepSize)
    // - AdjustPricePrecision(price, tickSize)
    return s.BinanceClient.PlaceOrder(req)
}

// Usage - Quantity dan price akan auto-adjust
// BTCUSDT: stepSize = 0.001, tickSize = 0.1
// Input: qty = 0.12345, price = 50000.123
// Auto-adjusted: qty = 0.123, price = 50000.1
order, err := s.PlaceOrderWithPrecision("BTCUSDT", binance.OrderSideBuy, 0.12345, 50000.123)
```

### 4. Set Leverage

```go
func (s *Services) SetSymbolLeverage(symbol string, leverage int) error {
    req := &binance.LeverageRequest{
        Symbol:   symbol,
        Leverage: leverage,
    }
    
    _, err := s.BinanceClient.SetLeverage(req)
    return err
}

// Usage - Set 10x leverage for ETHUSDT
err := s.SetSymbolLeverage("ETHUSDT", 10)
```

### 5. Get Account Balance

```go
func (s *Services) GetUSDTBalance() (float64, error) {
    balance, err := s.BinanceClient.GetBalance("USDT")
    if err != nil {
        return 0, err
    }
    
    return balance.AvailableBalance, nil
}
```

### 6. Get All Positions

```go
func (s *Services) GetAllPositions() ([]binance.PositionInfo, error) {
    return s.BinanceClient.GetPositions()
}

// Usage
positions, err := s.GetAllPositions()
for _, pos := range positions {
    fmt.Printf("Symbol: %s, Amount: %f, PnL: %f\n", 
        pos.Symbol, pos.PositionAmt, pos.UnrealizedProfit)
}
```

### 7. Close All Positions

```go
func (s *Services) EmergencyCloseAllPositions() error {
    return s.BinanceClient.CloseAllPositions()
}
```

### 8. Cancel All Orders

```go
func (s *Services) CancelAllOrdersForSymbol(symbol string) error {
    return s.BinanceClient.CancelAllOrders(symbol)
}
```

### 9. Get Multi-Timeframe Klines (Parallel)

```go
func (s *Services) GetMultiTimeframeData(symbol string) (map[string][][]float64, error) {
    // Define timeframes yang diinginkan
    requests := []binance.MultiKlineRequest{
        {Interval: "5m", Limit: 100},    // 5 minutes, 100 candles
        {Interval: "15m", Limit: 100},   // 15 minutes, 100 candles
        {Interval: "1h", Limit: 100},    // 1 hour, 100 candles
        {Interval: "4h", Limit: 100},    // 4 hours, 100 candles
        {Interval: "1d", Limit: 100},    // 1 day, 100 candles
    }
    
    // Get OHLCV data (parallel fetch)
    ohlcvMap, err := s.BinanceClient.GetMultiKlinesOHLCV(symbol, requests)
    if err != nil {
        return nil, err
    }
    
    // Access data by timeframe
    klines1h := ohlcvMap["1h"]  // [][]float64 [open, high, low, close, volume]
    klines4h := ohlcvMap["4h"]
    klines1d := ohlcvMap["1d"]
    
    return ohlcvMap, nil
}

// Usage
ohlcvData, err := s.GetMultiTimeframeData("BTCUSDT")
// ohlcvData["1h"][0] = [open, high, low, close, volume]
```

### 10. Get Multi-Timeframe Klines (Raw Format)

```go
func (s *Services) GetMultiTimeframeRawData(symbol string) (map[string][]binance.KlineInfo, error) {
    requests := []binance.MultiKlineRequest{
        {Interval: "5m", Limit: 50},
        {Interval: "1h", Limit: 50},
        {Interval: "1d", Limit: 50},
    }
    
    // Get raw kline data (parallel fetch)
    klinesMap, err := s.BinanceClient.GetMultiKlines(symbol, requests)
    if err != nil {
        return nil, err
    }
    
    return klinesMap, nil
}

// Usage
klinesMap, err := s.GetMultiTimeframeRawData("BTCUSDT")
for interval, klines := range klinesMap {
    fmt.Printf("Timeframe: %s, Candles: %d\n", interval, len(klines))
    for _, k := range klines {
        fmt.Printf("Open: %f, High: %f, Low: %f, Close: %f, Volume: %f\n",
            k.Open, k.High, k.Low, k.Close, k.Volume)
    }
}
```

---

## 📚 API Reference

### Market Data Methods

| Method | Description | Parameters | Returns | Cache |
|--------|-------------|------------|---------|-------|
| `GetPrice()` | Get current mark price | `symbol` | `*PriceInfo, error` | ✅ 5s |
| `GetKlines()` | Get candlestick data | `symbol, interval, limit` | `[]KlineInfo, error` | ✅ 30s |
| `GetAllPrices()` | Get all prices | - | `[]PriceInfo, error` | ❌ |
| `GetMultiKlines()` | Get multi-timeframe klines (parallel) | `symbol, []MultiKlineRequest` | `map[string][]KlineInfo, error` | ❌ |
| `GetMultiKlinesOHLCV()` | Get multi-timeframe OHLCV (parallel) | `symbol, []MultiKlineRequest` | `map[string][][]float64, error` | ❌ |

**Note:** Cache menggunakan Redis untuk mengurangi API rate limit dan meningkatkan response time.

### Helper Functions (Precision Adjustment)

| Function | Description | Parameters | Returns |
|----------|-------------|------------|---------|
| `AdjustQuantityPrecision()` | Adjust quantity to match step size | `quantity, stepSize` | `float64` |
| `AdjustPricePrecision()` | Adjust price to match tick size | `price, tickSize` | `float64` |

**Auto-Adjustment in PlaceOrder:**

Function `PlaceOrder()` sekarang **otomatis** menyesuaikan quantity dan price:

```go
// Example: BTCUSDT (stepSize = 0.001, tickSize = 0.1)
req := &binance.PlaceOrderRequest{
    Symbol:   "BTCUSDT",
    Side:     binance.OrderSideBuy,
    Type:     binance.OrderTypeLimit,
    Quantity: 0.12345,  // ← Auto-adjusted to 0.123
    Price:    50000.123, // ← Auto-adjusted to 50000.1
}

// Tidak perlu manual adjust!
order, err := client.PlaceOrder(req)
```

**Manual Adjustment (jika diperlukan):**

```go
// Get symbol info
symbolInfo, err := client.GetSymbolInfo("BTCUSDT")
// symbolInfo.StepSize = 0.001, symbolInfo.TickSize = 0.1

// Manual adjustment
adjustedQty := binance.AdjustQuantityPrecision(0.12345, 0.001)  // Returns: 0.123
adjustedPrice := binance.AdjustPricePrecision(50000.123, 0.1)   // Returns: 50000.1
```

### Account Methods

| Method | Description | Parameters | Returns | Cache |
|--------|-------------|------------|---------|-------|
| `GetAccountInfo()` | Get full account info | - | `*AccountInfo, error` | ❌ |
| `GetBalance()` | Get balance by asset | `asset` | `*BalanceInfo, error` | ❌ |
| `GetPosition()` | Get position by symbol | `symbol` | `*PositionInfo, error` | ✅ 10s |
| `GetPositions()` | Get all open positions | - | `[]PositionInfo, error` | ❌ |

### Order Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `PlaceOrder()` | Place new order | `*PlaceOrderRequest` | `*OrderResponse, error` |
| `CancelOrder()` | Cancel specific order | `*CancelOrderRequest` | `*CancelOrderResponse, error` |
| `CancelAllOrders()` | Cancel all orders | `symbol` | `error` |
| `GetOrder()` | Get order detail | `*GetOrdersRequest` | `*OrderResponse, error` |
| `GetOpenOrders()` | Get all open orders | `symbol` | `[]OrderResponse, error` |

### Position Management Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `SetLeverage()` | Set leverage | `*LeverageRequest` | `*LeverageResponse, error` |
| `SetMarginMode()` | Set margin mode | `*MarginModeRequest` | `*MarginModeResponse, error` |
| `SetPositionSide()` | Set hedge mode | `*PositionSideRequest` | `*PositionSideResponse, error` |
| `ClosePosition()` | Close position | `symbol, qty, side` | `*OrderResponse, error` |
| `CloseAllPositions()` | Close all positions | - | `error` |

### Health Check Methods

| Method | Description | Parameters | Returns |
|--------|-------------|------------|---------|
| `Ping()` | Test API connectivity | - | `error` |
| `GetServerTime()` | Get server time | - | `time.Time, error` |
| `GetExchangeInfo()` | Get exchange info | - | `*futures.ExchangeInfo, error` |

---

## ⚠️ Error Handling

### Common Errors

```go
var (
    ErrAPIKeyRequired      = errors.New("API key is required")
    ErrSecretKeyRequired   = errors.New("secret key is required")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrInvalidSymbol       = errors.New("invalid symbol")
    ErrOrderFailed         = errors.New("order failed")
    ErrAPIRateLimit        = errors.New("API rate limit exceeded")
    ErrInvalidLeverage     = errors.New("invalid leverage")
    ErrPositionNotFound    = errors.New("position not found")
)
```

### Example Error Handling

```go
func (s *Services) SafePlaceOrder(req *binance.PlaceOrderRequest) (*binance.OrderResponse, error) {
    ctx := context.Background()
    
    order, err := s.BinanceClient.PlaceOrder(ctx, req)
    if err != nil {
        switch {
        case errors.Is(err, binance.ErrInsufficientBalance):
            // Handle insufficient balance
            return nil, fmt.Errorf("saldo tidak mencukupi")
        case errors.Is(err, binance.ErrInvalidSymbol):
            // Handle invalid symbol
            return nil, fmt.Errorf("symbol tidak valid")
        case errors.Is(err, binance.ErrAPIRateLimit):
            // Handle rate limit
            return nil, fmt.Errorf("terlalu banyak request")
        default:
            return nil, err
        }
    }
    
    return order, nil
}
```

---

## 🚦 Rate Limits

Binance Futures API memiliki rate limits:

| Weight | Limit | Description |
|--------|-------|-------------|
| 1-10 | 1200/min | Most endpoints |
| 20 | 1200/min | Place order |
| 40 | 1200/min | Cancel order |

**Best Practices:**
- ✅ Gunakan `context.WithTimeout()` untuk semua API calls
- ✅ Implement retry logic dengan exponential backoff
- ✅ Cache data yang tidak sering berubah (exchange info, dll)
- ✅ Monitor API weight usage

---

## 📝 DTOs Reference

### Order Types

```go
const (
    OrderTypeLimit        OrderType = "LIMIT"
    OrderTypeMarket       OrderType = "MARKET"
    OrderTypeStopMarket   OrderType = "STOP_MARKET"
    OrderTypeStopLimit    OrderType = "STOP_LOSS_LIMIT"
    OrderTypeTakeProfit   OrderType = "TAKE_PROFIT"
    OrderTypeTrailingStop OrderType = "TRAILING_STOP_MARKET"
)
```

### Order Side

```go
const (
    OrderSideBuy  OrderSide = "BUY"
    OrderSideSell OrderSide = "SELL"
)
```

### Time In Force

```go
const (
    TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
    TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
    TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
    TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing (post only)
)
```

### Position Side

```go
const (
    PositionSideBoth  PositionSide = "BOTH"  // One-way mode
    PositionSideLong  PositionSide = "LONG"  // Hedge mode
    PositionSideShort PositionSide = "SHORT" // Hedge mode
)
```

---

## 🔒 Security Best Practices

1. **Jangan commit API keys** ke Git
2. **Gunakan Testnet** untuk development
3. **Enable IP whitelist** di Binance dashboard
4. **Gunakan API key dengan permission minimal** yang diperlukan
5. **Rotate API keys** secara berkala
6. **Monitor API usage** untuk deteksi anomaly

---

## 🧪 Testing

### Test dengan Testnet

```go
// Set testnet di .env
TESTNET_API_KEY=your_testnet_key
TESTNET_SECRET_KEY=your_testnet_secret

// Testnet URL: https://testnet.binancefuture.com
```

### Example Test

```go
func TestBinanceClient_GetPrice(t *testing.T) {
    cfg := &config.BinanceConfig{
        IsTestnet:  true,
        APIKey:     os.Getenv("TESTNET_API_KEY"),
        SecretKey:  os.Getenv("TESTNET_SECRET_KEY"),
        Timeout:    30,
        MaxRetries: 3,
    }
    
    client := binance.NewClient(cfg)
    ctx := context.Background()
    
    price, err := client.GetPrice(ctx, "BTCUSDT")
    
    assert.NoError(t, err)
    assert.Greater(t, price.Price, float64(0))
}
```

---

## 💾 Redis Caching

Client ini menggunakan **Redis cache** untuk mengurangi API rate limit dan meningkatkan response time.

### **Cached Methods**

| Method | TTL | Cache Key Pattern |
|--------|-----|-------------------|
| `GetPrice()` | 5 seconds | `binance:futures:price:{symbol}` |
| `GetKlines()` | 30 seconds | `binance:futures:klines:{symbol}:{interval}:{limit}` |
| `GetPosition()` | 10 seconds | `binance:futures:position:{symbol}` |
| `SetLeverage()` | 72 hours | `binance:futures:leverage:{symbol}` |
| `GetExchangeInfo()` | 7 days | `binance:futures:exchange_info:all` |

### **Cache Benefits**

- ⚡ **Faster Response**: Redis ~1ms vs API ~100-500ms
- 📉 **Reduce API Calls**: Cache hit rate ~60-80% untuk read operations
- 🛡️ **Rate Limit Protection**: Mengurangi API weight consumption

### **Cache Configuration**

```go
// Default cache configuration
cacheCfg := &binance.CacheConfig{
    Enabled:    true,
    DefaultTTL: 10 * time.Second,
    PriceTTL:   5 * time.Second,
    KlineTTL:   30 * time.Second,
    AccountTTL: 10 * time.Second,
}

// Set custom cache config
client.SetCacheConfig(cacheCfg)
```

### **Cache Invalidation**

Cache otomatis di-invalidate setelah:
- Order placement (`PlaceOrder`)
- Order cancellation (`CancelOrder`)
- Position change (manual close)

**Manual Cache Clear:**
```go
// Clear specific cache
cacheKey := "binance:futures:price:BTCUSDT"
client.cache.Delete(ctx, cacheKey)

// Clear all price caches
client.cache.DeleteByPattern(ctx, "binance:futures:price:*")
```

---

## 📖 References

- [Binance Futures API Documentation](https://binance-docs.github.io/apidocs/futures/en/)
- [go-binance GitHub](https://github.com/adshao/go-binance)
- [Binance Testnet](https://testnet.binancefuture.com/)

---

*Last Updated: March 8, 2026 - Added Redis caching for GetPrice, GetKlines, GetPosition*
