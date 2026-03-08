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
