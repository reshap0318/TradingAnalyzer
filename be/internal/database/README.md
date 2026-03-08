# Database Layer - TradingBot

Database layer untuk mengelola koneksi MySQL dan Redis.

## 📁 Struktur Folder

```
internal/database/
├── mysql.go           # MySQL connection helper
├── redis.go           # Redis connection helper
├── redis_cache.go     # Redis cache client wrapper
└── redis_test.go      # Unit tests untuk Redis
```

## 🚀 Quick Start

### 1. Setup Redis

Menggunakan Docker:
```bash
docker run -d -p 6379:6379 --name trading-bot-redis redis:latest
```

Atau install Redis secara manual:
- Windows: Download dari https://github.com/microsoftarchive/redis/releases
- Linux: `sudo apt-get install redis-server`
- macOS: `brew install redis`

### 2. Update Environment Variables

Tambahkan konfigurasi Redis di `.env`:

```env
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_USER=default        # Redis 6+ ACL username (optional, default: "default")
REDIS_PASSWORD=           # Redis password (optional)
REDIS_DB=0                # Database number (0-15)
```

**Notes:**
- `REDIS_USER`: Diperlukan untuk Redis 6+ dengan ACL (Access Control List)
- `REDIS_PASSWORD`: Kosongkan jika Redis tidak menggunakan password
- `REDIS_DB`: Gunakan 0 untuk default database

### 3. Initialize dalam Application

Redis sudah terintegrasi dalam DI Container:

```go
// cmd/bot/main.go
container, err := di.NewContainer(cfg)
if err != nil {
    log.Fatalf("Failed to create container: %v", err)
}

// Access Redis client
if container.Redis != nil {
    // Redis is available
    err := container.Redis.Ping()
    if err != nil {
        log.Printf("Redis ping failed: %v", err)
    } else {
        log.Println("Redis connected successfully")
    }
}
```

## 📊 Usage Examples

### Basic Operations

```go
import (
    "context"
    "time"
    "github.com/reshap/trading-bot/internal/database"
)

ctx := context.Background()

// Set string value
err := cache.Set(ctx, "key", "value", time.Minute*5)

// Get string value
value, err := cache.Get(ctx, "key")

// Delete key
err := cache.Delete(ctx, "key")

// Check if key exists
exists, err := cache.Exists(ctx, "key")
```

### JSON Operations

```go
type UserData struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Set JSON
user := UserData{Name: "John", Age: 30, Email: "john@example.com"}
err := cache.SetJSON(ctx, "user:1", user, time.Hour)

// Get JSON
var result UserData
err = cache.GetJSON(ctx, "user:1", &result)
```

### Counter Operations

```go
// Increment
count, err := cache.Incr(ctx, "visits")

// Decrement
count, err := cache.Decr(ctx, "visits")
```

### Set Operations

```go
// Add to set
err := cache.SAdd(ctx, "active_users", "user1", "user2", "user3")

// Check membership
isMember, err := cache.SIsMember(ctx, "active_users", "user1")

// Get all members
members, err := cache.SMembers(ctx, "active_users")

// Remove from set
err := cache.SRem(ctx, "active_users", "user1")
```

### List Operations

```go
// Push to list (left)
err := cache.LPush(ctx, "queue", "task1", "task2", "task3")

// Get list range
tasks, err := cache.LRange(ctx, "queue", 0, -1)

// Pop from list (right)
task, err := cache.RPop(ctx, "queue")

// Get list length
length, err := cache.LLen(ctx, "queue")
```

### Hash Operations

```go
// Set hash fields
err := cache.HSet(ctx, "user:100", "name", "John", "age", "30", "city", "NYC")

// Get hash field
name, err := cache.HGet(ctx, "user:100", "name")

// Get all hash fields
fields, err := cache.HGetAll(ctx, "user:100")

// Delete hash field
err := cache.HDel(ctx, "user:100", "age")
```

### TTL Operations

```go
// Set with TTL
err := cache.SetWithTTL(ctx, "temp_key", "temp_value", time.Second*30)

// Get remaining TTL
ttl, err := cache.GetTTL(ctx, "temp_key")

// Set expiration on existing key
err := cache.Expire(ctx, "key", time.Hour)
```

### SetNX (Set if Not Exists)

```go
// Set only if key doesn't exist
ok, err := cache.SetNX(ctx, "lock:key", "locked", time.Minute*5)
if ok {
    // Successfully acquired lock
}
```

### Pattern Matching

```go
// Get all keys matching pattern
keys, err := cache.Keys(ctx, "user:*")

// Flush database (be careful!)
err := cache.FlushDB(ctx)
```

## 🔧 Available Methods

### CacheClient Methods

