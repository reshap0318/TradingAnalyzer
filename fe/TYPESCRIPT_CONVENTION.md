# TypeScript Coding Conventions

## 📋 Naming Conventions

### Interface: Prefix `I`

Semua **interface** WAJIB menggunakan prefix `I`:

```typescript
// ✅ GOOD
interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
}

interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

interface IWatchlistRequest {
  symbol: string
  is_active?: boolean
}

// ❌ BAD
interface Watchlist { ... }
interface ApiResponse<T> { ... }
```

### Type: Prefix `T`

Semua **type** WAJIB menggunakan prefix `T`:

```typescript
// ✅ GOOD
type TSignalAction = 'BUY' | 'SELL' | 'WAIT'
type TOrderStatus = 'PENDING' | 'FILLED' | 'CANCELLED'
type TTimestamp = string
type TCoordinate = [number, number]

type TWatchlistWithTrades = IWatchlist & {
  trades: ITrade[]
}

// ❌ BAD
type SignalAction = ...
type OrderStatus = ...
```

### Generic Type Parameters

Untuk **generic type parameters**, gunakan `T` + descriptive name (lebih readable):

```typescript
// ✅ GOOD
interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

interface IRepository<TEntity> {
  findById(id: number): Promise<TEntity | null>
  findAll(): Promise<TEntity[]>
}

interface IMapper<TSource, TDestination> {
  map(source: TSource): TDestination
}

// ❌ BAD: Terlalu singkat, kurang deskriptif
interface IApiResponse<T> { ... }
interface IRepository<T> { ... }
```

---

## 🏗️ When to Use Interface vs Type

### Gunakan `interface` (prefix `I`) untuk:

**1. Object Shapes (DTOs, Models, API Responses)**
```typescript
interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
}

interface IStrategy {
  id: number
  strategy_name: string
  primary_tf: string
  is_active: boolean
}
```

**2. Service & Repository Contracts**
```typescript
interface IWatchlistService {
  getWatchlists(): Promise<IApiResponse<IWatchlist[]>>
  createWatchlist(data: IWatchlistRequest): Promise<IApiResponse<IWatchlist>>
}

interface IRepository<TEntity> {
  findById(id: number): Promise<TEntity | null>
  findAll(): Promise<TEntity[]>
  create(entity: TEntity): Promise<TEntity>
}
```

**3. Component Props**
```typescript
interface IWatchlistProps {
  watchlist: IWatchlist
  onSelect: (id: number) => void
  onDelete: (id: number) => void
}
```

---

### Gunakan `type` (prefix `T`) untuk:

**1. Union Types**
```typescript
type TSignalAction = 'BUY' | 'SELL' | 'WAIT'
type TSignalStrength = 'STRONG_BUY' | 'BUY' | 'WAIT' | 'SELL' | 'STRONG_SELL'
type TOrderSide = 'LONG' | 'SHORT'
type TTimeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d'
```

**2. Primitive Aliases**
```typescript
type TUserID = string
type TTimestamp = string
type TPercentage = number
type TCryptoSymbol = `${string}${string}`
```

**3. Tuple Types**
```typescript
type TCoordinate = [number, number]
type TTimeRange = [string, string] // [start, end]
type TRGB = [number, number, number]
```

**4. Mapped & Utility Types**
```typescript
type TReadonly<T> = {
  readonly [K in keyof T]: T[K]
}

type TNullable<T> = {
  [K in keyof T]: T[K] | null
}

type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at'>
type TWatchlistUpdateInput = Partial<IWatchlistCreateInput>
```

**5. Intersection Types**
```typescript
type TWatchlistWithTrades = IWatchlist & {
  trades: ITrade[]
}

type TPaginatedResponse<TData> = IApiResponse<TData[]> & {
  pagination: {
    page: number
    limit: number
    total: number
  }
}
```

---

## 📝 File Organization

### Single Interface/Type per Concept

```typescript
// ✅ GOOD: Group related interfaces in one file
// src/interfaces/watchlist.ts
export interface IWatchlist { ... }
export interface IWatchlistRequest { ... }
export interface IWatchlistResponse { ... }

// ❌ BAD: Multiple unrelated interfaces
```

### File Naming

