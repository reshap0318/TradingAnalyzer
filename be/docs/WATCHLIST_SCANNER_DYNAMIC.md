# Watchlist Background Scanner (Simple)

Feature untuk melakukan scanning otomatis terhadap semua symbol di watchlist dan mengeksekusi trade berdasarkan signal yang valid.

---

## 📋 Overview

**Background Scanner** adalah fitur yang:
1. Scan semua symbol di watchlist secara berkala (interval dinamis)
2. Call `TradeExecute` untuk setiap symbol
3. Execute trade jika signal valid
4. Log semua aktivitas ke file log
5. **Tidak menyimpan state ke database** (in-memory only)

---

## 🔧 API Endpoints

### **1. Activate Scanner**

**Endpoint:** `POST /api/v1/watchlist/activate`

**Request Body:**
```json
{
  "strategy_id": 1  // Optional: jika kosong, akan menggunakan active strategy
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Scanner activated successfully",
  "data": {
    "is_active": true,
    "message": "Scanner activated successfully",
    "scan_interval": 15,  // Interval dalam menit (dari primary timeframe strategy)
    "strategy_id": 1      // Strategy ID yang digunakan (active strategy jika tidak dikirim)
  }
}
```

**Notes:**
- Jika `strategy_id` tidak dikirim atau 0, scanner akan menggunakan **active strategy**
- Response akan include `strategy_id` yang digunakan
- `scan_interval` ditentukan oleh primary timeframe dari strategy

---

### **2. Deactivate Scanner**

**Endpoint:** `POST /api/v1/watchlist/deactivate`

**Response:**
```json
{
  "code": 200,
  "message": "Scanner deactivated successfully",
  "data": {
    "is_active": false,
    "message": "Scanner deactivated successfully"
  }
}
```

---

### **3. Get Status**

**Endpoint:** `GET /api/v1/watchlist/status`

**Response:**
```json
{
  "code": 200,
  "message": "Scanner status retrieved successfully",
  "data": {
    "is_active": true
  }
}
```

---

## 🔄 Flow Scanner

```
1. User Activate Scanner (dengan/tanpa strategy_id)
   ↓
2. Jika strategy_id kosong:
   - Get Active Strategy dari database
   ↓
3. Get Strategy detail (primary_tf)
   ↓
4. Get Primary Timeframe detail (in_minutes)
   ↓
5. Set scan interval = timeframe.in_minutes
   ↓
6. Set in-memory flag: scannerActive = true
   ↓
7. Start Background Goroutine
   - Scan interval: dinamis (dari strategy)
   - Running until deactivated
   ↓
8. Every X Minutes (based on interval):
   a. Get active symbols from watchlist (DB)
   b. For each symbol:
      - Call TradeExecute()
      - Log result
   ↓
9. Log to ./logs/watchlist_scanner.log
```

---

## ⏱️ Scan Interval

Scan interval ditentukan oleh **primary timeframe** dari strategy yang digunakan:

| Strategy Primary TF | in_minutes | Scan Interval |
|---------------------|------------|---------------|
| 1m                  | 1          | 1 menit       |
| 5m                  | 5          | 5 menit       |
| 15m                 | 15         | 15 menit      |
| 30m                 | 30         | 30 menit      |
| 1h                  | 60         | 60 menit      |
| 4h                  | 240        | 240 menit     |
| 1d                  | 1440       | 1440 menit    |

**Jika tidak ada strategy_id:**
- Default scan interval: **5 menit**

---

## 📝 Log Format

File: `./logs/watchlist_scanner.log`

```
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Background scanner started (interval: 15 minutes)
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Starting scan cycle...
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Found 5 active symbols in watchlist
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Scanning symbol: BTCUSDT
[2025-03-10T14:35:02Z] [SCANNER] [INFO] ✅ Trade executed for BTCUSDT
[2025-03-10T14:35:02Z] [SCANNER] [INFO] Scanning symbol: ETHUSDT
[2025-03-10T14:35:04Z] [SCANNER] [INFO] ⏭️ No trade for ETHUSDT
[2025-03-10T14:50:00Z] [SCANNER] [INFO] Starting scan cycle... (next cycle after 15 minutes)
```

