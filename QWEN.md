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
│   ├── public/                  # Static assets (favicon, robots.txt)
│   ├── src/
│   │   ├── assets/              # Images, styles
│   │   │   ├── style/
│   │   │   │   └── main.css     # TailwindCSS + custom styles
│   │   │   └── *.svg            # Images, icons
│   │   ├── components/          # Reusable Vue components
│   │   │   ├── common/          # Shared UI components
│   │   │   │   ├── index.ts     # Barrel export
│   │   │   │   ├── UiInput.vue
│   │   │   │   ├── UiButton.vue
│   │   │   │   └── UiPassword.vue
│   │   │   └── features/        # Feature-specific components
│   │   ├── helpers/             # Helper functions (validation, formatters)
│   │   ├── layouts/             # Layout components
│   │   ├── lib/                 # Library configurations + generic interfaces
│   │   │   ├── axios.ts         # Axios instance, interceptors, IApiResponse<T>
│   │   │   ├── storage.ts       # LocalStorage utilities
│   │   │   ├── sweetalert.ts    # SweetAlert2 wrapper
│   │   │   └── tailwind.ts      # Tailwind config
│   │   ├── pages/               # Page-level components (route-level)
│   │   ├── router/              # Vue Router configuration
│   │   ├── stores/              # Pinia stores + module-specific interfaces
│   │   │                        # + form state & validation (Vuelidate)
│   │   ├── App.vue              # Root component
│   │   └── main.ts              # Application entry point
│   ├── .env                     # Environment variables
│   ├── .env.example             # Environment template
│   ├── eslint.config.js         # ESLint configuration
│   ├── package.json             # Dependencies
│   ├── tailwind.config.js       # TailwindCSS config
│   ├── tsconfig.json            # TypeScript config
│   └── vite.config.ts           # Vite configuration
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

> **📚 Looking for more details?**
> Lihat [`be/docs/CODING_RULES.md`](./be/docs/CODING_RULES.md) untuk panduan lengkap Backend dengan:
> - Clean Architecture patterns & layer responsibilities
> - Generic Repository usage & transaction management
> - Service & Controller naming conventions
> - DTO transformation patterns
> - Error handling best practices
> - Complete API response format

---

# 🎨 Frontend (Vue 3 + TS)

## Architecture

```
Pages → Components → Stores → API
        (UI)      (State)   (HTTP)
```

**Layer Responsibilities:**

| Layer | Responsibility | Should NOT |
|-------|---------------|------------|
| **Pages** | Route-level components, compose stores & components | Business logic, API calls |
| **Components** | UI logic, user interaction, emit events | API calls, business logic |
| **Stores** | State management, form state, validation, API calls | HTTP specifics, UI logic |
| **Lib** | Shared utilities (axios, storage, sweetalert) | Business logic |

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
| **ESLint** | Linting | 10.x |
| **Prettier** | Formatting | 3.x |
| **Yarn** | Package Manager | 4.x |

## Naming Conventions

### TypeScript Naming

| Construct | Pattern | Example |
|-----------|---------|---------|
| **Interface** | `I` + PascalCase | `IWatchlist`, `IApiResponse` |
| **Type** | `T` + PascalCase | `TSignalAction`, `TTimestamp` |
| **Generic Param** | `T` + descriptive | `TData`, `TEntity`, `TSource` |
| **Class** | PascalCase | `WatchlistClass` |
| **Enum** | PascalCase | `SignalAction` |
| **Function** | camelCase | `getWatchlists`, `createWatchlist` |
| **Variable** | camelCase | `watchlistData`, `isLoading` |
| **Constant** | UPPER_CASE | `BASE_URL`, `TOKEN_KEY` |

### File Naming

| Type | Convention | Example |
|------|-----------|---------|
| **Vue Components** | PascalCase | `WatchlistTable.vue`, `AppHeader.vue` |
| **Pages** | PascalCase dengan suffix `Page` | `LoginPage.vue`, `HomePage.vue` |
| **Composables** | camelCase dengan `use` prefix | `useWatchlist.ts`, `useAuth.ts` |
| **Stores** | camelCase dengan `.store.ts` | `watchlist.store.ts`, `auth.store.ts` |
| **Helpers** | camelCase | `formatters.ts`, `validators.ts` |

### Folder Naming

- **Lowercase** dengan hyphen jika perlu: `components/`, `api-services/`
- **Kecuali**: `components/common/`, `components/features/` (subfolder)

## Critical Patterns

### 1. Interface Location

**Generic interfaces** → `src/lib/axios.ts`
```typescript
// src/lib/axios.ts
export interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

export interface IPaginatedResponse<TData> {
  data: TData[]
  total: number
  page: number
  per_page: number
}
```

