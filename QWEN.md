# Trading Bot v4 - Project Context

## 📋 Table of Contents

1. [Global Overview](#-global-overview)
2. [Backend (Go)](#-backend-go)
3. [Frontend (Vue 3 + TS)](#-frontend-vue-3--ts)
4. [Infrastructure](#-infrastructure)
5. [Development Workflow](#-development-workflow)
6. [Documentation](#-documentation)

---

# 🌍 Global Overview

## Project Description

**Trading Bot** adalah aplikasi automated trading untuk Binance Futures dengan arsitektur modern.

### Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.25.0, Gin, GORM, Redis |
| **Frontend** | Vue 3, TypeScript, Pinia, Vue Router, TailwindCSS 4 |
| **Database** | MySQL |
| **Cache** | Redis |
| **External API** | Binance Futures (Testnet & Mainnet) |
| **Infrastructure** | Docker, Docker Compose, Nginx |

### Project Structure

```
appv4/
├── be/                          # Backend (Go)
│   ├── cmd/
│   │   └── main/                # Application entry point
│   ├── internal/
│   │   ├── clients/             # External API clients (Binance, S3)
│   │   ├── config/              # Configuration management
│   │   ├── controller/          # HTTP request handlers
│   │   ├── database/            # DB & Redis connections
│   │   ├── di/                  # Dependency Injection
│   │   ├── dtos/                # Data Transfer Objects
│   │   ├── helpers/             # Utility functions
│   │   ├── middleware/          # HTTP middleware (Auth, CORS)
│   │   ├── models/              # Domain models / entities
│   │   ├── repository/          # Data access layer (Generic CRUD)
│   │   ├── routes/              # Route definitions
│   │   └── service/             # Business logic layer
│   ├── docs/                    # Documentation (API, flows, rules)
│   ├── go.mod                   # Go module definition
│   └── .env.example             # Environment template
│
├── fe/                          # Frontend (Vue 3 + TS)
│   ├── docs/                    # Frontend documentation
│   ├── public/                  # Static assets
│   ├── src/
│   │   ├── assets/              # Images, styles
│   │   │   └── style/
│   │   │       └── main.css     # TailwindCSS + custom styles
│   │   ├── components/
│   │   │   ├── common/          # Reusable UI components
│   │   │   │   ├── index.ts     # Barrel export
│   │   │   │   ├── UiInput.vue
│   │   │   │   ├── UiButton.vue
│   │   │   │   └── UiPassword.vue
│   │   │   └── features/        # Feature-specific components
│   │   ├── layouts/             # Layout components
│   │   ├── lib/
│   │   │   ├── axios.ts         # Axios instance + IApiResponse<T>
│   │   │   ├── storage.ts       # LocalStorage utilities
│   │   │   ├── sweetalert.ts    # SweetAlert2 wrapper
│   │   │   └── validation.ts    # Validation helper
│   │   ├── pages/               # Page-level components
│   │   ├── router/              # Vue Router config
│   │   ├── stores/              # Pinia stores + form state
│   │   ├── App.vue              # Root component
│   │   └── main.ts              # Entry point
│   ├── .env.example
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── etc/
│   ├── go/Dockerfile            # Backend Docker config
│   ├── node/Dockerfile          # Frontend Docker config
│   └── nginx/conf.d/            # Nginx reverse proxy config
│
└── docker-compose.yml           # Docker orchestration
```

### Key Features

1. **Automated Trading Bot**
   - Signal generation dari multiple indicators (MACD, RSI, Stochastic, Bollinger, MA, Volume, Support/Resistance)
   - Risk management dengan 5-layer validation
   - Auto TP/SL placement via Binance Algo Orders
   - Position management (ISOLATED margin)

2. **Money Management**
   - Max daily trades limit
   - Max consecutive loss limit
   - Max daily loss percentage
   - Minimum confidence threshold
   - Risk-reward ratio validation

3. **Multi-Timeframe Analysis**
   - Primary & secondary timeframe support
   - Signal scoring system
   - Weighted indicator system

4. **Trade Management**
   - Active trade monitoring
   - Auto TP/SL adjustment
   - Trade history & statistics

---

# � Backend (Go)

## Architecture

**Clean Architecture:**
```
Routes → Controller → Service → Repository → Database
         (HTTP)      (Logic)    (Data)       (Storage)
```

**Layer Responsibilities:**

| Layer | Responsibility | Should NOT |
|-------|---------------|------------|
| **Controller** | Handle HTTP requests, validation, response formatting | Business logic, DB queries |
| **Service** | Business logic, transactions, DTO transformation | HTTP specifics, direct DB access |
| **Repository** | Data access, CRUD operations | Business logic, HTTP specifics |
| **Models** | Domain entities, table mapping | Business logic, DB queries |

## Naming Conventions

**Function Naming Pattern:** `<Module><Action><Entity>`

```go
// Service
func (s *Services) WatchlistGetAll(ctx *gin.Context) (res []dtos.WatchlistData, err error)
func (s *Services) WatchlistCreate(ctx *gin.Context, req *dtos.WatchlistRequest) (res *dtos.WatchlistData, err error)

// Controller
func (c *Controller) WatchlistIndex(ctx *gin.Context)
func (c *Controller) WatchlistCreate(ctx *gin.Context)
```

## Critical Rules

1. **JANGAN modifikasi file core:**
   - `repository/00_generic.go` - Generic CRUD foundation
   - `repository/00_transaction.go` - Transaction management core

2. **SELALU gunakan Generic Repository:**
   ```go
   type YourEntityRepository struct {
       *GenericRepository[models.YourEntity]
   }
   ```

3. **WAJIB gunakan transaction untuk write operations:**
   ```go
   result, err := s.repo.TxManager.WithinTransactionWithResult(func(tx *gorm.DB) (interface{}, error) {
       // Pass tx ke repository methods
       entity, err = s.repo.YourEntity.Create(tx, entity)
       return entity, nil
   })
   ```

4. **Single struct per file** di Service & Controller

## API Response Format

**Success:**
```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

**Error:**
```json
{
    "code": 400,
    "message": "Error message here",
    "error": null
}
```

## Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/timeframes` | Get all timeframes |
| `GET` | `/api/indicators` | Get all indicators |
| `GET` | `/api/thresholds` | Get signal thresholds |
| `GET` | `/api/configs` | Get system configs |
| `POST` | `/api/trade/execute` | Execute single trade |
| `POST` | `/api/trade/monitor/:id` | Monitor trade by ID |
| `GET` | `/api/watchlists` | Get watchlist |
| `GET` | `/api/strategies` | Get trading strategies |
| `GET` | `/api/signals` | Get signal history |
| `GET` | `/api/backtests` | Get backtest results |

## Setup & Running

### Environment Setup

```bash
# Copy from .env.example
cp be/.env.example be/.env

# Edit be/.env dengan konfigurasi Anda
APP_HOST=0.0.0.0
APP_PORT=9000
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=trading_bot
REDIS_HOST=localhost
REDIS_PORT=6379
TESTNET_API_KEY=your_testnet_api_key
TESTNET_SECRET_KEY=your_testnet_secret_key
```

### Running Locally

```bash
cd be

# Install dependencies
go mod download

# Run with hot reload (air)
air -c .air-wind.toml    # Windows
air -c .air-linux.toml   # Linux

# Or run directly
go run cmd/main/main.go
```

### Testing

```bash
cd be
go test ./... -v
```

---

# 🎨 Frontend (Vue 3 + TS)

## Architecture

```
Pages → Components → Stores → Services → API
        (UI)      (State)   (HTTP)   (Backend)
```

## Tech Stack

| Technology | Purpose | Version |
|------------|---------|---------|
| **Vue 3** | Frontend Framework | 3.5+ |
| **TypeScript** | Type Safety | 5.9+ |
| **Vite** | Build Tool | 7.x |
| **Pinia** | State Management | 3.x |
| **Vue Router** | Routing | 5.x |
| **Axios** | HTTP Client | 1.x |
| **SweetAlert2** | Alerts/Modals | 11.x |
| **TailwindCSS** | Styling | 4.x |
| **Phosphor Icons** | Icon Library | 3.x (50,000+ icons) |
| **Vuelidate** | Form Validation | 3.x |

## Naming Conventions

| Construct | Pattern | Example |
|-----------|---------|---------|
| **Interface** | `I` + PascalCase | `IWatchlist`, `IApiResponse` |
| **Type** | `T` + PascalCase | `TSignalAction`, `TTimestamp` |
| **Generic Param** | `T` + descriptive | `TData`, `TEntity`, `TSource` |
| **Vue Component** | PascalCase | `WatchlistTable.vue`, `LoginPage.vue` |
| **Store** | camelCase + `.store.ts` | `watchlist.store.ts` |
| **Service** | camelCase | `watchlist.service.ts` |
| **Composable** | `use` + camelCase | `useWatchlist.ts` |

## Critical Patterns

### 1. Form State & Validation di Store

```typescript
// Store menyimpan form state dan validation
export const useAuthStore = defineStore('auth', () => {
  const loginReq = ref({ username: '', password: '' })
  const loginRules = ref({ username: { required }, password: { required } })
  const loginReqValid = useVuelidate(loginRules, loginReq)

  async function loginAction(): Promise<boolean> {
    const valid = await loginReqValid.value.$validate()
    if (!valid) return false
    // API call
  }

  return { loginReq, loginReqValid, loginAction }
})
```

### 2. Component Menggunakan Store

```vue
<script setup lang="ts">
const authStore = useAuthStore()
const loginReq = authStore.loginReq
const v$ = authStore.loginReqValid

const handleSubmit = async () => {
  const valid = await v$.value.$validate()
  if (!valid) return
  const success = await authStore.loginAction()
}
</script>
```

### 3. Generic API Response

```typescript
interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

// Usage
const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
```

### 4. Phosphor Icons

```vue
<script setup lang="ts">
import { PhPlus, PhTrash, PhPencilSimple } from '@phosphor-icons/vue'
</script>

<template>
  <PhPlus :size="20" weight="bold" class="text-blue-500" />
</template>
```

## Setup & Running

### Environment Setup

```bash
# Copy from .env.example
cp fe/.env.example fe/.env

# Edit fe/.env
VITE_API_BASE_URL=http://localhost:8000/api
VITE_APP_TITLE=TradingAnalyzer
```

### Running Locally

```bash
cd fe

# Install dependencies
npm install

# Run dev server
npm run dev

# Build for production
npm run build
```

### Linting & Testing

```bash
cd fe
npm run lint
npm run build
```

---

# 🏗️ Infrastructure

## Docker Services

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Stop all services
docker-compose down
```

**Services:**
- Backend: `http://localhost:9000`
- Frontend: `http://localhost:5173`
- Nginx (Production): `http://localhost:9001`

## Database Setup

```sql
CREATE DATABASE trading_bot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Migration akan berjalan otomatis saat aplikasi pertama kali dijalankan.

---

# 🛠️ Development Workflow

## Common Tasks

### Add New Feature (Backend)

1. **Create Model** (`internal/models/<entity>.go`)
2. **Create DTO** (`internal/dtos/<entity>_dto.go`)
3. **Create Repository** (`internal/repository/<entity>_repo.go`)
4. **Register Repository** (`internal/repository/00_repository.go`)
5. **Create Service** (`internal/service/<entity>_service.go`)
6. **Create Controller** (`internal/controller/<entity>_controller.go`)
7. **Create Routes** (`internal/routes/<entity>_routes.go`)
8. **Register Routes** (`cmd/main/main.go`)

### Add New Feature (Frontend)

1. **Create Interfaces/Types** (`src/interfaces/<module>.ts`, `src/types/<module>.ts`)
2. **Create API Service** (`src/services/<module>.service.ts`)
3. **Create Pinia Store** (`src/stores/<module>.store.ts`)
   - Define form state (`ref<Interface>()`)
   - Define validation rules
   - Create Vuelidate instance
   - Create actions with validation
4. **Create Components** (`src/components/features/<module>/`)
5. **Create Page** (`src/pages/<module>Page.vue`)
6. **Add Route** (`src/router/index.ts`)

### Database Migration

```bash
cd be
go run cmd/migration/main.go
```

### Generate Password

```bash
cd be
go run cmd/genpass/main.go
```

## Commit Task Guidelines

### Backend Commit Pattern

```bash
# Format: feat(BE): <description> / fix(BE): <description>
```

**Examples:**
```bash
# New feature
feat(BE): add watchlist CRUD endpoints
feat(BE): implement signal analysis service
feat(BE): add 5-layer validation for trade execution

# Bug fix
fix(BE): fix incorrect TP/SL calculation for SHORT positions
fix(BE): handle null values in signal response

# Refactor
refactor(BE): extract validation logic to separate service
refactor(BE): optimize database queries in watchlist repository

# Documentation
docs(BE): update API_DOCUMENTATION.md with new endpoints
docs(BE): add TRADE_EXECUTE_FLOW.md documentation

# Configuration
config(BE): add new threshold config for RSI indicator
config(BE): update .env.example with BINANCE_TESTNET key

# Testing
test(BE): add unit tests for signal calculation service
test(BE): add integration tests for trade execution flow
```

### Frontend Commit Pattern

```bash
# Format: feat(FE): <description> / fix(FE): <description>
```

**Examples:**
```bash
# New feature
feat(FE): add watchlist management page
feat(FE): implement signal analysis dashboard
feat(FE): add trade execution modal with validation
feat(FE): create reusable UiInput and UiButton components

# Bug fix
fix(FE): fix form validation not triggering on submit
fix(FE): handle null user data in auth store
fix(FE): fix SweetAlert import path in watchlist store
fix(FE): resolve TypeScript error in signal page

# Refactor
refactor(FE): move form state and validation to auth store
refactor(FE): extract API response interface to common.ts
refactor(FE): simplify component logic using composables

# Documentation
docs(FE): add AUTH_MODULE.md documentation
docs(FE): update CODING_RULES.md with Vuelidate pattern
docs(FE): add AXIOS_USAGE.md guide
docs(FE): document Phosphor Icons usage in CODING_RULES.md

# Style
style(FE): update Tailwind config with trading signal colors
style(FE): improve button loading state animation

# Configuration
config(FE): update Vite config for production build
config(FE): add new API base URL to .env.example
config(FE): update TypeScript config for strict mode

# Testing
test(FE): add unit tests for validation helper
test(FE): add E2E tests for login flow
```

### Commit Message Structure

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code refactoring
- `docs` - Documentation
- `style` - Styling/formatting
- `config` - Configuration changes
- `test` - Tests
- `chore` - Maintenance tasks

**Scopes:**
- `BE` - Backend
- `FE` - Frontend
- `infra` - Infrastructure (Docker, Nginx)
- (empty) - Full stack or general changes

---

# 📚 Documentation

## Backend Documentation

Lihat folder `be/docs/` untuk dokumentasi lengkap:

| File | Description |
|------|-------------|
| `API_DOCUMENTATION.md` | Complete API reference |
| `CODING_RULES.md` | Coding standards & conventions |
| `TRADE_EXECUTE_FLOW.md` | Trade execution flow (5-layer validation) |
| `TRADE_MONITOR_FLOW.md` | Trade monitoring flow |
| `SIGNAL_ANALYZE_RESPONSE.md` | Signal analysis response structure |
| `SIGNAL_BREAKDOWN.md` | Signal calculation per indicator |

## Frontend Documentation

Lihat folder `fe/docs/` untuk dokumentasi lengkap:

| File | Description |
|------|-------------|
| `README.md` | Documentation index |
| `AUTH_MODULE.md` | Authentication module & JWT token management |
| `AXIOS_USAGE.md` | API client configuration & usage |
| `CODING_RULES.md` | Frontend coding standards & conventions |
| `INTERFACES_AND_TYPES.md` | TypeScript interfaces & types structure |
| `SWEETALERT_USAGE.md` | SweetAlert2 usage guide |
| `TAILWIND_CONFIG.md` | TailwindCSS v4 configuration |
| `TYPESCRIPT_CONVENTION.md` | TypeScript naming conventions |

## Additional Resources

### Backend
- [Binance Futures API Documentation](https://binance-docs.github.io/apidocs/futures/en/)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Redis Documentation](https://redis.io/docs/)

### Frontend
- [Vue 3 Documentation](https://vuejs.org/)
- [Pinia Documentation](https://pinia.vuejs.org/)
- [Vue Router Documentation](https://router.vuejs.org/)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Vite Documentation](https://vitejs.dev/)
- [TailwindCSS v4](https://tailwindcss.com/docs)
- [Phosphor Icons](https://phosphoricons.com/)
- [SweetAlert2](https://sweetalert2.github.io/)
- [Vuelidate](https://vuelidate.js.org/)
- [Axios](https://axios-http.com/docs/intro)

### Infrastructure
- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Nginx Documentation](https://nginx.org/en/docs/)

---

# ⚠️ Important Notes

1. **Binance API:** Default menggunakan testnet. Set `BINANCE_TESTNET=false` di database untuk production.
2. **Auto-run Bot:** Set `BOT_AUTO_RUN=true` di `.env` untuk auto-start bot.
3. **Health Check:** `http://localhost:9000/health`
4. **CORS:** Configured untuk `localhost:5173`, `localhost:9001`, dan IP lokal.
