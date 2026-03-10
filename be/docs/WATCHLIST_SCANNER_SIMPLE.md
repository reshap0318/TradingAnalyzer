# Watchlist Background Scanner (Simple)

Feature untuk melakukan scanning otomatis terhadap semua symbol di watchlist dan mengeksekusi trade berdasarkan signal yang valid.

---

## 📋 Overview

**Background Scanner** adalah fitur yang:
1. Scan semua symbol di watchlist secara berkala (setiap 5 menit)
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
  "strategy_id": 1  // Optional
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Scanner activated successfully",
  "data": {
    "is_active": true,
    "message": "Scanner activated successfully"
  }
}
```

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
1. User Activate Scanner
   ↓
2. Set in-memory flag: scannerActive = true
   ↓
3. Start Background Goroutine
   - Scan interval: 5 minutes
   - Running until deactivated
   ↓
4. Every 5 Minutes:
   a. Get active symbols from watchlist (DB)
   b. For each symbol:
      - Call TradeExecute()
      - Log result
   ↓
5. Log to ./logs/watchlist_scanner.log
```

---

## 📝 Log Format

File: `./logs/watchlist_scanner.log`

```
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Background scanner started
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Starting scan cycle...
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Found 5 active symbols in watchlist
[2025-03-10T14:35:00Z] [SCANNER] [INFO] Scanning symbol: BTCUSDT
[2025-03-10T14:35:02Z] [SCANNER] [INFO] ✅ Trade executed for BTCUSDT
[2025-03-10T14:35:02Z] [SCANNER] [INFO] Scanning symbol: ETHUSDT
[2025-03-10T14:35:04Z] [SCANNER] [INFO] ⏭️ No trade for ETHUSDT
```

---

## ⚙️ Configuration

| Setting | Default | Location |
|---------|---------|----------|
| Scan Interval | 5 minutes | `watchlist_service.go:172` |
| Delay Between Symbols | 2 seconds | `watchlist_service.go:233` |

---

## 🔐 Safety Features

1. ✅ **Single Instance** - Mutex lock untuk prevent race condition
2. ✅ **Auto-Stop** - Berhenti saat context di-cancel
3. ✅ **Rate Limiting** - 2 detik delay antar symbol
4. ✅ **Error Handling** - Error di-log, scanner tetap jalan
5. ✅ **No Database** - In-memory state, no persistence

---

## 🧪 Testing

```bash
# 1. Add symbol to watchlist
curl -X POST http://localhost:8000/api/v1/watchlists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"symbol": "BTCUSDT", "is_active": true}'

# 2. Activate scanner
curl -X POST http://localhost:8000/api/v1/watchlist/activate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"strategy_id": 1}'

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

## ⚠️ Important Notes

1. **State tidak persistent** - Scanner akan stop jika aplikasi restart
2. **No database tracking** - Tidak ada statistik atau history
3. **In-memory mutex** - Thread-safe untuk concurrent access
4. **Uses existing Watchlist** - Symbol diambil dari watchlist yang sudah ada

---

## 📚 Files Modified

- `internal/service/watchlist_service.go` - Scanner logic
- `internal/controller/watchlist_controller.go` - Scanner endpoints
- `internal/routes/watchlist_routes.go` - Scanner routes

---

## 📞 Support

For issues or questions, create an issue in the repository.