```
interfaces/
├── watchlist.ts       # IWatchlist, IWatchlistRequest, IWatchlistResponse
├── signal.ts          # ISignal, ISignalEntry, ISignalRequest
├── strategy.ts        # IStrategy, IStrategyTimeframe
└── common.ts          # IApiResponse, IPagination, IMeta

types/
├── signal.ts          # TSignalAction, TSignalStrength, TOrderSide
├── timeframe.ts       # TTimeframe, TTimeframeMinutes
└── utils.ts           # TNullable, TReadonly, TPaginated
```

---

## 🎯 Examples

### API Service Pattern

```typescript
// src/interfaces/watchlist.ts
export interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
}

export interface IWatchlistRequest {
  symbol: string
  is_active?: boolean
}

export interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

// src/types/watchlist.ts
export type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at'>
export type TWatchlistUpdateInput = Partial<TWatchlistCreateInput>

// src/services/watchlist.service.ts
import { IWatchlist, IWatchlistRequest, IApiResponse } from '@/interfaces/watchlist'

export const getWatchlists = async (): Promise<IApiResponse<IWatchlist[]>> => {
  const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
  return response.data
}
```

### Store Pattern (Pinia)

```typescript
// src/stores/watchlist.store.ts
import { defineStore } from 'pinia'
import { IWatchlist } from '@/interfaces/watchlist'

interface IWatchlistState {
  items: IWatchlist[]
  selected: IWatchlist | null
  loading: boolean
  error: string | null
}

export const useWatchlistStore = defineStore('watchlist', {
  state: (): IWatchlistState => ({
    items: [],
    selected: null,
    loading: false,
    error: null,
  }),

  actions: {
    async fetchWatchlists(): Promise<void> {
      this.loading = true
      try {
        const response = await watchlistService.getWatchlists()
        this.items = response.data
      } catch (error) {
        this.error = 'Failed to fetch watchlists'
      } finally {
        this.loading = false
      }
    },
  },
})
```

### Component Props

```typescript
// src/components/WatchlistTable.vue
<script setup lang="ts">
import { IWatchlist } from '@/interfaces/watchlist'

interface IWatchlistTableProps {
  watchlists: IWatchlist[]
  loading?: boolean
  selectable?: boolean
}

const props = withDefaults(defineProps<IWatchlistTableProps>(), {
  loading: false,
  selectable: true,
})

const emit = defineEmits<{
  select: [id: number]
  delete: [id: number]
}>()
</script>
```

---

## 🔧 ESLint Rules

Untuk enforce conventions, tambahkan ke `eslint.config.js`:

```javascript
{
  rules: {
    '@typescript-eslint/naming-convention': [
      'error',
      {
        selector: 'interface',
        format: ['PascalCase'],
        custom: {
          regex: '^I[A-Z]',
          match: true,
        },
      },
      {
        selector: 'typeAlias',
        format: ['PascalCase'],
        custom: {
          regex: '^T[A-Z]',
          match: true,
        },
      },
    ],
  },
}
```

---

## 📚 Quick Reference

| Construct | Prefix | Example |
|-----------|--------|---------|
| **Interface** | `I` | `IWatchlist`, `IApiResponse` |
| **Type** | `T` | `TSignalAction`, `TTimestamp` |
| **Generic Param** | `T` + descriptive | `TData`, `TEntity`, `TSource` |
| **Class** | None (PascalCase) | `WatchlistService`, `ApiClient` |
| **Enum** | None (PascalCase) | `SignalAction`, `OrderStatus` |
| **Function** | camelCase | `getWatchlists`, `createWatchlist` |
| **Variable** | camelCase | `watchlistData`, `isLoading` |
| **Constant** | UPPER_CASE | `BASE_URL`, `TOKEN_KEY` |

---

## 🎓 Summary

```
┌──────────────────────────────────────────────────────┐
│  TypeScript Naming Convention - Quick Guide          │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Interface  → I + PascalCase    (IWatchlist)         │
│  Type       → T + PascalCase    (TSignalAction)      │
│  Generic    → T + Descriptive   (TData, TEntity)     │
│  Class      → PascalCase       (WatchlistService)    │
│  Function   → camelCase         (getWatchlists)      │
│  Variable   → camelCase         (watchlistData)      │
│  Constant   → UPPER_CASE        (BASE_URL)           │
│                                                      │
│  Rule of Thumb:                                      │
│  - Interface = Object shapes & contracts             │
│  - Type      = Unions, primitives, utilities         │
│  - Generic   = T + descriptive (TData, TEntity)      │
└──────────────────────────────────────────────────────┘
```
