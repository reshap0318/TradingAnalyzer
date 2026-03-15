# API Documentation

## 📋 Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Response Format](#response-format)
4. [Endpoints](#endpoints)
   - [Health Check](#health-check)
   - [Timeframes](#timeframes)
   - [Indicators](#indicators)
   - [Thresholds](#thresholds)
   - [Configs](#configs)
   - [Trade (Bot & Execution)](#trade-bot--execution)
     - [Trade Execute](#post-apitradeexecute)
     - [Trade Monitor](#trade-monitor)
     - [Trade Bot Control](#trade-bot-control)
     - [Trade Bot Get All](#get-apitradebot)
   - [Trade Response DTOs](#trade-response-dtos)
     - [TradeData](#tradedata)
     - [TradeDayStat](#tradedaystat)
     - [OrderInfo](#orderinfo)
   - [Watchlists](#watchlists)
   - [Strategies](#strategies)

---

## 🌐 Overview

**Base URL:** `http://localhost:8000/api`

**API Version:** v1

**Content Type:** `application/json`

---

## 🔐 Authentication

All endpoints (except `/health`) require authentication via Bearer Token.

**Header:**
```
Authorization: Bearer <your_token>
```

---

## 📦 Response Format

### **Success Response**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

### **Error Response**
```json
{
    "code": 400,
    "message": "Error message here",
    "error": null
}
```

### **Validation Error Response**
```json
{
    "code": 400,
    "message": "Format JSON tidak valid",
    "error": null
}
```

---

## 🏥 Health Check

### **GET /health**

Check if the API is running.

**Request:**
```bash
GET http://localhost:8000/health
```

**Response:**
```json
{
    "status": "ok"
}
```

---

## ⏱️ Timeframes

Manage trading timeframes (e.g., 1m, 5m, 1h, 1d).

**Identifier:** `name` (string) - e.g., "1m", "5m", "15m"

### **GET /api/timeframes**

Get all timeframes.

**Request:**
```bash
GET http://localhost:8000/api/timeframes
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "name": "1m",
            "in_minutes": 1,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "name": "5m",
            "in_minutes": 5,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "name": "15m",
            "in_minutes": 15,
            "created_at": "2025-03-08T10:00:00Z"
        }
    ]
}
```

---

### **GET /api/timeframes/:name**

Get timeframe by name.

**Request:**
```bash
GET http://localhost:8000/api/timeframes/5m
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "name": "5m",
        "in_minutes": 5,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **POST /api/timeframes**

Create new timeframe.

**Request:**
```bash
POST http://localhost:8000/api/timeframes
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "10m",
    "in_minutes": 10
}
```

**Request Body:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `name` | string | Yes | max: 5 characters |
| `in_minutes` | integer | Yes | greater than 0 |

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "name": "10m",
        "in_minutes": 10,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **PUT /api/timeframes/:name**

Update timeframe.

**Request:**
```bash
PUT http://localhost:8000/api/timeframes/10m
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "12m",
    "in_minutes": 12
}
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "name": "12m",
        "in_minutes": 12,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **DELETE /api/timeframes/:name**

Delete timeframe.

**Request:**
```bash
DELETE http://localhost:8000/api/timeframes/12m
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "name": "12m",
        "in_minutes": 12,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

## 📊 Indicators

Manage trading indicators (e.g., RSI, MACD, Stochastic).

### **GET /api/indicators**

Get all indicators.

**Request:**
```bash
GET http://localhost:8000/api/indicators
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "name": "Moving Average",
            "indicator": "moving_average",
            "description": "Moving Average - Trend indicator using multiple SMAs and EMAs",
            "params": {
                "sma_periods": [20, 50, 200],
                "ema_periods": [12, 26]
            },
            "is_active": true,
            "weight": 0.30,
            "order_view": 1,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 2,
            "name": "MACD",
            "indicator": "macd",
            "description": "Moving Average Convergence Divergence",
            "params": {
                "fast_period": 12,
                "slow_period": 26,
                "signal_period": 9
            },
            "is_active": true,
            "weight": 0.22,
            "order_view": 2,
            "created_at": "2025-03-08T10:00:00Z"
        }
    ]
}
```

---

### **GET /api/indicators/:id**

Get indicator by ID.

**Request:**
```bash
GET http://localhost:8000/api/indicators/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "name": "Moving Average",
        "indicator": "moving_average",
        "description": "Moving Average - Trend indicator",
        "params": {
            "sma_periods": [20, 50, 200],
            "ema_periods": [12, 26]
        },
        "is_active": true,
        "weight": 0.30,
        "order_view": 1,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **POST /api/indicators**

Create new indicator.

**Request:**
```bash
POST http://localhost:8000/api/indicators
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "Custom RSI",
    "indicator": "custom_rsi",
    "description": "Custom RSI indicator",
    "params": {
        "period": 14,
        "overbought": 70,
        "oversold": 30
    },
    "is_active": true,
    "weight": 0.15,
    "order_view": 10
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique indicator name |
| `indicator` | string | Yes | Indicator key/identifier |
| `description` | string | No | Description |
| `params` | object | No | JSON parameters |
| `is_active` | boolean | No | Default: true |
| `weight` | float | No | Default: 1.0 |
| `order_view` | integer | No | Display order |

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 10,
        "name": "Custom RSI",
        "indicator": "custom_rsi",
        "description": "Custom RSI indicator",
        "params": {
            "period": 14,
            "overbought": 70,
            "oversold": 30
        },
        "is_active": true,
        "weight": 0.15,
        "order_view": 10,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **PUT /api/indicators/:id**

Update indicator.

**Request:**
```bash
PUT http://localhost:8000/api/indicators/10
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "Custom RSI Updated",
    "indicator": "custom_rsi",
    "description": "Updated custom RSI indicator",
    "params": {
        "period": 14,
        "overbought": 70,
        "oversold": 30
    },
    "is_active": true,
    "weight": 0.20,
    "order_view": 10
}
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 10,
        "name": "Custom RSI Updated",
        "indicator": "custom_rsi",
        ...
    }
}
```

---

### **DELETE /api/indicators/:id**

Delete indicator.

**Request:**
```bash
DELETE http://localhost:8000/api/indicators/10
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

## 🎯 Thresholds

Manage signal thresholds (e.g., STRONG_BUY, BUY, WAIT, SELL, STRONG_SELL).

### **GET /api/thresholds**

Get all thresholds.

**Request:**
```bash
GET http://localhost:8000/api/thresholds
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "category": "STRONG_BUY",
            "min_value": 70,
            "max_value": 100,
            "action": "BUY",
            "color": "green",
            "order_display": 1,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 2,
            "category": "BUY",
            "min_value": 45,
            "max_value": 70,
            "action": "BUY",
            "color": "light-green",
            "order_display": 2,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 3,
            "category": "WAIT",
            "min_value": -45,
            "max_value": 45,
            "action": "WAIT",
            "color": "gray",
            "order_display": 3,
            "created_at": "2025-03-08T10:00:00Z"
        }
    ]
}
```

---

### **GET /api/thresholds/:id**

Get threshold by ID.

**Request:**
```bash
GET http://localhost:8000/api/thresholds/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "category": "STRONG_BUY",
        "min_value": 70,
        "max_value": 100,
        "action": "BUY",
        "color": "green",
        "order_display": 1,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **POST /api/thresholds**

Create new threshold.

**Request:**
```bash
POST http://localhost:8000/api/thresholds
Authorization: Bearer <token>
Content-Type: application/json

{
    "category": "CUSTOM_BUY",
    "min_value": 60,
    "max_value": 75,
    "action": "BUY",
    "color": "blue",
    "order_display": 6
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `category` | string | Yes | Unique category name |
| `min_value` | integer | Yes | Minimum value |
| `max_value` | integer | Yes | Maximum value |
| `action` | string | Yes | BUY/SELL/WAIT |
| `color` | string | Yes | Display color |
| `order_display` | integer | Yes | Display order |

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **PUT /api/thresholds/:id**

Update threshold.

**Request:**
```bash
PUT http://localhost:8000/api/thresholds/1
Authorization: Bearer <token>
Content-Type: application/json

{
    "category": "STRONG_BUY",
    "min_value": 75,
    "max_value": 100,
    "action": "BUY",
    "color": "dark-green",
    "order_display": 1
}
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **DELETE /api/thresholds/:id**

Delete threshold.

**Request:**
```bash
DELETE http://localhost:8000/api/thresholds/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

## ⚙️ Configs

Manage system configurations (Money Management, Binance settings).

### **GET /api/configs**

Get all configs.

**Request:**
```bash
GET http://localhost:8000/api/configs
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "config_key": "MIN_CONFIDENCE",
            "value": "45",
            "category": "MONEY_MANAGEMENT",
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 2,
            "config_key": "MAX_DAILY_TRADES",
            "value": "10",
            "category": "MONEY_MANAGEMENT",
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 3,
            "config_key": "LEVERAGE",
            "value": "5",
            "category": "MONEY_MANAGEMENT",
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 14,
            "config_key": "BINANCE_TESTNET",
            "value": "true",
            "category": "BINANCE",
            "created_at": "2025-03-08T10:00:00Z"
        }
    ]
}
```

---

### **GET /api/configs/:id**

Get config by ID.

**Request:**
```bash
GET http://localhost:8000/api/configs/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "config_key": "MIN_CONFIDENCE",
        "value": "45",
        "category": "MONEY_MANAGEMENT",
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **POST /api/configs**

Create new config.

**Request:**
```bash
POST http://localhost:8000/api/configs
Authorization: Bearer <token>
Content-Type: application/json

{
    "config_key": "CUSTOM_SETTING",
    "value": "100",
    "category": "CUSTOM"
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `config_key` | string | Yes | Unique key (max: 100) |
| `value` | string | Yes | Configuration value |
| `category` | string | Yes | Category (max: 50) |

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **PUT /api/configs/:id**

Update config.

**Request:**
```bash
PUT http://localhost:8000/api/configs/1
Authorization: Bearer <token>
Content-Type: application/json

{
    "config_key": "MIN_CONFIDENCE",
    "value": "50",
    "category": "MONEY_MANAGEMENT"
}
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **DELETE /api/configs/:id**

Delete config.

**Request:**
```bash
DELETE http://localhost:8000/api/configs/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

## 🤖 Trade (Bot & Execution)

Trade bot untuk automated trading dan manual execution.

### **POST /api/trade/execute**

Execute single trade manually (trigger trading bot untuk symbol tertentu).

**Request:**
```bash
POST http://localhost:8000/api/trade/execute
Authorization: Bearer <token>
Content-Type: application/json

{
    "symbol": "BTCUSDT",
    "strategy_id": 1
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `strategy_id` | integer | No | Strategy ID. Jika tidak disediakan, akan menggunakan active strategy |

**Response:**

Response mengikuti format **SignalAnalyze Response** dengan struktur lengkap:

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "symbol": "BTCUSDT",
        "primary_timeframe": "15m",
        "timestamp": "2025-03-10T08:30:00Z",
        "signal": {
            "valid": true,
            "signal": "BUY",
            "current_price": 50000.00,
            "trading_plan": {
                "mode": "CONSERVATIVE",
                "entries": [...],
                "take_profit": 50750.00,
                "stop_loss": 48511.00,
                "risk_reward_ratio": 0.51,
                "buffer_percent": 1.50,
                "summary": {
                    "total_entries": 1,
                    "total_position_value": 400.00,
                    "max_risk_usdt": 59.16,
                    "risk_from_capital": 14.79,
                    "target_profit_usdt": 30.40,
                    "profit_from_capital": 7.60,
                    "effective_leverage": 5.00
                }
            }
        },
        "scoring": {
            "totalScore": 0.75,
            "confidence": 75.00,
            "breakdown": [...]
        },
        "execution_info": {
            "executed": true,
            "message": "Order placed successfully",
            "margin_type": "ISOLATED",
            "leverage": 5,
            "capital_used": 400.00,
            "orders": [...],
            "tp_order_id": 98765432,
            "sl_order_id": 87654321
        }
    }
}
```

**📚 Response Documentation:**

Untuk dokumentasi lengkap tentang struktur response (signal, scoring, trading_plan), lihat:
- **[SIGNAL_ANALYZE_RESPONSE.md](./SIGNAL_ANALYZE_RESPONSE.md)** - Complete response structure
- **[SIGNAL_BREAKDOWN.md](./SIGNAL_BREAKDOWN.md)** - Signal calculation per indicator

**Key Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | string | Trading symbol |
| `primary_timeframe` | string | Primary timeframe used |
| `timestamp` | string | Execution timestamp |
| `signal` | object | **SignalInfo** - Trading signal dengan plan lengkap |
| `signal.valid` | boolean | Apakah signal valid (confidence >= threshold) |
| `signal.signal` | string | BUY/SELL/STRONG_BUY/STRONG_SELL/WAIT |
| `signal.trading_plan` | object | **TradingPlan** lengkap dengan entry, TP, SL |
| `signal.trading_plan.summary` | object | **Pre-calculated summary** (risk, profit, leverage) |
| `scoring` | object | **ScoringBreakdown** - Breakdown score per indicator |
| `execution_info` | object | Execution details (orders, TP/SL IDs) |
| `execution_info.executed` | boolean | Whether trade was executed |
| `execution_info.orders` | array | List of orders placed |
| `execution_info.tp_order_id` | integer | Take-profit order ID |
| `execution_info.sl_order_id` | integer | Stop-loss order ID |

**💡 Quick Reference:**
- `signal.trading_plan.summary.risk_from_capital` = Risk sebenarnya (% dari modal)
- `signal.trading_plan.summary.profit_from_capital` = Profit yang diharapkan (% dari modal)
- `signal.trading_plan.summary.effective_leverage` = Leverage aktual yang digunakan
- `scoring.confidence` = Confidence level (0-100)

---

### **Trade Monitor**

Monitor dan trigger trades secara manual (untuk debugging).

#### **POST /api/trade/monitor/all**

Process all active trades (background monitoring).

**Request:**
```bash
POST http://localhost:8000/api/trade/monitor/all
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "Trade monitoring completed",
    "data": {
        "total_processed": 5,
        "results": [
            {
                "trade_id": 1,
                "symbol": "BTCUSDT",
                "status": "ACTIVE",
                "message": "Trade processed successfully",
                "entries_sync": 1,
                "tp_updated": true,
                "sl_updated": true,
                "updated_count": 3,
                "logs": [
                    "Fetching trade details...",
                    "Syncing entry orders...",
                    "Updating TP order...",
                    "Updating SL order..."
                ]
            }
        ]
    }
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `total_processed` | integer | Total trades processed |
| `results` | array | List of trade processing results |
| `results[].trade_id` | integer | Trade ID |
| `results[].symbol` | string | Trading symbol |
| `results[].status` | string | Trade status (ACTIVE/COMPLETED/CANCELLED) |
| `results[].message` | string | Processing message |
| `results[].entries_sync` | integer | Number of entries synced |
| `results[].tp_updated` | boolean | Whether TP was updated |
| `results[].sl_updated` | boolean | Whether SL was updated |
| `results[].updated_count` | integer | Total updates performed |
| `results[].logs` | array | Detailed execution logs |

---

#### **POST /api/trade/monitor/:id**

Process single trade by ID (manual trigger untuk debugging).

**Request:**
```bash
POST http://localhost:8000/api/trade/monitor/1
Authorization: Bearer <token>
Content-Type: application/json

{
    "trade_id": 1
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `trade_id` | integer | Yes | Trade ID to process |

**Response:**
```json
{
    "code": 200,
    "message": "Trade processed successfully",
    "data": {
        "trade_id": 1,
        "symbol": "BTCUSDT",
        "status": "ACTIVE",
        "message": "Trade processed successfully",
        "entries_sync": 1,
        "tp_updated": true,
        "sl_updated": true,
        "updated_count": 3,
        "logs": [...]
    }
}
```

---

#### **POST /api/trade/monitor/:id/close**

Close an active trade manually by user request (Manual Close).

**Request:**
```bash
POST http://localhost:8000/api/trade/monitor/1/close
Authorization: Bearer <token>
```

**Request Parameters:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` (URL param) | integer | Yes | Trade ID to close |

**Response:**
```json
{
    "code": 200,
    "message": "Trade closed manually successfully",
    "data": {
        "trade_id": 1,
        "symbol": "BTCUSDT",
        "status": "CLOSED",
        "message": "Trade closed manually by user.",
        "entries_sync": 1,
        "tp_updated": false,
        "sl_updated": false,
        "updated_count": 0,
        "logs": [
            "Starting Manual Close for trade #1 (BTCUSDT)",
            "Syncing exact filled quantities from Binance open orders...",
            "Canceling ALL open orders for symbol BTCUSDT...",
            "All open orders canceled successfully.",
            "Executing Market SELL for TotalQty: 0.01 to close position...",
            "SUCCESS Market Close. OrderID: 12345678"
        ]
    }
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `trade_id` | integer | Trade ID that was closed |
| `symbol` | string | Trading symbol |
| `status` | string | Final trade status (CLOSED/SKIPPED) |
| `message` | string | Closing message |
| `entries_sync` | integer | Number of entries synced before close |
| `tp_updated` | boolean | Whether TP was updated (always false for manual close) |
| `sl_updated` | boolean | Whether SL was updated (always false for manual close) |
| `updated_count` | integer | Total updates performed during sync |
| `logs` | array | Detailed execution logs of the close process |

**Process Flow:**

1. **Fetch Trade** - Load trade with entries from database
2. **Validate Status** - Only ACTIVE trades can be closed manually
3. **Sync Entries** - Sync filled quantities from Binance open orders (Fase 2)
4. **Netting Calculation** - Calculate final TotalQty & AvgEntryPrice (Fase 3)
5. **Cancel Orders** - Cancel ALL pending open orders on Binance
6. **Market Close** - Execute market order to close position (ReduceOnly)
7. **Update Database** - Update trade status to CLOSED, record PnL, clear TP/SL order IDs

**Exit Info:**
- `ExitPrice`: Current market price at close time
- `ExitReason`: "MANUAL_CLOSE_BY_USER"
- `PnL`: Calculated profit/loss from close
- `PnLPct`: PnL percentage
- `Status`: Updated to "CLOSED"
- `ClosedAt`: Timestamp of close

**⚠️ Notes:**
- This endpoint will **cancel all pending orders** before executing market close
- Uses **ReduceOnly** order to close existing position
- If trade status is not ACTIVE, returns status "SKIPPED"
- PnL is calculated and recorded in database

---

### **Trade Bot Control**

Control automated trading bot.

#### **GET /api/trade/bot/status**

Get bot status (active/inactive, config, etc).

**Request:**
```bash
GET http://localhost:8000/api/trade/bot/status
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "Trade bot status retrieved successfully",
    "data": {
        "id": 1,
        "is_active": true,
        "active_since": "2025-03-10T08:00:00Z",
        "strategy_id": 1,
        "last_scan": "2025-03-10T12:30:00Z",
        "trades_executed": 15,
        "scan_interval": 15
    }
}
```

---

#### **GET /api/trade/bot/status**

Get current trade bot status.

**Request:**
```bash
GET http://localhost:8000/api/trade/bot/status
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "is_active": true,
        "strategy": {
            "id": 1,
            "strategy_name": "Day Trading Pro",
            "primary_tf": "15m",
            "is_active": true,
            "created_at": "2025-03-08T10:00:00Z",
            "updated_at": "2025-03-10T08:00:00Z",
            "timeframes": [
                {
                    "tf": "15m",
                    "weight": 0.6
                },
                {
                    "tf": "1h",
                    "weight": 0.4
                }
            ],
            "indicator_weights": [
                {
                    "indicator_id": 1,
                    "weight": 0.3
                },
                {
                    "indicator_id": 2,
                    "weight": 0.2
                }
            ],
            "money_management": {
                "min_confidence": 45,
                "max_daily_trades": 10,
                "max_daily_loss_percent": 5,
                "max_position_size": 0.15,
                "risk_reward_ratio": 1.5,
                "leverage": 5,
                "is_agressive": false,
                "order_expiration_hours": 4
            }
        },
        "trade_executor": {
            "is_running": true,
            "started_at": "2025-03-14T10:30:00Z",
            "duration": "5m30s",
            "duration_sec": 330.5
        },
        "trade_monitor": {
            "is_running": false
        }
    }
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_active` | boolean | Whether trade bot service is active |
| `strategy` | object | Active strategy details (null if bot not active) |
| `strategy.id` | integer | Strategy ID |
| `strategy.strategy_name` | string | Strategy name |
| `strategy.primary_tf` | string | Primary timeframe |
| `strategy.is_active` | boolean | Whether strategy is marked as active |
| `strategy.timeframes` | array | List of timeframes with weights |
| `strategy.indicator_weights` | array | List of indicators with weights |
| `strategy.money_management` | object | Money management configuration |
| `trade_executor` | object | Trade executor worker status |
| `trade_executor.is_running` | boolean | Whether executor cycle is currently running |
| `trade_executor.started_at` | string | RFC3339 timestamp when current cycle started (null if not running) |
| `trade_executor.duration` | string | Human-readable duration since cycle started (e.g., "5m30s") |
| `trade_executor.duration_sec` | number | Duration in seconds (float) |
| `trade_monitor` | object | Trade monitor worker status |
| `trade_monitor.is_running` | boolean | Whether monitor cycle is currently running |
| `trade_monitor.started_at` | string | RFC3339 timestamp when current cycle started (null if not running) |
| `trade_monitor.duration` | string | Human-readable duration since cycle started |
| `trade_monitor.duration_sec` | number | Duration in seconds (float) |

**Example: Bot Not Active**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "is_active": false,
        "strategy": null,
        "trade_executor": {
            "is_running": false
        },
        "trade_monitor": {
            "is_running": false
        }
    }
}
```

**Example: Bot Active, Workers Idle**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "is_active": true,
        "strategy": { ... },
        "trade_executor": {
            "is_running": false
        },
        "trade_monitor": {
            "is_running": false
        }
    }
}
```

**Example: Bot Active, Executor Running**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "is_active": true,
        "strategy": { ... },
        "trade_executor": {
            "is_running": true
        },
        "trade_monitor": {
            "is_running": false
        },
        "bot_started_at": "2025-03-14T10:30:00Z"
    }
}
```

---

#### **GET /api/trade/bot/summary**

Get summary statistics for current trading session (since bot was activated).

**Request:**
```bash
GET http://localhost:8000/api/trade/bot/summary
Authorization: Bearer <token>
```

**Response (Success):**
```json
{
    "code": 200,
    "message": "Session summary retrieved successfully",
    "data": {
        "total_trades": 5,
        "executed": 4,
        "skipped": 1,
        "success_rate": 75.0,
        "total_pnl": 150.50,
        "symbols_traded": ["BTCUSDT", "ETHUSDT", "BNBUSDT"],
        "session_started": "2025-03-14T10:30:00Z"
    }
}
```

**Response (Bot Not Running):**
```json
{
    "code": 400,
    "message": "trade bot is not running. Please activate the bot first",
    "error": "trade bot is not running. Please activate the bot first"
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `total_trades` | integer | Total trades found in current session |
| `executed` | integer | Trades with status ACTIVE or COMPLETED |
| `skipped` | integer | Other trades (CANCELLED, REJECTED, etc.) |
| `success_rate` | number | Percentage of profitable COMPLETED trades (0-100) |
| `total_pnl` | number | Sum of PnL from all trades in session |
| `symbols_traded` | array | List of unique symbols traded |
| `session_started` | string | RFC3339 timestamp when bot was activated |

**Notes:**
- Returns **400 error** if bot is not active
- Filter: `created_at > bot_started_at` (all trades since bot activation)
- `success_rate` calculated from COMPLETED trades only
- `total_pnl` includes both profitable and losing trades

---

#### **GET /api/trade/bot/**

Get list of trades executed in current trading session.

**Request:**
```bash
GET http://localhost:8000/api/trade/bot/
Authorization: Bearer <token>
```

**Response (Success):**
```json
{
    "code": 200,
    "message": "Executed trades retrieved successfully",
    "data": [
        {
            "id": 1,
            "symbol": "BTCUSDT",
            "interval": "15m",
            "side": "LONG",
            "confidence": 85.5,
            "total_score": 0.92,
            "is_aggressive": false,
            "tp_price": 52000.00,
            "sl_price": 48000.00,
            "risk_reward_ratio": 1.5,
            "avg_entry_price": 50000.00,
            "leverage": 5,
            "capital_used": 1000.00,
            "total_qty": 0.02,
            "status": "ACTIVE",
            "description": "Strong buy signal with high confidence",
            "tp_order_id": 12345678,
            "sl_order_id": 87654321,
            "tp_sl_status": "ACTIVE",
            "exit_price": 0,
            "pnl": 0,
            "pnl_pct": 0,
            "created_at": "2025-03-14T10:35:00Z",
            "updated_at": "2025-03-14T10:35:00Z",
            "closed_at": null
        },
        {
            "id": 2,
            "symbol": "ETHUSDT",
            "interval": "15m",
            "side": "LONG",
            "confidence": 78.2,
            "total_score": 0.85,
            "is_aggressive": false,
            "tp_price": 3200.00,
            "sl_price": 3000.00,
            "risk_reward_ratio": 1.8,
            "avg_entry_price": 3100.00,
            "leverage": 5,
            "capital_used": 800.00,
            "total_qty": 0.25,
            "status": "COMPLETED",
            "description": "Good buy signal",
            "tp_order_id": 12345679,
            "sl_order_id": 87654322,
            "tp_sl_status": "TP_HIT",
            "exit_price": 3200.00,
            "pnl": 62.50,
            "pnl_pct": 7.81,
            "created_at": "2025-03-14T10:40:00Z",
            "updated_at": "2025-03-14T11:15:00Z",
            "closed_at": "2025-03-14T11:15:00Z"
        }
    ]
}
```

**Response (Bot Not Running):**
```json
{
    "code": 400,
    "message": "trade bot is not running. Please activate the bot first",
    "error": "trade bot is not running. Please activate the bot first"
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Trade ID |
| `symbol` | string | Trading symbol (e.g., "BTCUSDT") |
| `interval` | string | Timeframe used (e.g., "15m") |
| `side` | string | Trade direction (LONG/SHORT) |
| `confidence` | number | Confidence score (0-100) |
| `total_score` | number | Total signal score (0-1) |
| `is_aggressive` | boolean | Whether aggressive mode was used |
| `tp_price` | number | Take-profit price |
| `sl_price` | number | Stop-loss price |
| `risk_reward_ratio` | number | Risk-reward ratio |
| `avg_entry_price` | number | Average entry price |
| `leverage` | integer | Leverage used |
| `capital_used` | number | Capital used in USDT |
| `total_qty` | number | Total quantity purchased |
| `status` | string | Trade status (ACTIVE/COMPLETED/CANCELLED) |
| `description` | string | Trade description |
| `tp_order_id` | integer | Take-profit order ID |
| `sl_order_id` | integer | Stop-loss order ID |
| `tp_sl_status` | string | TP/SL status (ACTIVE/TP_HIT/SL_HIT) |
| `exit_price` | number | Exit price (if closed) |
| `pnl` | number | Profit/Loss in USDT |
| `pnl_pct` | number | Profit/Loss percentage |
| `created_at` | string | RFC3339 timestamp when trade was created |
| `updated_at` | string | RFC3339 timestamp when trade was last updated |
| `closed_at` | string | RFC3339 timestamp when trade was closed (null if active) |

**Notes:**
- Returns **400 error** if bot is not active
- Returns **empty array** if no trades executed yet
- Filter: `created_at > bot_started_at` (all trades since bot activation)
- Trades sorted by `created_at DESC` (newest first)

---

#### **GET /api/trade/bot/active**

Get list of currently active trades (all active trades, not limited to current session).

**Request:**
```bash
GET http://localhost:8000/api/trade/bot/active
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "Active trades retrieved successfully",
    "data": [
        {
            "id": 1,
            "symbol": "BTCUSDT",
            "interval": "15m",
            "side": "LONG",
            "status": "ACTIVE",
            "tp_price": 52000.00,
            "sl_price": 48000.00,
            "pnl": 25.50,
            "pnl_pct": 2.55,
            "created_at": "2025-03-14T10:35:00Z",
            "updated_at": "2025-03-14T10:50:00Z",
            "closed_at": null
        },
        {
            "id": 3,
            "symbol": "BNBUSDT",
            "interval": "1h",
            "side": "LONG",
            "status": "ACTIVE",
            "tp_price": 350.00,
            "sl_price": 330.00,
            "pnl": -5.20,
            "pnl_pct": -1.52,
            "created_at": "2025-03-14T09:00:00Z",
            "updated_at": "2025-03-14T10:45:00Z",
            "closed_at": null
        }
    ]
}
```

**Response Fields:**

Same as **GET /api/trade/bot/** endpoint.

**Notes:**
- Returns **all active trades** in database (not limited to current session)
- Filter: `status = 'ACTIVE'`
- Returns **empty array** if no active trades
- Trades sorted by `created_at DESC` (newest first)
- Does **NOT** require bot to be running (can check active trades anytime)

---

#### **GET /api/trade/bot**

Get all trades with optional filters.

**Request:**
```bash
GET http://localhost:8000/api/trade/bot?status=ACTIVE&symbol=BTCUSDT&interval=15m&min_confidence=70&side=BUY&date_start=2025-03-01&date_end=2025-03-15
Authorization: Bearer <token>
```

**Query Parameters (All Optional):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `status[]` | string | No | Filter by status. Can be multiple: `ACTIVE`, `CLOSED`, `CANCELLED`, `COMPLETED`, `REJECTED`. Example: `?status=ACTIVE&status=CLOSED` |
| `symbol[]` | string | No | Filter by symbol. Can be multiple: `BTCUSDT`, `ETHUSDT`, etc. Example: `?symbol=BTCUSDT&symbol=ETHUSDT` |
| `interval` | string | No | Filter by timeframe interval: `1m`, `5m`, `15m`, `1h`, `4h`, `1d`, etc. |
| `min_confidence` | number | No | Filter trades with confidence >= this value (0-100) |
| `side` | string | No | Filter by side: `BUY` or `SELL` |
| `date_start` | string | No | Filter trades created on or after this date (format: `YYYY-MM-DD`) |
| `date_end` | string | No | Filter trades created on or before this date (format: `YYYY-MM-DD`) |

**Response:**
```json
{
    "code": 200,
    "message": "Trades retrieved successfully",
    "data": [
        {
            "id": 1,
            "symbol": "BTCUSDT",
            "interval": "15m",
            "side": "BUY",
            "confidence": 85.5,
            "total_score": 0.92,
            "is_aggressive": false,
            "tp_price": 52000.00,
            "sl_price": 48000.00,
            "risk_reward_ratio": 1.5,
            "avg_entry_price": 50000.00,
            "leverage": 5,
            "capital_used": 1000.00,
            "total_qty": 0.02,
            "status": "ACTIVE",
            "description": "Strong buy signal with high confidence",
            "tp_order_id": 12345678,
            "sl_order_id": 87654321,
            "tp_sl_status": "ACTIVE",
            "exit_price": 0,
            "pnl": 0,
            "pnl_pct": 0,
            "created_at": "2025-03-14T10:35:00Z",
            "updated_at": "2025-03-14T10:35:00Z",
            "closed_at": null,
            "orders": [
                {
                    "entry_number": 1,
                    "binance_order_id": 987654321,
                    "price": 50000.00,
                    "quantity": 0.02,
                    "type": "LIMIT",
                    "status": "FILLED"
                }
            ]
        },
        {
            "id": 2,
            "symbol": "ETHUSDT",
            "interval": "1h",
            "side": "SELL",
            "confidence": 72.3,
            "total_score": 0.78,
            "is_aggressive": false,
            "tp_price": 3000.00,
            "sl_price": 3200.00,
            "risk_reward_ratio": 1.8,
            "avg_entry_price": 3100.00,
            "leverage": 5,
            "capital_used": 800.00,
            "total_qty": 0.25,
            "status": "CLOSED",
            "description": "Good sell signal",
            "tp_order_id": 12345679,
            "sl_order_id": 87654322,
            "tp_sl_status": "TP_HIT",
            "exit_price": 3000.00,
            "pnl": 62.50,
            "pnl_pct": 7.81,
            "created_at": "2025-03-14T10:40:00Z",
            "updated_at": "2025-03-14T11:15:00Z",
            "closed_at": "2025-03-14T11:15:00Z",
            "orders": []
        }
    ]
}
```

**Response Fields:**

Same as **GET /api/trade/bot/session** endpoint, with additional `orders` array.

| Field | Type | Description |
|-------|------|-------------|
| `orders` | array | List of entry orders for this trade (optional - only included if trade has entries) |
| `orders[].entry_number` | integer | Entry number (1, 2, 3, etc.) |
| `orders[].binance_order_id` | integer | Binance order ID |
| `orders[].price` | number | Entry price |
| `orders[].quantity` | number | Entry quantity |
| `orders[].type` | string | Order type: `MARKET` or `LIMIT` |
| `orders[].status` | string | Order status: `NEW`, `FILLED`, `CANCELLED`, etc. |

**Example Requests:**

```bash
# Get all ACTIVE trades
GET /api/trade/bot?status=ACTIVE

# Get all CLOSED trades for BTCUSDT
GET /api/trade/bot?status=CLOSED&symbol=BTCUSDT

# Get trades with confidence >= 80
GET /api/trade/bot?min_confidence=80

# Get BUY trades in last 7 days
GET /api/trade/bot?side=BUY&date_start=2025-03-08&date_end=2025-03-15

# Get multiple symbols with multiple statuses
GET /api/trade/bot?symbol=BTCUSDT&symbol=ETHUSDT&status=ACTIVE&status=CLOSED
```

**Notes:**
- All query parameters are **optional** - omitting a parameter means no filter is applied for that field
- Multiple values for `status[]` and `symbol[]` are supported by repeating the parameter
- Returns **empty array** if no trades match the filters
- Trades sorted by `created_at DESC` (newest first)
- Does **NOT** require bot to be running (can query trades anytime)

---

#### **POST /api/trade/bot/activate**

Activate automated trading bot.

**Request:**
```bash
POST http://localhost:8000/api/trade/bot/activate
Authorization: Bearer <token>
Content-Type: application/json

{
    "strategy_id": 1
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `strategy_id` | integer | No | Strategy ID to use. Jika tidak disediakan, akan menggunakan active strategy |

**Response:**
```json
{
    "code": 200,
    "message": "Trade bot activated successfully",
    "data": {
        "id": 1,
        "is_active": true,
        "active_since": "2025-03-10T08:00:00Z",
        "strategy_id": 1
    }
}
```

---

#### **POST /api/trade/bot/deactivate**

Deactivate automated trading bot.

**Request:**
```bash
POST http://localhost:8000/api/trade/bot/deactivate
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "Trade bot deactivated successfully",
    "data": {
        "id": 1,
        "is_active": false,
        "deactivated_at": "2025-03-10T12:00:00Z"
    }
}
```

---

## 🤖 Trade Response DTOs

This section contains detailed response structures used across Trade endpoints.

### **TradeData**

Represents a single trade record in responses.

**Structure:**

```typescript
interface TradeData {
  id: number
  symbol: string
  interval: string
  side: string
  confidence: number
  total_score: number
  is_aggressive: boolean
  tp_price: number
  sl_price: number
  risk_reward_ratio: number
  avg_entry_price: number
  leverage: number
  capital_used: number
  total_qty: number
  status: string
  description: string
  tp_order_id: number
  sl_order_id: number
  tp_sl_status: string
  exit_price: number
  pnl: number
  pnl_pct: number
  created_at: string
  updated_at: string
  closed_at: string | null
  orders?: OrderInfo[]
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | number | Trade ID |
| `symbol` | string | Trading symbol (e.g., "BTCUSDT") |
| `interval` | string | Timeframe used (e.g., "15m", "1h") |
| `side` | string | Trade direction (BUY/SELL) |
| `confidence` | number | Confidence score (0-100) |
| `total_score` | number | Total signal score (0-1) |
| `is_aggressive` | boolean | Whether aggressive mode was used |
| `tp_price` | number | Take-profit price |
| `sl_price` | number | Stop-loss price |
| `risk_reward_ratio` | number | Risk-reward ratio |
| `avg_entry_price` | number | Average entry price |
| `leverage` | number | Leverage used |
| `capital_used` | number | Capital used in USDT |
| `total_qty` | number | Total quantity purchased |
| `status` | string | Trade status (ACTIVE/COMPLETED/CANCELLED/CLOSED/REJECTED) |
| `description` | string | Trade description |
| `tp_order_id` | number | Take-profit order ID on Binance |
| `sl_order_id` | number | Stop-loss order ID on Binance |
| `tp_sl_status` | string | TP/SL status (ACTIVE/TP_HIT/SL_HIT/CANCELLED) |
| `exit_price` | number | Exit price (if closed) |
| `pnl` | number | Profit/Loss in USDT |
| `pnl_pct` | number | Profit/Loss percentage |
| `created_at` | string | RFC3339 timestamp when trade was created |
| `updated_at` | string | RFC3339 timestamp when trade was last updated |
| `closed_at` | string \| null | RFC3339 timestamp when trade was closed (null if active) |
| `orders` | OrderInfo[] | List of entry orders (optional) |

---

### **TradeDayStat**

Daily trade statistics.

**Structure:**

```typescript
interface TradeDayStat {
  active: number
  count: number
  tp_hits: number
  sl_hits: number
  total_loss: number
  total_profit: number
  consecutive_lossess: number
  pnl: number
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `active` | number | Number of active trades today |
| `count` | number | Total trades count today |
| `tp_hits` | number | Number of take-profit hits today |
| `sl_hits` | number | Number of stop-loss hits today |
| `total_loss` | number | Total loss in USDT today |
| `total_profit` | number | Total profit in USDT today |
| `consecutive_lossess` | number | Consecutive losses count |
| `pnl` | number | Net PnL for today |

---

### **OrderInfo**

Information about a single order.

**Structure:**

```typescript
interface OrderInfo {
  entry_number: number
  binance_order_id: number
  price: number
  quantity: number
  type: string
  status: string
}
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `entry_number` | number | Entry number (1, 2, 3, etc.) |
| `binance_order_id` | number | Binance order ID |
| `price` | number | Order price |
| `quantity` | number | Order quantity |
| `type` | string | Order type: `MARKET` or `LIMIT` |
| `status` | string | Order status: `NEW`, `FILLED`, `CANCELED`, `EXPIRED`, etc. |

---

## 📋 Watchlists

Manage watchlist symbols.

### **GET /api/watchlists**

Get all watchlists.

**Request:**
```bash
GET http://localhost:8000/api/watchlists
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "symbol": "BTCUSDT",
            "is_active": true,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 2,
            "symbol": "ETHUSDT",
            "is_active": true,
            "created_at": "2025-03-08T10:00:00Z"
        },
        {
            "id": 3,
            "symbol": "BNBUSDT",
            "is_active": false,
            "created_at": "2025-03-08T10:00:00Z"
        }
    ]
}
```

---

### **GET /api/watchlists/:id**

Get watchlist by ID.

**Request:**
```bash
GET http://localhost:8000/api/watchlists/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "symbol": "BTCUSDT",
        "is_active": true,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **POST /api/watchlists**

Create new watchlist.

**Request:**
```bash
POST http://localhost:8000/api/watchlists
Authorization: Bearer <token>
Content-Type: application/json

{
    "symbol": "SOLUSDT",
    "is_active": true
}
```

**Request Body:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `symbol` | string | Yes | max: 20 characters |
| `is_active` | boolean | No | Default: true |

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 4,
        "symbol": "SOLUSDT",
        "is_active": true,
        "created_at": "2025-03-08T10:00:00Z"
    }
}
```

---

### **PUT /api/watchlists/:id**

Update watchlist.

**Request:**
```bash
PUT http://localhost:8000/api/watchlists/1
Authorization: Bearer <token>
Content-Type: application/json

{
    "symbol": "BTCUSDT",
    "is_active": false
}
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **DELETE /api/watchlists/:id**

Delete watchlist.

**Request:**
```bash
DELETE http://localhost:8000/api/watchlists/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

## 🎯 Strategies

Manage trading strategies with multi-timeframe analysis, indicator weights, and money management.

### **GET /api/strategies**

Get all strategies.

**Request:**
```bash
GET http://localhost:8000/api/strategies
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "strategy_name": "Day Trading Pro",
            "primary_tf": "15m",
            "is_active": true,
            "created_at": "2025-03-08T10:00:00Z",
            "updated_at": "2025-03-08T10:00:00Z",
            "timeframes": [
                {
                    "id": 1,
                    "tf": "15m",
                    "weight": 0.50,
                    "timeframe_detail": {
                        "name": "15m",
                        "in_minutes": 15,
                        "created_at": "2025-03-08T10:00:00Z"
                    }
                },
                {
                    "id": 2,
                    "tf": "30m",
                    "weight": 0.30,
                    "timeframe_detail": {
                        "name": "30m",
                        "in_minutes": 30,
                        "created_at": "2025-03-08T10:00:00Z"
                    }
                }
            ],
            "indicator_weights": [
                {
                    "id": 1,
                    "indicator_id": 1,
                    "weight": 0.30,
                    "indicator_detail": {
                        "id": 1,
                        "name": "Moving Average",
                        "indicator": "moving_average",
                        "description": "Moving Average - Trend indicator",
                        "is_active": true,
                        "weight": 0.30,
                        "order_view": 1,
                        "created_at": "2025-03-08T10:00:00Z"
                    }
                }
            ],
            "money_management": {
                "min_confidence": 45,
                "max_daily_trades": 10,
                "max_daily_loss_percent": 0.05,
                "max_position_size": 0.15,
                "risk_reward_ratio": 1.5,
                "leverage": 5,
                "is_agressive": false,
                "order_expiration_hours": 4
            }
        }
    ]
}
```

**Note:** `money_management` fields with value 0 are hidden (using `omitempty`).

---

### **GET /api/strategies/active**

Get the currently active strategy (only 1 strategy can be active at a time).

**Request:**
```bash
GET http://localhost:8000/api/strategies/active
Authorization: Bearer <token>
```

**Response (if active strategy exists):**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "strategy_name": "Day Trading Pro",
        "primary_tf": "15m",
        "is_active": true,
        ...
    }
}
```

**Response (if no active strategy):**
```json
{
    "code": 400,
    "message": "data not found",
    "error": null
}
```

---

### **GET /api/strategies/:id**

Get strategy by ID.

**Request:**
```bash
GET http://localhost:8000/api/strategies/1
Authorization: Bearer <token>
```

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **POST /api/strategies**

Create new strategy.

**Request:**
```bash
POST http://localhost:8000/api/strategies
Authorization: Bearer <token>
Content-Type: application/json

{
    "strategy_name": "Scalping Master",
    "primary_tf": "5m",
    "is_active": true,
    "timeframes": [
        {"tf": "1m", "weight": 0.10},
        {"tf": "5m", "weight": 0.60},
        {"tf": "15m", "weight": 0.30}
    ],
    "indicator_weights": [
        {"indicator_id": 1, "weight": 0.25},
        {"indicator_id": 2, "weight": 0.25},
        {"indicator_id": 3, "weight": 0.20},
        {"indicator_id": 4, "weight": 0.15},
        {"indicator_id": 5, "weight": 0.15}
    ],
    "money_management": [
        {"parameter": "risk_per_trade", "value": "0.01"},
        {"parameter": "max_drawdown", "value": "0.10"},
        {"parameter": "position_size", "value": "calculated"}
    ]
}
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `strategy_name` | string | Yes | Unique name (max: 100) |
| `primary_tf` | string | Yes | Primary timeframe (e.g., "5m") |
| `is_active` | boolean | No | Default: true |
| `timeframes` | array | No | List of timeframes with weights |
| `indicator_weights` | array | No | List of indicators with weights |
| `money_management` | array | No | List of MM parameters |

**Timeframe Object:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tf` | string | Yes | Timeframe name (max: 5) |
| `weight` | float | Yes | 0 ≤ weight ≤ 1 |

**Indicator Weight Object:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `indicator_id` | uint | Yes | Reference to m_indicator |
| `weight` | float | Yes | 0 ≤ weight ≤ 1 |

**Money Management Object:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `parameter` | string | Yes | Config key (e.g., "MIN_CONFIDENCE") |
| `value` | string | Yes | Parameter value |

**Important:** If `is_active: true`, all other strategies will be automatically deactivated.

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **PUT /api/strategies/:id**

Update strategy.

**Request:**
```bash
PUT http://localhost:8000/api/strategies/1
Authorization: Bearer <token>
Content-Type: application/json

{
    "strategy_name": "Day Trading Pro Updated",
    "primary_tf": "15m",
    "is_active": true,
    "timeframes": [
        {"tf": "15m", "weight": 0.60},
        {"tf": "1h", "weight": 0.40}
    ],
    "indicator_weights": [
        {"indicator_id": 1, "weight": 0.30},
        {"indicator_id": 2, "weight": 0.30}
    ],
    "money_management": [
        {"parameter": "MIN_CONFIDENCE", "value": "50"},
        {"parameter": "LEVERAGE", "value": "10"}
    ]
}
```

**Important:** 
- If `is_active: true`, all other strategies will be deactivated first
- All existing relationships (timeframes, indicators, money_management) will be replaced

**Response:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

---

### **DELETE /api/strategies/:id**

Delete strategy.

**Request:**
```bash
DELETE http://localhost:8000/api/strategies/1
Authorization: Bearer <token>
```

**Important:** Cannot delete an active strategy. Must deactivate first.

**Response (success):**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

**Response (error - active strategy):**
```json
{
    "code": 400,
    "message": "cannot delete active strategy. Please deactivate it first",
    "error": null
}
```

---

## 📊 Money Management Parameters

Supported parameters for strategy money management:

| Parameter | Type | Description |
|-----------|------|-------------|
| `MIN_CONFIDENCE` | int8 | Minimum confidence level (0-100) |
| `MAX_DAILY_TRADES` | int8 | Maximum trades per day |
| `MAX_DAILY_LOSS_PERCENT` | float32 | Maximum daily loss % from balance |
| `MAX_DAILY_LOSS_COUNT` | int8 | Maximum consecutive losses |
| `RISK_REWARD_RATIO` | float32 | Minimum R:R ratio |
| `RISK_REWARD_TARGET` | float32 | Target R:R ratio |
| `MAX_POSITION_SIZE` | float32 | Maximum position size % of balance |
| `MAX_RISK_PER_TRADE` | float32 | Maximum risk % per trade |
| `LEVERAGE` | int8 | Leverage to use |
| `IS_AGRESSIVE` | bool | Aggressive mode (true/false) |
| `ORDER_EXPIRATION_HOURS` | int8 | Order expiration in hours |

---

## 📝 Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request (validation error, business logic error) |
| 401 | Unauthorized (missing/invalid token) |
| 404 | Not Found |
| 500 | Internal Server Error |

---

## 🔧 Migration & Seeding

To setup the database with default data:

```bash
cd be

# Drop all tables
go run cmd/migration/main.go down

# Create all tables
go run cmd/migration/main.go up

# Seed default data
go run cmd/migration/main.go seed
```

This will create:
- Default thresholds (STRONG_BUY, BUY, WAIT, SELL, STRONG_SELL)
- Default timeframes (1m, 3m, 5m, 15m, 30m, 1h, 4h, 1d, 1w, 1M)
- Default indicators (Moving Average, MACD, RSI, Stochastic, Bollinger Bands, etc.)
- Default configs (Money Management, Binance settings)
- Sample strategy ("Day Trading Pro")

---

*Last Updated: March 8, 2026*