| Method | Description | Example |
|--------|-------------|---------|
| `Ping()` | Check Redis connection | `cache.Ping()` |
| `Set(ctx, key, value, expiration)` | Set string value | `cache.Set(ctx, "k", "v", time.Hour)` |
| `Get(ctx, key)` | Get string value | `cache.Get(ctx, "k")` |
| `Delete(ctx, key)` | Delete key | `cache.Delete(ctx, "k")` |
| `Exists(ctx, keys...)` | Check if keys exist | `cache.Exists(ctx, "k1", "k2")` |
| `SetJSON(ctx, key, data, expiration)` | Set JSON value | `cache.SetJSON(ctx, "k", data, time.Hour)` |
| `GetJSON(ctx, key, &dest)` | Get JSON value | `cache.GetJSON(ctx, "k", &result)` |
| `SetWithTTL(ctx, key, value, ttl)` | Set with TTL | `cache.SetWithTTL(ctx, "k", "v", time.Minute*5)` |
| `GetTTL(ctx, key)` | Get remaining TTL | `cache.GetTTL(ctx, "k")` |
| `Expire(ctx, key, expiration)` | Set expiration | `cache.Expire(ctx, "k", time.Hour)` |
| `SetNX(ctx, key, value, expiration)` | Set if not exists | `cache.SetNX(ctx, "k", "v", time.Hour)` |
| `Incr(ctx, key)` | Increment counter | `cache.Incr(ctx, "count")` |
| `Decr(ctx, key)` | Decrement counter | `cache.Decr(ctx, "count")` |
| `SAdd(ctx, key, members...)` | Add to set | `cache.SAdd(ctx, "s", "m1", "m2")` |
| `SRem(ctx, key, members...)` | Remove from set | `cache.SRem(ctx, "s", "m1")` |
| `SMembers(ctx, key)` | Get set members | `cache.SMembers(ctx, "s")` |
| `SIsMember(ctx, key, member)` | Check set membership | `cache.SIsMember(ctx, "s", "m1")` |
| `LPush(ctx, key, values...)` | Push to list | `cache.LPush(ctx, "l", "v1", "v2")` |
| `RPop(ctx, key)` | Pop from list | `cache.RPop(ctx, "l")` |
| `LRange(ctx, key, start, stop)` | Get list range | `cache.LRange(ctx, "l", 0, -1)` |
| `LLen(ctx, key)` | Get list length | `cache.LLen(ctx, "l")` |
| `HSet(ctx, key, values...)` | Set hash fields | `cache.HSet(ctx, "h", "f1", "v1")` |
| `HGet(ctx, key, field)` | Get hash field | `cache.HGet(ctx, "h", "f1")` |
| `HGetAll(ctx, key)` | Get all hash fields | `cache.HGetAll(ctx, "h")` |
| `HDel(ctx, key, fields...)` | Delete hash fields | `cache.HDel(ctx, "h", "f1")` |
| `Keys(ctx, pattern)` | Get keys by pattern | `cache.Keys(ctx, "user:*")` |
| `FlushDB(ctx)` | Flush database | `cache.FlushDB(ctx)` |

## 🗂️ Redis Key Structure (Recommended)

```
# General cache
cache:{module}:{key}              → General purpose cache

# User data
user:{id}                         → User profile
user:{id}:session                 → User session

# Trading data
prices:{symbol}                   → Latest price
prices:{symbol}:{interval}        → Candlestick data
signals:active:{symbol}:{interval} → Active signals (future)
signals:history:{symbol}:{interval} → Historical signals (future)

# Watchlist
watchlist:active                  → Active watchlist symbols

# WebSocket
websocket:status                  → Connection status
websocket:streams                 → Active streams

# Counters
counter:{name}                    → Generic counter

# Locks
lock:{resource}                   → Distributed lock
```

## 🧪 Running Tests

```bash
# Make sure Redis is running
docker start trading-bot-redis

# Run all tests
go test -v ./internal/database/...

# Run specific test
go test -v -run TestCacheClient_Basic ./internal/database/...

# Run with coverage
go test -v -cover ./internal/database/...
```

## ⚠️ Important Notes

1. **Redis is Optional**: Application akan tetap berjalan jika Redis tidak tersedia. Error akan di-log dan Redis client akan `nil`.

2. **Connection Pooling**: Redis client menggunakan connection pooling otomatis. Tidak perlu manual manage connections.

3. **Context Usage**: Selalu gunakan `context.Context` untuk semua operations agar bisa di-cancel jika diperlukan.

4. **Expiration**: Selalu set expiration time untuk keys yang tidak perlu permanent untuk menghindari memory leak.

5. **Key Naming**: Gunakan naming convention yang konsisten untuk memudahkan debugging dan maintenance.

## 🔍 Monitoring

Check Redis connection status:

```bash
# Redis CLI
redis-cli ping
redis-cli info

# From application
err := container.Redis.Ping()
if err != nil {
    log.Printf("Redis connection issue: %v", err)
}
```

## 📚 References

- [Redis Documentation](https://redis.io/documentation)
- [go-redis Documentation](https://redis.uptrace.dev/)
- [Redis Commands](https://redis.io/commands/)
