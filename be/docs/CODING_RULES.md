# Coding Rules & Standards

## 📋 Table of Contents

1. [Project Structure](#project-structure)
2. [Architecture Pattern](#architecture-pattern)
3. [Naming Conventions](#naming-conventions)
4. [Code Style](#code-style)
5. [Database Conventions](#database-conventions)
6. [API Standards](#api-standards)
7. [Error Handling](#error-handling)
8. [Testing Guidelines](#testing-guidelines)
9. [External Clients](#external-clients)

---

## 🏗️ Project Structure

```
be/
├── cmd/
│   └── main/                    # Application entry point
├── internal/
│   ├── clients/                 # External API clients (Binance, S3, etc.)
│   │   ├── binance/             # Binance client
│   │   ├── s3/                  # S3 storage client
│   │   └── coinmarketcap/       # CoinMarketCap client
│   ├── config/                  # Configuration management
│   ├── controller/              # HTTP request handlers
│   ├── database/                # Database connections & migrations
│   ├── di/                      # Dependency Injection container
│   ├── dtos/                    # Data Transfer Objects
│   ├── helpers/                 # Utility functions (simple)
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Domain models / entities
│   ├── repository/              # Data access layer
│   ├── routes/                  # Route definitions
│   └── service/                 # Business logic layer
├── docs/                        # Documentation
├── .env                         # Environment variables
├── .env.example                 # Environment template
└── go.mod                       # Go module definition
```

---

## 🏛️ Architecture Pattern

Project ini menggunakan **Clean Architecture** dengan 4 layer utama:

### **Layer Flow**
```
Routes → Controller → Service → Repository → Database
         (HTTP)      (Logic)    (Data)       (Storage)
```

### **Responsibilities**

| Layer | Responsibility | Should NOT |
|-------|---------------|------------|
| **Controller** | Handle HTTP requests, validation, response formatting | Business logic, DB queries |
| **Service** | Business logic, transactions, DTO transformation | HTTP specifics, direct DB access |
| **Repository** | Data access, CRUD operations | Business logic, HTTP specifics |
| **Models** | Domain entities, table mapping | Business logic, DB queries |

---

## ⚠️ Critical Rules (MUST FOLLOW)

### **1. 🚫 JANGAN PERNAH Memodifikasi File Core**

File-file berikut adalah **CORE FRAMEWORK** dan TIDAK BOLEH diubah:

| File | Status | Alasan |
|------|--------|--------|
| `repository/00_generic.go` | ❌ **IMMUTABLE** | Generic CRUD foundation |
| `repository/00_transaction.go` | ❌ **IMMUTABLE** | Transaction management core |

> **Pelanggaran = CRITICAL ERROR** 🚨
> 
> File ini adalah foundation dari seluruh architecture. Perubahan akan berdampak ke semua module.

---

### **2. ⚡ Minimalisir Perubahan di `00_repository.go`**

Jika menambah repository baru:

```go
// ✅ GOOD: Tambah 1 baris di NewRepositories
func NewRepositories(db *gorm.DB) (*Repositories, error) {
    // ... existing code
    YourEntityRepo := NewYourEntityRepository(db)  // ← Tambah ini
    
    return &Repositories{
        // ... existing
        YourEntity: YourEntityRepo,  // ← Tambah ini
    }, nil
}

// ❌ BAD: Mengubah struktur Repositories tanpa keperluan
```

**Aturan:**
- ✅ BOLEH: Menambah field di struct `Repositories`
- ✅ BOLEH: Menambah initialization di `NewRepositories()`
- ❌ JANGAN: Mengubah existing fields atau logic

---

### **3. 🎯 SELALU Gunakan Generic Function untuk CRUD di Repository**

```go
// ✅ GOOD: Extend GenericRepository
type WatchlistRepository struct {
    *GenericRepository[models.Watchlist]  // ← Embed generic
}

func NewWatchlistRepository(db *gorm.DB) *WatchlistRepository {
    return &WatchlistRepository{
        GenericRepository: NewGenericRepository(db, &models.Watchlist{}),
    }
}

// ✅ GOOD: Tambah custom method HANYA jika perlu
func (r *WatchlistRepository) FindAllActive(tx *gorm.DB) ([]models.Watchlist, error) {
    db := r.getDB(tx)
    var watchlists []models.Watchlist
    if err := db.Model(&models.Watchlist{}).Where("is_active = ?", true).Find(&watchlists).Error; err != nil {
        return nil, err
    }
    return watchlists, nil
}

// ❌ BAD: Implementasi ulang CRUD yang sudah ada di GenericRepository
func (r *WatchlistRepository) FindByID(tx *gorm.DB, id uint) (*models.Watchlist, error) {
    // JANGAN lakukan ini! GenericRepository sudah punya method ini
}
```

**CRUD yang SUDAH TERSEDIA di GenericRepository:**
- `FindByID(tx, id)` → Find by ID
- `FindAll(tx)` → Find all records
- `FindByField(tx, filter)` → Find by filter
- `Create(tx, request)` → Create new record
- `CreateMany(tx, request)` → Bulk create
- `Update(tx, filter, update)` → Update by filter
- `Delete(tx, id)` → Delete by ID
- `Count(tx)` → Count records

---

### **4. 🔒 SELALU Gunakan Transaction di Service untuk Data Modification**

**Pattern Utama: Using WithinTransactionWithResult (WAJIB)**

```go
// ✅ GOOD: Using WithinTransactionWithResult helper
func (s *Services) WatchlistCreate(ctx *gin.Context, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error) {
    result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
        watchlist := &models.Watchlist{
            Symbol:   req.Symbol,
            IsActive: req.IsActive,
        }

        watchlist, err = s.repo.Watchlist.Create(tx, watchlist)  // ← Pass tx
        if err != nil {
            return nil, err
        }

        // Bisa operasi lain di sini dengan tx yang sama
        // Contoh: s.repo.OtherEntity.Create(tx, ...)

        return watchlist, nil
    })

    if err != nil {
        return nil, err
    }

    watchlist := result.(*models.Watchlist)
    return &dtos.WatchlistData{
        ID:        watchlist.ID,
        Symbol:    watchlist.Symbol,
        IsActive:  watchlist.IsActive,
        CreatedAt: helpers.FormatDateTime(watchlist.CreatedAt),
    }, nil
}
```

**✅ GOOD: Read operation TIDAK perlu transaction**
```go
func (s *Services) WatchlistGetAll(ctx *gin.Context) (res []dtos.WatchlistData, err error) {
    watchlists, err := s.repo.Watchlist.FindAll(nil)  // ← nil = no transaction needed
    if err != nil {
        return nil, err
    }

    for _, wl := range watchlists {
        res = append(res, dtos.WatchlistData{
            ID:        wl.ID,
            Symbol:    wl.Symbol,
            IsActive:  wl.IsActive,
            CreatedAt: helpers.FormatDateTime(wl.CreatedAt),
        })
    }

    return
}
```

**❌ BAD: Tidak menggunakan transaction untuk write operation**
```go
func (s *Services) WatchlistCreate(ctx *gin.Context, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error) {
    watchlist, err := s.repo.Watchlist.Create(nil, watchlist)  // ← nil = no transaction!
    // ...
}
```

**Kapan WAJIB Transaction:**
- ✅ Create data
- ✅ Update data
- ✅ Delete data
- ✅ Multiple operations (bulk)
- ⚠️ Read-only: BOLEH tanpa transaction (`nil`)

**Benefits Menggunakan `WithinTransactionWithResult`:**
- ✅ Auto rollback on error
- ✅ Auto rollback on panic
- ✅ Auto commit on success
- ✅ Cleaner code (no manual tx management)
- ✅ Consistent pattern across codebase

**Contoh Multiple Operations:**
```go
func (s *Services) SignalCreateWithEntries(ctx *gin.Context, req *dtos.SignalRequest) (res *dtos.SignalData, err error) {
    result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
        // 1. Create parent signal
        signal := &models.Signal{
            Symbol: req.Symbol,
            Side:   req.Side,
            // ...
        }
        signal, err = s.repo.Signal.Create(tx, signal)
        if err != nil {
            return nil, err
        }

        // 2. Create multiple entries
        for _, entryReq := range req.Entries {
            entry := &models.SignalEntry{
                SignalID:   signal.ID,
                EntryNumber: entryReq.Number,
                // ...
            }
            _, err = s.repo.SignalEntry.Create(tx, entry)
            if err != nil {
                return nil, err
            }
        }

        return signal, nil
    })

    if err != nil {
        return nil, err
    }

    // Convert to DTO
    signal := result.(*models.Signal)
    // ... return DTO
}
```

---

### **5. 📦 Service & Controller Menggunakan Single Struct**

**JANGAN gunakan multiple structs dalam 1 file:**

```go
// ✅ GOOD: Single struct per file
// File: watchlist_service.go
package service

func (s *Services) WatchlistGetAll(ctx *gin.Context) (res []dtos.WatchlistData, err error) {
    // implementation
}

func (s *Services) WatchlistGetByID(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error) {
    // implementation
}

// File: watchlist_controller.go
package controller

func (c *Controller) WatchlistIndex(ctx *gin.Context) {
    // implementation
}

func (c *Controller) WatchlistDetail(ctx *gin.Context) {
    // implementation
}

// ❌ BAD: Multiple structs dalam 1 file
type WatchlistService struct {
    // fields
}

type WatchlistHelper struct {
    // fields
}
```

**Aturan:**
- ✅ 1 file = 1 module = multiple functions
- ✅ Semua functions sebagai method dari `*Services` atau `*Controller`
- ❌ JANGAN buat struct baru di dalam file service/controller

---

### **6. 🏷️ Penamaan Function dengan Prefix Module**

**Format:** `<Module><Action><Entity>`

```go
// ✅ GOOD: Prefix module (Watchlist)
func (s *Services) WatchlistGetAll(ctx *gin.Context) (res []dtos.WatchlistData, err error)
func (s *Services) WatchlistGetByID(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error)
func (s *Services) WatchlistCreate(ctx *gin.Context, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error)
func (s *Services) WatchlistUpdate(ctx *gin.Context, id uint, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error)
func (s *Services) WatchlistDelete(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error)

// Controller juga sama
func (c *Controller) WatchlistIndex(ctx *gin.Context)
func (c *Controller) WatchlistDetail(ctx *gin.Context)
func (c *Controller) WatchlistCreate(ctx *gin.Context)
func (c *Controller) WatchlistUpdate(ctx *gin.Context)
func (c *Controller) WatchlistDelete(ctx *gin.Context)

// ❌ BAD: Tanpa prefix module
func (s *Services) GetAll(ctx *gin.Context)  // ← Tidak jelas module mana
func (s *Services) Create(ctx *gin.Context)  // ← Conflicting names
```

**Pattern Naming:**

| Layer | Pattern | Contoh |
|-------|---------|--------|
| Service (Read) | `<Module>GetAll`, `<Module>GetByID` | `WatchlistGetAll` |
| Service (Write) | `<Module>Create`, `<Module>Update`, `<Module>Delete` | `WatchlistCreate` |
| Controller (Index) | `<Module>Index` | `WatchlistIndex` |
| Controller (Detail) | `<Module>Detail` | `WatchlistDetail` |
| Controller (Action) | `<Module>Create`, `<Module>Update`, `<Module>Delete` | `WatchlistCreate` |

---

## 📝 Naming Conventions

### **Files & Directories**

| Type | Convention | Example |
|------|-----------|---------|
| Directory | `snake_case` | `internal/models/` |
| Model File | `<entity>.go` | `watchlist.go` |
| Repository File | `<entity>_repo.go` | `watchlist_repo.go` |
| Service File | `<entity>_service.go` | `watchlist_service.go` |
| Controller File | `<entity>_controller.go` | `watchlist_controller.go` |
| Route File | `<entity>_routes.go` | `watchlist_routes.go` |
| DTO File | `<entity>_dto.go` | `watchlist_dto.go` |

### **Go Code**

| Element | Convention | Example |
|---------|-----------|---------|
| Package | `snake_case` | `package models` |
| Struct | `PascalCase` | `type Watchlist struct` |
| Function | `PascalCase` | `func WatchlistGetAll()` |
| Variable | `camelCase` | `var watchlistData` |
| Constant | `PascalCase` | `const MaxRetry` |
| Interface | `PascalCase` | `type Repository interface` |
| Method | `PascalCase` | `func (r *Repo) FindByID()` |

### **Database**

| Element | Convention | Example |
|---------|-----------|---------|
| Table | `snake_case` | `watchlist`, `m_indicator` |
| Column | `snake_case` | `is_active`, `created_at` |
| Primary Key | `id` | `id` |
| Foreign Key | `<table>_id` | `watchlist_id` |
| Join Table | `<table1>_<table2>` | `user_roles` |

### **JSON Fields**

```go
type WatchlistData struct {
    ID        uint   `json:"id"`         // lowercase
    Symbol    string `json:"symbol"`     // lowercase
    IsActive  bool   `json:"is_active"`  // snake_case
    CreatedAt string `json:"created_at"` // snake_case
}
```

---

## 💻 Code Style

### **1. Import Order**

```go
package controller

import (
    // Standard library
    "strconv"

    // Third-party
    "github.com/gin-gonic/gin"

    // Internal
    "github.com/reshap/trading-bot/internal/dtos"
    "github.com/reshap/trading-bot/internal/helpers"
)
```

### **2. Function Structure**

```go
// ✅ Good: Clear error handling
func (c *Controller) WatchlistGetByID(ctx *gin.Context, id uint) (res *dtos.WatchlistData, err error) {
    watchlist, err := s.repo.Watchlist.FindByID(nil, id)
    if err != nil {
        return nil, err
    }

    res = &dtos.WatchlistData{
        ID:        watchlist.ID,
        Symbol:    watchlist.Symbol,
        IsActive:  watchlist.IsActive,
        CreatedAt: helpers.FormatDateTime(watchlist.CreatedAt),
    }

    return
}
```

### **3. Struct Definition**

```go
type Watchlist struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Symbol    string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"symbol"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
```

### **4. Error Handling**

```go
// ✅ Good: Proper error propagation
if err != nil {
    helpers.RespondError(ctx, err)
    return
}

// ✅ Good: Custom errors
var (
    ErrNotFound = errors.New("data not found")
    ErrInternal = errors.New("internal server error")
)
```

---

## 🗄️ Database Conventions

### **1. Model Definition**

```go
package models

import (
    "time"
)

type Watchlist struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Symbol    string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"symbol"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Watchlist) TableName() string {
    return "watchlist"
}
```

### **2. GORM Tags**

| Tag | Purpose | Example |
|-----|---------|---------|
| `primaryKey` | Primary key | `gorm:"primaryKey"` |
| `autoIncrement` | Auto-increment | `gorm:"autoIncrement"` |
| `type` | Column type | `gorm:"type:varchar(20)"` |
| `uniqueIndex` | Unique constraint | `gorm:"uniqueIndex"` |
| `not null` | NOT NULL constraint | `gorm:"not null"` |
| `default` | Default value | `gorm:"default:true"` |
| `column` | Custom column name | `gorm:"column:tp_price"` |
| `foreignKey` | Foreign key relation | `gorm:"foreignKey:SignalID"` |

### **3. Master Table Prefix**

Untuk master data, gunakan prefix `m_`:

```go
func (Indicators) TableName() string {
    return "m_indicator"  // ✅ Master table
}

func (Threshold) TableName() string {
    return "m_threshold"  // ✅ Master table
}
```

---

## 🌐 API Standards

### **1. Route Pattern**

```go
// RegisterWatchlistRoutes registers all watchlist-related routes
func RegisterWatchlistRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
    watchlistGroup := router.Group("/watchlists")
    {
        watchlistGroup.GET("", ctrl.WatchlistIndex)
        watchlistGroup.GET("/:id", ctrl.WatchlistDetail)
        watchlistGroup.POST("", ctrl.WatchlistCreate)
        watchlistGroup.PUT("/:id", ctrl.WatchlistUpdate)
        watchlistGroup.DELETE("/:id", ctrl.WatchlistDelete)
    }
}
```

### **2. Endpoint Naming**

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/watchlists` | Get all |
| `GET` | `/api/watchlists/:id` | Get by ID |
| `POST` | `/api/watchlists` | Create |
| `PUT` | `/api/watchlists/:id` | Update |
| `DELETE` | `/api/watchlists/:id` | Delete |

### **3. Response Format**

**Success Response:**
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

**Error Response:**
```json
{
    "code": 400,
    "message": "Invalid symbol format",
    "error": null
}
```

### **4. Controller Pattern**

```go
func (c *Controller) WatchlistCreate(ctx *gin.Context) {
    var req dtos.WatchlistRequest

    // 1. Validate input
    if err := ctx.ShouldBindJSON(&req); err != nil {
        helpers.ResponseJsonNotValid(ctx)
        return
    }

    // 2. Call service
    watchlist, err := c.srvc.WatchlistCreate(ctx, &req)
    if err != nil {
        helpers.RespondError(ctx, err)
        return
    }

    // 3. Return response
    helpers.ResponsedWithData(ctx, 200, "success", watchlist)
}
```

---

## ⚠️ Error Handling

### **1. Helper Functions**

```go
// Success responses
helpers.ResponsedWithData(ctx, 200, "success", data)
helpers.RespondWithMessage(ctx, 200, "Operation successful")
helpers.ReponseSuccess(ctx)

// Error responses
helpers.RespondError(ctx, err)
helpers.ResponseJsonNotValid(ctx)
```

### **2. Error Types**

```go
var (
    ErrNotFound = errors.New("data not found")
    ErrInternal = errors.New("internal server error")
)

func RespondError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        RespondWithMessage(c, http.StatusNotFound, err.Error())
    case errors.Is(err, ErrInternal):
        RespondWithMessage(c, http.StatusInternalServerError, err.Error())
    default:
        RespondWithMessage(c, 400, err.Error())
    }
}
```

---

## 🧪 Testing Guidelines

### **1. Test File Naming**

```
<entity>_test.go
```

Example: `watchlist_service_test.go`

### **2. Test Structure**

```go
func TestWatchlistGetAll(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### **3. Table-Driven Tests**

```go
func TestWatchlistCreate(t *testing.T) {
    tests := []struct {
        name    string
        req     *dtos.WatchlistRequest
        wantErr bool
    }{
        {
            name: "success_create_watchlist",
            req:  &dtos.WatchlistRequest{Symbol: "BTCUSDT", IsActive: true},
            wantErr: false,
        },
        {
            name: "fail_empty_symbol",
            req:  &dtos.WatchlistRequest{Symbol: "", IsActive: true},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

---

## 🔐 Security Best Practices

1. **Never commit `.env`** - Use `.env.example` as template
2. **Use middleware auth** - Protected routes require authentication
3. **Validate all inputs** - Use `binding` tags in DTOs
4. **Sanitize user input** - Prevent SQL injection (GORM handles this)
5. **Use parameterized queries** - Never concatenate SQL strings

---

## 📦 Dependency Injection

```go
// container.go
func NewContainer(cfg *config.Config) (*Container, error) {
    // 1. Initialize MySQL
    db, err := database.NewMySQLConnection(&cfg.DB)

    // 2. Initialize Redis
    redisClient, err := database.NewRedisClient(&cfg.Redis)

    // 3. Create CacheClient
    cacheClient := database.NewCacheClient(redisClient)

    // 4. Initialize repositories
    repo, _ := repository.NewRepositories(db)

    // 5. Create services
    srvc, _ := service.NewServices(repo, cfg, cacheClient)

    // 6. Initialize controller
    ctrl := controller.NewController(srvc, cfg)

    return &Container{Cfg: cfg, DB: db, Repo: repo, Srvc: srvc, Ctrl: ctrl}, nil
}
```

---

## 🚀 Quick Start: Adding New Feature

### **1. Create Model** (`models/<entity>.go`)
```go
type YourEntity struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"type:varchar(100);not null" json:"name"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (YourEntity) TableName() string {
    return "your_table"
}
```

### **2. Create DTO** (`dtos/<entity>_dto.go`)
```go
type YourEntityRequest struct {
    Name string `json:"name" binding:"required"`
}

type YourEntityData struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}
```

### **3. Create Repository** (`repository/<entity>_repo.go`)
```go
type YourEntityRepository struct {
    *GenericRepository[models.YourEntity]
}

func NewYourEntityRepository(db *gorm.DB) *YourEntityRepository {
    return &YourEntityRepository{
        GenericRepository: NewGenericRepository(db, &models.YourEntity{}),
    }
}
```

### **4. Register Repository** (`repository/00_repository.go`)
```go
func NewRepositories(db *gorm.DB) (*Repositories, error) {
    // ... existing code
    YourEntityRepo := NewYourEntityRepository(db)
    
    return &Repositories{
        // ... existing
        YourEntity: YourEntityRepo,
    }, nil
}
```

### **5. Create Service** (`service/<entity>_service.go`)
```go
// READ (tanpa transaction)
func (s *Services) YourEntityGetAll(ctx *gin.Context) (res []dtos.YourEntityData, err error) {
    entities, err := s.repo.YourEntity.FindAll(nil)  // ← nil = no transaction
    if err != nil {
        return nil, err
    }

    for _, e := range entities {
        res = append(res, dtos.YourEntityData{
            ID:        e.ID,
            Name:      e.Name,
            CreatedAt: helpers.FormatDateTime(e.CreatedAt),
        })
    }

    return
}

// WRITE (WAJIB menggunakan WithinTransactionWithResult)
func (s *Services) YourEntityCreate(ctx *gin.Context, req *dtos.YourEntityRequest) (res *dtos.YourEntityData, err error) {
    result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
        entity := &models.YourEntity{
            Name: req.Name,
        }

        entity, err = s.repo.YourEntity.Create(tx, entity)  // ← Pass tx
        if err != nil {
            return nil, err
        }

        return entity, nil
    })

    if err != nil {
        return nil, err
    }

    entity := result.(*models.YourEntity)
    return &dtos.YourEntityData{
        ID:        entity.ID,
        Name:      entity.Name,
        CreatedAt: helpers.FormatDateTime(entity.CreatedAt),
    }, nil
}
```

### **6. Create Controller** (`controller/<entity>_controller.go`)
```go
func (c *Controller) YourEntityIndex(ctx *gin.Context) {
    entities, err := c.srvc.YourEntityGetAll(ctx)
    if err != nil {
        helpers.RespondError(ctx, err)
        return
    }

    helpers.ResponsedWithData(ctx, 200, "success", entities)
}
```

### **7. Create Routes** (`routes/<entity>_routes.go`)
```go
func RegisterYourEntityRoutes(router *gin.RouterGroup, ctrl *controller.Controller) {
    group := router.Group("/your-entities")
    {
        group.GET("", ctrl.YourEntityIndex)
        group.GET("/:id", ctrl.YourEntityDetail)
        group.POST("", ctrl.YourEntityCreate)
        group.PUT("/:id", ctrl.YourEntityUpdate)
        group.DELETE("/:id", ctrl.YourEntityDelete)
    }
}
```

### **8. Register Routes** (`cmd/main/main.go`)
```go
protected := apiGroup.Group("")
protected.Use(middleware.AuthMiddleware(engine.Srvc))
{
    // ... existing routes
    routes.RegisterYourEntityRoutes(protected, engine.Ctrl)
}
```

---

## 🔌 External Clients

External clients adalah module untuk integrasi dengan **third-party services** atau **external APIs** (e.g., Binance, AWS S3, CoinMarketCap).

### **Kapan Menggunakan Clients?**

✅ **Gunakan Clients untuk:**
- Integrasi dengan external APIs (Binance, CoinMarketCap)
- Cloud services (AWS S3, Google Cloud Storage)
- Payment gateways (Stripe, PayPal)
- Email services (SendGrid, AWS SES)
- SMS services (Twilio)

❌ **JANGAN gunakan untuk:**
- Internal business logic (gunakan `service/`)
- Simple utility functions (gunakan `helpers/`)
- Database operations (gunakan `repository/`)

---

### **Folder Structure**

**Location:** `internal/clients/<service_name>/`

```
internal/
└── clients/
    └── binance/
        ├── client.go            # Main client struct + constructor
        ├── service.go           # Business logic methods
        ├── dto.go               # Request/Response DTOs
        ├── config.go            # Configuration struct (optional)
        └── helper.go            # Helper functions (optional)
```

---

### **File Conventions**

#### **1. client.go - Main Client**

```go
package binance

import (
    "context"
    "github.com/binance/binance-connector-go"
    "github.com/reshap/trading-bot/internal/config"
)

// Client represents Binance API client
type Client struct {
    apiClient *binance.Client
    config    *Config
}

// Config holds Binance configuration
type Config struct {
    APIKey    string
    SecretKey string
    IsTestnet bool
    BaseURL   string
}

// NewClient creates new Binance client
func NewClient(cfg *config.BinanceConfig) *Client {
    client := binance.NewClient(cfg.APIKey, cfg.SecretKey)
    
    if cfg.IsTestnet {
        client.BaseURL = "https://testnet.binancefuture.com"
    }
    
    return &Client{
        apiClient: client,
        config:    &Config{
            APIKey:    cfg.APIKey,
            SecretKey: cfg.SecretKey,
            IsTestnet: cfg.IsTestnet,
        },
    }
}

// Close closes the client connection (if needed)
func (c *Client) Close() error {
    // Cleanup resources if needed
    return nil
}
```

---

#### **2. service.go - Business Logic Methods**

```go
package binance

import (
    "context"
    "fmt"
)

// GetPrice gets current price for a symbol
func (c *Client) GetPrice(ctx context.Context, symbol string) (float64, error) {
    resp, err := c.apiClient.NewKlinesService().
        Symbol(symbol).
        Interval("1m").
        Limit(1).
        Do(ctx)
    
    if err != nil {
        return 0, fmt.Errorf("failed to get price: %w", err)
    }
    
    if len(resp) == 0 {
        return 0, fmt.Errorf("no data found for symbol: %s", symbol)
    }
    
    return parsePrice(resp[0]), nil
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*OrderResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Place order via API
    resp, err := c.apiClient.NewCreateOrderService().
        Symbol(req.Symbol).
        Side(req.Side).
        Type(req.Type).
        Quantity(fmt.Sprintf("%f", req.Quantity)).
        Do(ctx)
    
    if err != nil {
        return nil, fmt.Errorf("failed to place order: %w", err)
    }
    
    return &OrderResponse{
        OrderID:     resp.OrderID,
        Symbol:      resp.Symbol,
        Status:      resp.Status,
        FilledPrice: resp.AveragePrice,
        FilledQty:   resp.FilledQty,
    }, nil
}

// GetAccountInfo gets account information
func (c *Client) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
    resp, err := c.apiClient.NewGetAccountService().Do(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get account info: %w", err)
    }
    
    return mapAccountInfo(resp), nil
}
```

**Best Practices:**
- ✅ Selalu gunakan `context.Context` untuk timeout/cancellation
- ✅ Wrap error dengan `fmt.Errorf("...: %w", err)` untuk error tracing
- ✅ Validate input sebelum call API
- ✅ Return custom DTOs, bukan response mentah dari API

---

#### **3. dto.go - Request/Response DTOs**

```go
package binance

// PlaceOrderRequest represents request to place order
type PlaceOrderRequest struct {
    Symbol      string  `json:"symbol" binding:"required"`
    Side        string  `json:"side" binding:"required,oneof=BUY SELL"`
    Type        string  `json:"type" binding:"required"`
    Quantity    float64 `json:"quantity" binding:"required,gt=0"`
    Price       float64 `json:"price"`
    TimeInForce string  `json:"time_in_force"`
}

// Validate validates the request
func (r *PlaceOrderRequest) Validate() error {
    if r.Symbol == "" {
        return fmt.Errorf("symbol is required")
    }
    if r.Quantity <= 0 {
        return fmt.Errorf("quantity must be greater than 0")
    }
    return nil
}

// OrderResponse represents order response
type OrderResponse struct {
    OrderID     int64   `json:"order_id"`
    Symbol      string  `json:"symbol"`
    Status      string  `json:"status"`
    FilledPrice float64 `json:"filled_price"`
    FilledQty   float64 `json:"filled_qty"`
    CreatedAt   int64   `json:"created_at"`
}

// AccountInfo represents account information
type AccountInfo struct {
    MakerCommission  int             `json:"maker_commission"`
    TakerCommission  int             `json:"taker_commission"`
    Balances         []BalanceInfo   `json:"balances"`
    CanTrade         bool            `json:"can_trade"`
    CanDeposit       bool            `json:"can_deposit"`
    CanWithdraw      bool            `json:"can_withdraw"`
}

// BalanceInfo represents balance information
type BalanceInfo struct {
    Asset  string  `json:"asset"`
    Free   float64 `json:"free"`
    Locked float64 `json:"locked"`
}
```

---

#### **4. config.go - Configuration (Optional)**

```go
package binance

// Config holds Binance configuration
type Config struct {
    APIKey        string
    SecretKey     string
    IsTestnet     bool
    BaseURL       string
    Timeout       time.Duration
    MaxRetries    int
    RetryDelay    time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
    return &Config{
        IsTestnet:  true,
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        RetryDelay: 1 * time.Second,
    }
}

// Validate validates configuration
func (c *Config) Validate() error {
    if c.APIKey == "" {
        return fmt.Errorf("API key is required")
    }
    if c.SecretKey == "" {
        return fmt.Errorf("secret key is required")
    }
    return nil
}
```

---

#### **5. helper.go - Helper Functions (Optional)**

```go
package binance

// parsePrice parses price from kline response
func parsePrice(kline *binance.Kline) float64 {
    price, _ := strconv.ParseFloat(kline.Close, 64)
    return price
}

// mapAccountInfo maps API response to AccountInfo
func mapAccountInfo(resp *binance.Account) *AccountInfo {
    balances := make([]BalanceInfo, len(resp.Balances))
    for i, b := range resp.Balances {
        balances[i] = BalanceInfo{
            Asset:  b.Asset,
            Free:   parseFloat(b.Free),
            Locked: parseFloat(b.Locked),
        }
    }
    
    return &AccountInfo{
        MakerCommission: resp.MakerCommission,
        TakerCommission: resp.TakerCommission,
        Balances:        balances,
        CanTrade:        resp.CanTrade,
        CanDeposit:      resp.CanDeposit,
        CanWithdraw:     resp.CanWithdraw,
    }
}

// parseFloat parses float string
func parseFloat(s string) float64 {
    val, _ := strconv.ParseFloat(s, 64)
    return val
}
```

---

### **Dependency Injection**

Register clients in `internal/di/container.go`:

```go
type Container struct {
    Cfg           *config.Config
    DB            *gorm.DB
    Redis         *database.CacheClient
    Repo          *repository.Repositories
    Srvc          *service.Services
    Ctrl          *controller.Controller
    BinanceClient *binance.Client  // ← Add client
}

func NewContainer(cfg *config.Config) (*Container, error) {
    // ... existing initialization
    
    // Initialize Binance client
    binanceClient := binance.NewClient(&cfg.Binance)
    
    return &Container{
        Cfg:           cfg,
        DB:            db,
        Repo:          repo,
        Srvc:          srvc,
        Ctrl:          ctrl,
        BinanceClient: binanceClient,  // ← Add to container
    }, nil
}
```

---

### **Usage in Service Layer**

```go
package service

import (
    "context"
    "github.com/reshap/trading-bot/internal/clients/binance"
)

func (s *Services) GetCryptoPrice(symbol string) (float64, error) {
    ctx := context.Background()
    
    // Use Binance client
    price, err := s.binanceClient.GetPrice(ctx, symbol)
    if err != nil {
        return 0, fmt.Errorf("failed to get price: %w", err)
    }
    
    return price, nil
}

func (s *Services) ExecuteTrade(req *TradeRequest) error {
    ctx := context.Background()
    
    // Place order via Binance client
    orderReq := &binance.PlaceOrderRequest{
        Symbol:   req.Symbol,
        Side:     req.Side,
        Type:     "MARKET",
        Quantity: req.Quantity,
    }
    
    resp, err := s.binanceClient.PlaceOrder(ctx, orderReq)
    if err != nil {
        return fmt.Errorf("failed to place order: %w", err)
    }
    
    // Save to database
    return s.saveOrderToDB(resp)
}
```

---

### **Error Handling**

```go
package binance

import (
    "errors"
    "fmt"
)

// Common errors
var (
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrInvalidSymbol       = errors.New("invalid symbol")
    ErrOrderFailed         = errors.New("order failed")
    ErrAPIRateLimit        = errors.New("API rate limit exceeded")
)

// handleError handles API errors
func (c *Client) handleError(err error) error {
    var apiErr *binance.APIError
    
    if errors.As(err, &apiErr) {
        switch apiErr.Code {
        case -1013:
            return ErrInvalidSymbol
        case -1021:
            return ErrAPIRateLimit
        case -2010:
            return ErrInsufficientBalance
        default:
            return fmt.Errorf("%w: %s", ErrOrderFailed, apiErr.Msg)
        }
    }
    
    return err
}
```

---

### **Testing**

```go
package binance

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestClient_GetPrice(t *testing.T) {
    client := NewClient(testConfig)
    
    ctx := context.Background()
    price, err := client.GetPrice(ctx, "BTCUSDT")
    
    assert.NoError(t, err)
    assert.Greater(t, price, float64(0))
}

func TestClient_PlaceOrder(t *testing.T) {
    client := NewClient(testConfig)
    
    req := &PlaceOrderRequest{
        Symbol:   "BTCUSDT",
        Side:     "BUY",
        Type:     "MARKET",
        Quantity: 0.001,
    }
    
    ctx := context.Background()
    resp, err := client.PlaceOrder(ctx, req)
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Greater(t, resp.OrderID, int64(0))
}
```

---

### **Summary**

| File | Purpose | Required |
|------|---------|----------|
| `client.go` | Main client struct + constructor | ✅ Yes |
| `service.go` | Business logic methods | ✅ Yes |
| `dto.go` | Request/Response DTOs | ✅ Yes |
| `config.go` | Configuration | Optional |
| `helper.go` | Helper functions | Optional |

**Key Points:**
- ✅ Always use `context.Context` for API calls
- ✅ Wrap errors with descriptive messages
- ✅ Validate requests before API calls
- ✅ Return custom DTOs, not raw API responses
- ✅ Implement proper error handling
- ✅ Write tests for client methods

---

## 📚 References

- [Gin Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Go Best Practices](https://github.com/golang-standards/project-layout)
- [Clean Architecture](https://blog.cleancoder.com/)

---

*Last Updated: March 8, 2026*