**Module-specific interfaces** → `src/stores/<module>.store.ts` (di dalam file store)
```typescript
// src/stores/auth.store.ts
interface ILoginRequest {
  username: string
  password: string
}

export interface ILoginResponse {
  token: string
  user?: IUser
}

export interface IUser {
  id: number
  email: string
  name?: string
  created_at?: string
}
```

> **💡 Rationale:** Dengan menaruh interface di file yang menggunakannya, kita meningkatkan **cohesion** dan memudahkan maintenance. Perubahan API = edit 1 file saja.

### 2. Form State & Validation di Store

```typescript
// Store menyimpan form state dan validation
export const useAuthStore = defineStore('auth', () => {
  // Form state
  const loginReq = ref<ILoginRequest>({
    username: '',
    password: ''
  })

  // Validation rules
  const loginRules = ref({
    username: { required },
    password: { required }
  })

  // Vuelidate instance
  const loginReqValid = useVuelidate(loginRules, loginReq)

  async function loginAction(): Promise<boolean> {
    // Validate form before submit
    const valid = await loginReqValid.value.$validate()
    if (!valid) return false
    // API call
  }

  return { loginReq, loginReqValid, loginAction }
})
```

### 3. Component Menggunakan Store

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth.store'
import { getValidationErrors } from '@/helpers/validation'

const authStore = useAuthStore()

// Access form state dan validation dari store
const loginReq = authStore.loginReq
const v$ = authStore.loginReqValid
const isLoading = computed(() => authStore.loading)

const handleSubmit = async () => {
  // Trigger validation dari store
  const valid = await v$.value.$validate()
  if (!valid) return
  const success = await authStore.loginAction()
}

// Handle Enter key
const handleKeyPress = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" @keypress="handleKeyPress">
    <UiInput
      v-model="loginReq.username"
      label="Username"
      placeholder="Enter your username"
      :error="v$.username.$error"
      :error-message="getValidationErrors(v$.username).join(', ')"
    />

    <UiPassword
      v-model="loginReq.password"
      label="Password"
      placeholder="Enter your password"
      :error="v$.password.$error"
      :error-message="getValidationErrors(v$.password).join(', ')"
    />

    <UiButton
      type="submit"
      variant="primary"
      :loading="isLoading"
      full-width
    >
      {{ isLoading ? 'Signing in...' : 'Sign In' }}
    </UiButton>
  </form>
</template>
```

> **💡 Pattern:** Component hanya menangani UI dan user interaction. Semua logic (form state, validation, API call) ada di **store**.

### 4. Generic API Response

```typescript
// Import dari lib
import { IApiResponse } from '@/lib/axios'

// Usage
const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
```

### 5. Phosphor Icons ⭐

**Library:** `@phosphor-icons/vue` (50,000+ icons)

```vue
<script setup lang="ts">
// Import individual icons (JANGAN wildcard)
import { PhPlus, PhTrash, PhPencilSimple, PhCheckCircle } from '@phosphor-icons/vue'
</script>

<template>
  <!-- Basic usage -->
  <PhPlus />

  <!-- With size & weight -->
  <PhPlus :size="20" weight="bold" />

  <!-- With Tailwind classes -->
  <PhPlus :size="20" weight="bold" class="text-blue-500 hover:text-blue-600" />

  <!-- In buttons -->
  <button class="flex items-center gap-2">
    <PhPencilSimple :size="16" weight="bold" />
    Edit
  </button>

  <!-- Status icons -->
  <PhCheckCircle :size="20" class="text-green-500" weight="fill" />
</template>
```

**Icon Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `size` | `number \| string` | `24` | Icon size (pixels) |
| `weight` | `'thin' \| 'light' \| 'regular' \| 'bold' \| 'fill' \| 'duotone'` | `'regular'` | Icon weight/style |

**Common Icons:**
```typescript
// Actions
PhPlus          // Add/Create
PhTrash         // Delete
PhPencilSimple  // Edit
PhX             // Close/Cancel
PhCheck         // Confirm/Success

// Status
PhCheckCircle   // Success
PhXCircle       // Error
PhWarning       // Warning
PhInfo          // Info

// Navigation
PhCaretLeft     // Back
PhCaretRight    // Forward
PhHouse         // Home
PhGear          // Settings
```

**Browse icons:** https://phosphoricons.com/

### 6. UI Components

**Location:** `src/components/common/`

**Reusable Components:**
- `UiInput.vue` - Input dengan label, placeholder, error state
- `UiPassword.vue` - Password input dengan show/hide toggle
- `UiButton.vue` - Button dengan variants (primary, danger, outline)

**Usage:**
```vue
<script setup lang="ts">
import { UiInput, UiButton, UiPassword } from '@/components/common'
</script>