---

## 🔐 Safety Features

1. ✅ **Single Instance** - Mutex lock untuk prevent race condition
2. ✅ **Auto-Stop** - Berhenti saat context di-cancel
3. ✅ **Rate Limiting** - 2 detik delay antar symbol
4. ✅ **Error Handling** - Error di-log, scanner tetap jalan
5. ✅ **No Database** - In-memory state, no persistence
6. ✅ **Dynamic Interval** - Interval menyesuaikan strategy timeframe

---

## 🧪 Testing

```bash
# 1. Add symbol to watchlist
curl -X POST http://localhost:8000/api/v1/watchlists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"symbol": "BTCUSDT", "is_active": true}'

# 2. Activate scanner dengan strategy (interval dinamis)
curl -X POST http://localhost:8000/api/v1/watchlist/activate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"strategy_id": 1}'

# Response akan include scan_interval dari primary timeframe strategy

# 3. Check status
curl -X GET http://localhost:8000/api/v1/watchlist/status \
  -H "Authorization: Bearer YOUR_TOKEN"

# 4. Monitor logs
tail -f ./logs/watchlist_scanner.log

# 5. Deactivate scanner
curl -X POST http://localhost:8000/api/v1/watchlist/deactivate \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📊 Example Scenarios

### **Scenario 1: Strategy dengan Primary TF 15m**

```bash
# Activate dengan strategy_id=1 (primary_tf = "15m")
curl -X POST http://localhost:8000/api/v1/watchlist/activate \
  -H "Content-Type: application/json" \
  -d '{"strategy_id": 1}'

# Response:
{
  "is_active": true,
  "scan_interval": 15,  # Scan setiap 15 menit
  "strategy_id": 1
}

# Log:
[INFO] Background scanner started (interval: 15 minutes)
```

### **Scenario 2: Tanpa Strategy (Gunakan Active Strategy)**

```bash
# Activate tanpa strategy_id (akan menggunakan active strategy)
curl -X POST http://localhost:8000/api/v1/watchlist/activate \
  -H "Content-Type: application/json" \
  -d '{}'

# Response (misal active strategy punya primary_tf = "1h"):
{
  "is_active": true,
  "scan_interval": 60,  # Scan setiap 60 menit (dari active strategy)
  "strategy_id": 3      # Active strategy ID
}

# Log:
[INFO] Background scanner started (interval: 60 minutes)
```

### **Scenario 3: Error - No Active Strategy**

```bash
# Activate tanpa strategy_id dan tidak ada active strategy
curl -X POST http://localhost:8000/api/v1/watchlist/activate \
  -H "Content-Type: application/json" \
  -d '{}'

# Response:
{
  "code": 400,
  "message": "Failed to activate scanner",
  "error": "failed to get active strategy: no active strategy found"
}
```

---

## ⚠️ Important Notes

1. **State tidak persistent** - Scanner akan stop jika aplikasi restart
2. **No database tracking** - Tidak ada statistik atau history
3. **In-memory mutex** - Thread-safe untuk concurrent access
4. **Uses existing Watchlist** - Symbol diambil dari watchlist yang sudah ada
5. **Dynamic Interval** - Interval scan tergantung strategy primary timeframe
6. **Active Strategy Fallback** - Jika strategy_id tidak dikirim, akan menggunakan active strategy
7. **Error jika No Active Strategy** - Harus ada active strategy jika tidak kirim strategy_id

---

## 📚 Files Modified

- `internal/service/watchlist_scanner_service.go` - Scanner logic dengan dynamic interval
- `internal/service/watchlist_service.go` - Cleaned up (hanya CRUD)
- `internal/controller/watchlist_controller.go` - Scanner endpoints
- `internal/routes/watchlist_routes.go` - Scanner routes

---

## 📞 Support

For issues or questions, create an issue in the repository.
