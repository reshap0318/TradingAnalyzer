# Interfaces & Types Structure

## 📁 Folder Structure

```
src/
├── interfaces/              # TypeScript interfaces (object shapes)
│   ├── common.ts            # Generic/reusable interfaces
│   ├── watchlist.ts         # Watchlist module interfaces
│   ├── signal.ts            # Signal module interfaces (future)
│   ├── strategy.ts          # Strategy module interfaces (future)
│   └── index.ts             # Barrel export
│
└── types/                   # Type aliases & union types
    ├── common.ts            # Common type utilities
    ├── watchlist.ts         # Watchlist-specific types
    └── index.ts             # Barrel export
```

---

## 📝 Interface Guidelines

### **When to Use `interfaces/`**

Gunakan untuk **object shapes** dan **contracts**:
- API Responses
- Request DTOs
- Domain models
- Component props
- Service contracts

### **Naming Convention**

- Prefix: `I` (Interface)
- Generic params: `T` + descriptive

```typescript
// ✅ GOOD
interface IWatchlist {
  id: number
  symbol: string
}

interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

// ❌ BAD
interface Watchlist { ... }  // Missing I prefix
interface IApiResponse<T> { ... }  // Generic too short
```

---

## 📄 File Organization

### **1. `interfaces/common.ts`**

Untuk **generic interfaces** yang digunakan di banyak tempat:

```typescript
// Standard API response
export interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

// Paginated response
export interface IPaginatedResponse<TData> extends IApiResponse<TData[]> {
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

// Pagination params
export interface IPaginationParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
}
```

**Kapan pakai `common.ts`?**
- ✅ Interface digunakan di 2+ modules
- ✅ Standard response format
- ✅ Shared utilities (pagination, metadata)

---

### **2. `interfaces/<module>.ts`**

Untuk **module-specific interfaces**:

```typescript
// interfaces/watchlist.ts
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

export interface IWatchlistQueryParams {
  is_active?: boolean
  search?: string
}
```

**Kapan pakai file terpisah?**
- ✅ Interface hanya dipakai di 1 module
- ✅ Group by feature/domain
- ✅ Clear ownership

---

### **3. `interfaces/index.ts`**

**Barrel export** untuk easy importing:

```typescript
export * from './common'
export * from './watchlist'
// export * from './signal'  // Future modules
```

**Usage:**
```typescript
// ✅ Import specific
import { IWatchlist } from '@/interfaces/watchlist'
import { IApiResponse } from '@/interfaces/common'

// ✅ Or import all (use sparingly)
import { IWatchlist, IApiResponse } from '@/interfaces'
```

---

## 🔧 Type Guidelines

### **When to Use `types/`**

Gunakan untuk:
- Union types
- Primitive aliases
- Utility/mapped types
- Tuple types

### **Naming Convention**

- Prefix: `T` (Type)
- Descriptive name

```typescript
// ✅ GOOD
type TSignalAction = 'BUY' | 'SELL' | 'WAIT'
type TTimestamp = string
type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at'>

// ❌ BAD
type SignalAction = ...  // Missing T prefix
```

---

## 📄 Type File Organization

### **1. `types/common.ts`**

Common utility types:

```typescript
// HTTP methods
export type THttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

// ID type
export type TId = number | string

// Timestamp
export type TTimestamp = string

// Status
export type TStatus = 'active' | 'inactive' | 'pending' | 'error'

// Utility types
export type TNullable<T> = {
  [K in keyof T]: T[K] | null
}
```

---

### **2. `types/<module>.ts`**

Module-specific types:

```typescript
// types/watchlist.ts
import type { IWatchlist, IWatchlistRequest } from '@/interfaces/watchlist'

// Create input (excludes id, timestamps)
export type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at' | 'updated_at'>

// Update input (partial)
export type TWatchlistUpdateInput = Partial<TWatchlistCreateInput>

// Form data (for UI)
export type TWatchlistFormData = IWatchlistRequest & {
  confirm?: boolean
}
```

---

## 💡 Usage Examples

### **In Services**

```typescript
// src/services/watchlist.service.ts
import { get, post, put, del } from '@/lib/axios'
import { IApiResponse } from '@/interfaces/common'
import { IWatchlist, IWatchlistRequest } from '@/interfaces/watchlist'

export const getWatchlists = async (): Promise<IApiResponse<IWatchlist[]>> => {
  const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
  return response.data
}

export const createWatchlist = async (
  data: IWatchlistRequest,
): Promise<IApiResponse<IWatchlist>> => {
  const response = await post<IApiResponse<IWatchlist>>('/watchlists', data)
  return response.data
}
```

---

### **In Stores (Pinia)**