<template>
  <UiInput
    v-model="username"
    label="Username"
    placeholder="Enter your username"
    :error="v$.username.$error"
    :error-message="getValidationErrors(v$.username).join(', ')"
  />

  <UiPassword
    v-model="password"
    label="Password"
    placeholder="Enter your password"
    :error="v$.password.$error"
    :error-message="getValidationErrors(v$.password).join(', ')"
  />

  <UiButton
    type="submit"
    variant="primary"
    :loading="isLoading"
    full-width
  >
    Sign In
  </UiButton>
</template>
```

### 7. Layout Usage

**Location:** `src/layouts/`

**Pattern:**
```vue
<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store'
import AppHeader from '@/components/AppHeader.vue'

const authStore = useAuthStore()
const user = authStore.user
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <AppHeader />
    <main class="container mx-auto px-4 py-8">
      <slot />
    </main>
  </div>
</template>
```

**Usage di Pages:**
```vue
<script setup lang="ts">
import DefaultLayout from '@/layouts/DefaultLayout.vue'
</script>

<template>
  <DefaultLayout>
    <h1>Dashboard</h1>
    <!-- Page content -->
  </DefaultLayout>
</template>
```

### 8. Styling - TailwindCSS First ⭐

**Priority Order:**
1. ✅ **TailwindCSS utility classes** (PRIMARY choice)
2. ✅ **@apply directive** (untuk reusable patterns, gunakan sparingly)
3. ⚠️ **Custom CSS** (HANYA ketika Tailwind tidak support)

**✅ GOOD: Full Tailwind**
```vue
<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm p-5
              transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
    <h3 class="text-lg font-semibold text-gray-900">Title</h3>
  </div>
</template>
<!-- No <style> section needed! -->
```

**✅ ACCEPTABLE: @apply untuk reusable patterns**
```vue
<template>
  <button class="btn-primary">Click Me</button>
</template>

<style scoped>
.btn-primary {
  @apply px-4 py-2 bg-blue-500 text-white rounded-lg
         hover:bg-blue-600 transition-colors;
}
</style>
```

**⚠️ ONLY WHEN NECESSARY: Custom CSS**
```vue
<template>
  <div class="custom-animation">Content</div>
</template>

<style scoped>
.custom-animation {
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
</style>
```

**When to use Custom CSS:**
- ✅ Complex animations (@keyframes)
- ✅ CSS grid templates (complex layouts)
- ✅ Custom scrollbars
- ✅ Print styles
- ✅ CSS variables untuk theming

### 9. Component Structure

**Single File Component (SFC) Order:**
```vue
<script setup lang="ts">
// 1. Imports (Vue, libraries, interfaces, stores)
import { ref, computed, onMounted } from 'vue'
import { IUser } from '@/stores/auth.store'
import { useAuthStore } from '@/stores/auth.store'

// 2. Props & Emits
interface ILoginPageProps {
  initialEmail?: string
}

const props = withDefaults(defineProps<ILoginPageProps>(), {
  initialEmail: '',
})

const emit = defineEmits<{
  loginSuccess: [email: string]
}>()

// 3. State
const email = ref(props.initialEmail)
const password = ref('')

// 4. Computed
const isValidEmail = computed(() => email.value.includes('@'))

// 5. Methods
const handleSubmit = async () => {
  // submit logic
}

// 6. Lifecycle
onMounted(() => {
  // initialization
})
</script>

<template>
  <!-- Template content -->
</template>

<style scoped>
/* Styles */
</style>
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
yarn install

# Run dev server
yarn dev

# Build for production
yarn build
```

### Linting & Formatting

```bash
cd fe

# Run ESLint
yarn lint

# Format dengan Prettier
yarn format

# Build dan verify
yarn build
```

### Code Style

**Formatting Rules:**
- **Semi**: `false` (no semicolons)
- **Single Quote**: `true`
- **Tab Width**: `2`
- **Print Width**: `100`
- **Trailing Comma**: `none`

**Import Order:**
```typescript
// 1. Vue & libraries
import { ref, computed } from 'vue'

// 2. Libraries (axios, sweetalert)
import axios from 'axios'
import Swal from 'sweetalert2'

// 3. Interfaces & Types (from lib or stores)
import { type IApiResponse } from '@/lib/axios'
import { IUser, ILoginRequest } from '@/stores/auth.store'

// 4. Stores
import { useAuthStore } from '@/stores/auth.store'

// 5. Helpers
import { formatDate } from '@/helpers/formatters'

// 6. Components (in Vue templates)
import AppHeader from '@/components/AppHeader.vue'
```

> **📚 Looking for more details?**
> Lihat [`fe/docs/CODING_RULES.md`](./fe/docs/CODING_RULES.md) untuk panduan lengkap Frontend dengan:
> - Complete Vue component structure & SFC order
> - Form validation dengan Vuelidate (complete examples)
> - Phosphor Icons usage (50,000+ icons reference)
> - UI Components API (UiInput, UiButton, UiPassword)
> - Layout patterns & usage
> - TailwindCSS styling best practices
> - TypeScript naming conventions (detailed)
> - Code style & formatting rules

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

1. **Create Pinia Store** (`src/stores/<module>.store.ts`)
   - Define interfaces/types **langsung di dalam file store** (tidak dipisah di folder interfaces/ atau services/)
   - Define generic interfaces di `src/lib/axios.ts` jika reusable (contoh: `IApiResponse<T>`)
   - Define form state (`ref<IInterface>()`)
   - Define validation rules
   - Create Vuelidate instance
   - Create actions dengan API call langsung (tidak lewat service layer)
2. **Create Components** (`src/components/features/<module>/`)
   - Gunakan reusable UI components dari `src/components/common/` (UiInput, UiButton, UiPassword)
   - Import Phosphor Icons secara individual (`import { PhPlus } from '@phosphor-icons/vue'`)
   - Gunakan TailwindCSS utility classes (hindari custom CSS jika tidak perlu)
3. **Create Page** (`src/pages/<module>Page.vue`)
   - Gunakan layout pattern (`import DefaultLayout from '@/layouts/DefaultLayout.vue'`)
   - Access form state dan validation dari store
   - Handle user interaction dan emit events ke components
4. **Add Route** (`src/router/index.ts`)

**Important Notes:**
- ❌ **JANGAN** buat folder `src/interfaces/` atau `src/services/` untuk feature baru
- ✅ **Interfaces** → Langsung di `src/stores/*.store.ts`
- ✅ **API Calls** → Langsung di store actions (tidak ada service layer)
- ✅ **Composables** → Untuk reusable logic (`useWatchlist.ts`, `useAuth.ts`)
- ✅ **UI Components** → Gunakan dari `src/components/common/`
- ✅ **Icons** → Import individual dari `@phosphor-icons/vue`
- ✅ **Styling** → TailwindCSS utility classes first

**Example:**
```typescript
// 1. Store (src/stores/signal.store.ts)
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'

export interface ISignal {
  id: number
  symbol: string
  side: 'LONG' | 'SHORT'
  status: string
  created_at: string
}

export const useSignalStore = defineStore('signal', () => {
  const items = ref<ISignal[]>([])
  const loading = ref(false)

  const activeSignals = computed(() => items.value.filter(s => s.status === 'ACTIVE'))

  async function fetchSignals() {
    loading.value = true
    try {
      const response = await get<IApiResponse<ISignal[]>>('/signals')
      items.value = response.data.data
    } catch (error: any) {
      showError('Error', error.message)
    } finally {
      loading.value = false
    }
  }

  return { items, loading, activeSignals, fetchSignals }
})

// 2. Page (src/pages/SignalsPage.vue)
<script setup lang="ts">
import { onMounted } from 'vue'
import { useSignalStore } from '@/stores/signal.store'
import { PhPlus, PhRefresh } from '@phosphor-icons/vue'

const signalStore = useSignalStore()

onMounted(async () => {
  await signalStore.fetchSignals()
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Signals</h1>
      <button class="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg">
        <PhPlus :size="16" weight="bold" />
        Add Signal
      </button>
    </div>

    <div class="grid grid-cols-1 gap-4">
      <div v-for="signal in signalStore.items" :key="signal.id"
           class="bg-white border border-gray-200 rounded-xl shadow-sm p-5">
        <h3 class="text-lg font-semibold text-gray-900">{{ signal.symbol }}</h3>
        <p class="text-gray-600">{{ signal.side }} - {{ signal.status }}</p>
      </div>
    </div>
  </div>
</template>
```

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
| `CODING_RULES.md` | Frontend coding standards & conventions (lengkap!) |
| `INTERFACES_AND_TYPES.md` | TypeScript interfaces & types structure |
| `SWEETALERT_USAGE.md` | SweetAlert2 usage guide |
| `TAILWIND_CONFIG.md` | TailwindCSS v4 configuration |
| `TYPESCRIPT_CONVENTION.md` | TypeScript naming conventions |
| `VUELIDATE_USAGE.md` | Form validation dengan Vuelidate |
| `PHOSPHOR_ICONS_USAGE.md` | Phosphor Icons usage guide |
| `COMPONENT_GUIDE.md` | Reusable UI components guide |
| `LAYOUT_GUIDE.md` | Layout patterns & usage |

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
- [ESLint](https://eslint.org/docs/latest/)
- [Prettier](https://prettier.io/docs/en/)
- [Yarn](https://yarnpkg.com/)

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