```typescript
// src/stores/watchlist.store.ts
import { defineStore } from 'pinia'
import { IWatchlist } from '@/interfaces/watchlist'
import { TWatchlistCreateInput } from '@/types/watchlist'
import watchlistService from '@/services/watchlist.service'

interface IWatchlistState {
  items: IWatchlist[]
  selected: IWatchlist | null
  loading: boolean
}

export const useWatchlistStore = defineStore('watchlist', {
  state: (): IWatchlistState => ({
    items: [],
    selected: null,
    loading: false,
  }),

  actions: {
    async createWatchlist(data: TWatchlistCreateInput): Promise<void> {
      const response = await watchlistService.createWatchlist(data)
      this.items.push(response.data)
    },
  },
})
```

---

### **In Components**

```vue
<!-- src/components/WatchlistTable.vue -->
<script setup lang="ts">
import { IWatchlist } from '@/interfaces/watchlist'
import { TWatchlistTableRow } from '@/types/watchlist'

interface IWatchlistTableProps {
  watchlists: IWatchlist[]
  loading?: boolean
}

const props = withDefaults(defineProps<IWatchlistTableProps>(), {
  loading: false,
})

// Computed with type
const rows = computed<TWatchlistTableRow[]>(() => 
  props.watchlists.map(w => ({ ...w, isSelected: false }))
)
</script>

<template>
  <div>
    <!-- Template -->
  </div>
</template>
```

---

### **In Composables**

```typescript
// src/composables/useWatchlist.ts
import { ref, computed } from 'vue'
import { IWatchlist, IWatchlistRequest } from '@/interfaces/watchlist'
import { TWatchlistCreateInput } from '@/types/watchlist'
import watchlistService from '@/services/watchlist.service'

export const useWatchlist = () => {
  const items = ref<IWatchlist[]>([])
  const loading = ref(false)

  const createWatchlist = async (data: TWatchlistCreateInput) => {
    loading.value = true
    try {
      const response = await watchlistService.createWatchlist(data)
      items.value.push(response.data)
    } finally {
      loading.value = false
    }
  }

  return {
    items,
    loading,
    createWatchlist,
  }
}
```

---

## 🎯 Best Practices

### **1. Import Specific, Not All**

```typescript
// ✅ GOOD
import { IWatchlist } from '@/interfaces/watchlist'
import { IApiResponse } from '@/interfaces/common'

// ⚠️ AVOID (unless you need many)
import * as interfaces from '@/interfaces'
```

---

### **2. Use `type` for Derived Types**

```typescript
// ✅ GOOD
import { IWatchlist } from '@/interfaces/watchlist'

type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at'>
type TWatchlistUpdateInput = Partial<TWatchlistCreateInput>
```

---

### **3. Keep Interfaces Minimal**

```typescript
// ✅ GOOD: Focused interface
interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
}

// ❌ BAD: Too many optional fields
interface IWatchlist {
  id: number
  symbol: string
  is_active?: boolean  // Why optional?
  created_at?: string  // Always exists
  updated_at?: string  // Sometimes exists
  notes?: string       // Rarely used
  metadata?: any       // Avoid any!
}
```

---

### **4. Avoid `any` Type**

```typescript
// ❌ BAD
interface IApiResponse {
  data: any
}

// ✅ GOOD
interface IApiResponse<TData> {
  data: TData
}

// Usage: IApiResponse<IWatchlist[]>
```

---

### **5. Document Complex Types**

```typescript
/**
 * Watchlist form data with confirmation flag
 * Used in create/edit modals
 */
export type TWatchlistFormData = IWatchlistRequest & {
  confirm?: boolean  // Show confirmation dialog
  redirect?: boolean // Redirect after submit
}
```

---

## 📚 Quick Reference

| Location | Purpose | Example |
|----------|---------|---------|
| `interfaces/common.ts` | Generic interfaces | `IApiResponse<T>`, `IPaginatedResponse<T>` |
| `interfaces/<module>.ts` | Module interfaces | `IWatchlist`, `IWatchlistRequest` |
| `types/common.ts` | Utility types | `THttpMethod`, `TStatus`, `TNullable<T>` |
| `types/<module>.ts` | Module types | `TWatchlistCreateInput`, `TWatchlistUpdateInput` |
| `interfaces/index.ts` | Barrel export | `export * from './common'` |
| `types/index.ts` | Barrel export | `export * from './common'` |

---

## 🔄 Migration Guide

### **Before** (Inline interfaces)

```typescript
// ❌ service.ts
interface IWatchlist {
  id: number
  symbol: string
}

export const getWatchlists = async () => { ... }
```

### **After** (Centralized)

```typescript
// ✅ service.ts
import { IApiResponse } from '@/interfaces/common'
import { IWatchlist } from '@/interfaces/watchlist'

export const getWatchlists = async () => { ... }
```

**Benefits:**
- ✅ No duplication
- ✅ Single source of truth
- ✅ Easier to maintain
- ✅ Better IDE autocomplete

---

## 🔗 Related Documentation

- [CODING_RULES.md](./CODING_RULES.md) - Overall coding standards
- [TYPESCRIPT_CONVENTION.md](../TYPESCRIPT_CONVENTION.md) - TypeScript naming
- [AXIOS_USAGE.md](./AXIOS_USAGE.md) - API client usage
