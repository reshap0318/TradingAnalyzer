# Frontend Coding Rules & Standards

## 📋 Table of Contents

1. [Project Structure](#project-structure)
2. [Tech Stack](#tech-stack)
3. [Naming Conventions](#naming-conventions)
4. [TypeScript Guidelines](#typescript-guidelines)
5. [Vue Component Structure](#vue-component-structure)
6. [State Management (Pinia)](#state-management-pinia)
7. [API & Services](#api--services)
8. [Code Style](#code-style)
9. [Best Practices](#best-practices)

---

## 🏗️ Project Structure

```
fe/
├── docs/                    # Documentation
├── public/                  # Static assets
├── src/
│   ├── components/          # Reusable Vue components
│   │   ├── common/          # Shared components (Button, Input, Modal)
│   │   └── features/        # Feature-specific components
│   ├── composables/         # Reusable composables (useXxx)
│   ├── interfaces/          # TypeScript interfaces
│   │   ├── common.ts        # Generic interfaces (IApiResponse, IPaginated)
│   │   ├── watchlist.ts     # Module-specific interfaces
│   │   └── index.ts         # Barrel exports
│   ├── layouts/             # Layout components
│   ├── lib/                 # Library configurations
│   │   ├── axios.ts         # Axios instance & interceptors
│   │   ├── sweetalert.ts    # SweetAlert2 wrapper
│   │   └── tailwind.ts      # Tailwind config
│   ├── router/              # Vue Router configuration
│   ├── stores/              # Pinia stores
│   ├── types/               # TypeScript types (unions, utilities)
│   ├── utils/               # Utility functions
│   ├── views/               # Page components
│   ├── App.vue              # Root component
│   └── main.ts              # Application entry point
├── .env                     # Environment variables
├── .env.example             # Environment template
├── eslint.config.js         # ESLint configuration
├── package.json             # Dependencies
├── tailwind.config.js       # TailwindCSS config
├── tsconfig.json            # TypeScript config
└── vite.config.ts           # Vite configuration
```

---

## 🛠️ Tech Stack

| Technology | Purpose | Version |
|------------|---------|---------|
| **Vue 3** | Frontend Framework | 3.5+ |
| **TypeScript** | Type Safety | 5.9+ |
| **Vite** | Build Tool | 7.x |
| **Pinia** | State Management | 3.x |
| **Vue Router** | Routing | 4.x |
| **Axios** | HTTP Client | 1.x |
| **SweetAlert2** | Alerts/Modals | 11.x |
| **TailwindCSS** | Styling | 4.x |
| **ESLint** | Linting | 10.x |
| **Prettier** | Formatting | 3.x |
| **Yarn** | Package Manager | 4.x |

---

## 📝 Naming Conventions

### **1. TypeScript Naming**

| Construct | Pattern | Example |
|-----------|---------|---------|
| **Interface** | `I` + PascalCase | `IWatchlist`, `IApiResponse` |
| **Type** | `T` + PascalCase | `TSignalAction`, `TTimestamp` |
| **Generic Param** | `T` + descriptive | `TData`, `TEntity`, `TSource` |
| **Class** | PascalCase | `WatchlistService` |
| **Enum** | PascalCase | `SignalAction` |
| **Function** | camelCase | `getWatchlists`, `createWatchlist` |
| **Variable** | camelCase | `watchlistData`, `isLoading` |
| **Constant** | UPPER_CASE | `BASE_URL`, `TOKEN_KEY` |

### **2. File Naming**

| Type | Convention | Example |
|------|-----------|---------|
| **Vue Components** | PascalCase | `WatchlistTable.vue`, `AppHeader.vue` |
| **Composables** | camelCase with `use` prefix | `useWatchlist.ts`, `useAuth.ts` |
| **Stores** | camelCase with `.store.ts` | `watchlist.store.ts`, `auth.store.ts` |
| **Interfaces** | camelCase | `watchlist.ts`, `common.ts` |
| **Utils** | camelCase | `formatters.ts`, `validators.ts` |

### **3. Folder Naming**

- **Lowercase** dengan hyphen jika perlu: `components/`, `api-services/`
- **Kecuali**: `components/common/`, `components/features/` (subfolder)

---

## 💻 TypeScript Guidelines

### **1. Interface vs Type**

**Gunakan `interface` (prefix `I`) untuk:**
- Object shapes (DTOs, Models, API Responses)
- Service & Repository contracts
- Component props

```typescript
// ✅ GOOD
interface IWatchlist {
  id: number
  symbol: string
  is_active: boolean
  created_at: string
}

interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}
```

**Gunakan `type` (prefix `T`) untuk:**
- Union types
- Primitive aliases
- Mapped/utility types
- Tuple types

```typescript
// ✅ GOOD
type TSignalAction = 'BUY' | 'SELL' | 'WAIT'
type TTimestamp = string
type TWatchlistCreateInput = Omit<IWatchlist, 'id' | 'created_at'>
```

### **2. Generic Interfaces**

**WAJIB** menggunakan `T` + descriptive name:

```typescript
// ✅ GOOD
interface IApiResponse<TData> {
  code: number
  message: string
  data: TData
}

interface IRepository<TEntity> {
  findById(id: number): Promise<TEntity | null>
}

interface IMapper<TSource, TDestination> {
  map(source: TSource): TDestination
}

// ❌ BAD: Terlalu singkat
interface IApiResponse<T> { ... }
```

### **3. Interface Location**

**Common/Generic interfaces** → `src/interfaces/common.ts`
```typescript
export interface IApiResponse<TData> { ... }
export interface IPaginatedResponse<TData> { ... }
```

**Module-specific interfaces** → `src/interfaces/<module>.ts`
```typescript
// src/interfaces/watchlist.ts
export interface IWatchlist { ... }
export interface IWatchlistRequest { ... }
```

**Import dari common:**
```typescript
import { IApiResponse } from '@/interfaces/common'
import { IWatchlist } from '@/interfaces/watchlist'
```

---

## 🎨 Vue Component Structure

### **1. Single File Component (SFC) Order**

```vue
<script setup lang="ts">
// 1. Imports (Vue, libraries, interfaces, stores)
import { ref, computed, onMounted } from 'vue'
import { IWatchlist } from '@/interfaces/watchlist'
import { useWatchlistStore } from '@/stores/watchlist.store'
import { showSuccess, showError } from '@/lib/sweetalert'

// 2. Props & Emits
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

// 3. State
const selectedId = ref<number | null>(null)

// 4. Computed
const hasSelection = computed(() => selectedId.value !== null)

// 5. Methods
const handleSelect = (id: number) => {
  selectedId.value = id
  emit('select', id)
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

### **2. Component Naming**

- **PascalCase** untuk filename & component name
- **Multi-word** untuk root components (AppHeader, WatchlistTable)

```vue
<!-- ✅ GOOD -->
<!-- components/WatchlistTable.vue -->
<script setup lang="ts">
// Component logic
</script>

<!-- ❌ BAD: Single word for root components -->
<!-- components/Table.vue -->
```

---

## 🗄️ State Management (Pinia)

### **1. Store Structure**

```typescript
// src/stores/watchlist.store.ts
import { defineStore } from 'pinia'
import { get, post, put, del } from '@/lib/axios'
import { IWatchlist, IWatchlistRequest } from '@/interfaces/watchlist'
import { IApiResponse } from '@/interfaces/common'
import { showSuccess, showError } from '@/lib/sweetalert'

interface IWatchlistState {
  items: IWatchlist[]
  selected: IWatchlist | null
  loading: boolean
  error: string | null
}

export const useWatchlistStore = defineStore('watchlist', {
  // State
  state: (): IWatchlistState => ({
    items: [],
    selected: null,
    loading: false,
    error: null,
  }),

  // Getters
  getters: {
    activeWatchlists: (state) =>
      state.items.filter(w => w.is_active),

    hasItems: (state) =>
      state.items.length > 0,
  },

  // Actions
  actions: {
    async fetchWatchlists(): Promise<void> {
      this.loading = true
      this.error = null

      try {
        const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
        this.items = response.data
      } catch (error: any) {
        this.error = error.message || 'Failed to fetch watchlists'
        showError('Error', this.error)
      } finally {
        this.loading = false
      }
    },

    async createWatchlist(data: IWatchlistRequest): Promise<void> {
      try {
        const response = await post<IApiResponse<IWatchlist>>('/watchlists', data)
        this.items.push(response.data)
        showSuccess('Success', 'Watchlist created')
      } catch (error: any) {
        showError('Error', error.message)
        throw error
      }
    },
  },
})
```

### **2. Store Usage in Components**

```vue
<script setup lang="ts">
import { useWatchlistStore } from '@/stores/watchlist.store'

const watchlistStore = useWatchlistStore()

// Access state
const { items, loading, error } = storeToRefs(watchlistStore)

// Access actions
const { fetchWatchlists, createWatchlist } = watchlistStore

// Access getters
const activeWatchlists = watchlistStore.activeWatchlists
</script>
```

---

## 💅 Code Style

### **1. Import Order**

```typescript
// 1. Vue & libraries
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

// 2. Libraries (axios, sweetalert)
import axios from 'axios'
import Swal from 'sweetalert2'

// 3. Interfaces & Types
import { IWatchlist } from '@/interfaces/watchlist'
import { IApiResponse } from '@/interfaces/common'

// 4. Stores
import { useWatchlistStore } from '@/stores/watchlist.store'

// 5. Components (in Vue templates)
import WatchlistTable from '@/components/WatchlistTable.vue'
```

### **2. Formatting Rules**

- **Semi**: `false` (no semicolons)
- **Single Quote**: `true`
- **Tab Width**: `2`
- **Print Width**: `100`
- **Trailing Comma**: `none`

### **3. Error Handling**

```typescript
// ✅ GOOD: Specific error handling
try {
  await watchlistService.createWatchlist(data)
  showSuccess('Success', 'Watchlist created')
} catch (error: any) {
  if (error.response?.status === 400) {
    showError('Bad Request', error.response.data.message)
  } else if (error.response?.status === 401) {
    showError('Unauthorized', 'Please login again')
  } else {
    showError('Error', 'Something went wrong')
  }
}

// ❌ BAD: Empty catch
try {
  await watchlistService.createWatchlist(data)
} catch (error) {
  // Silent fail
}
```

---

## 🎯 Best Practices

### **1. Component Communication**

**Parent → Child**: Use `props`
```vue
<!-- Parent -->
<WatchlistTable :watchlists="watchlists" :loading="loading" />

<!-- Child -->
defineProps<{ watchlists: IWatchlist[]; loading?: boolean }>()
```

**Child → Parent**: Use `emit`
```vue
<!-- Child -->
const emit = defineEmits<{
  select: [id: number]
  delete: [id: number]
}>()

emit('select', 123)
```

**Sibling**: Use Pinia store or events

### **2. Async Operations**

```typescript
// ✅ GOOD: Loading & error states
const loading = ref(false)
const error = ref<string | null>(null)

const fetchData = async () => {
  loading.value = true
  error.value = null
  
  try {
    const response = await watchlistService.getWatchlists()
    items.value = response.data
  } catch (err: any) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
```

### **3. Computed vs Methods**

```typescript
// ✅ GOOD: Use computed for cached values
const activeWatchlists = computed(() => 
  watchlists.value.filter(w => w.is_active)
)

// ✅ GOOD: Use methods for actions
const handleSelect = (id: number) => {
  selectedId.value = id
}
```

### **4. Environment Variables**

```env
# .env
VITE_API_BASE_URL=http://localhost:8000/api
VITE_APP_TITLE=TradingAnalyzer
```

```typescript
// ✅ GOOD: Access with import.meta.env
const BASE_URL = import.meta.env.VITE_API_BASE_URL
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [TYPESCRIPT_CONVENTION.md](../TYPESCRIPT_CONVENTION.md) | TypeScript naming & usage |
| [AXIOS_USAGE.md](./AXIOS_USAGE.md) | Axios configuration |
| [SWEETALERT_USAGE.md](./SWEETALERT_USAGE.md) | SweetAlert2 usage |
| [TAILWIND_CONFIG.md](./TAILWIND_CONFIG.md) | TailwindCSS setup |

---

## 🚀 Quick Start: Adding New Feature

### **1. Create Interfaces**

```typescript
// src/interfaces/signal.ts
export interface ISignal {
  id: number
  symbol: string
  side: 'LONG' | 'SHORT'
  status: TOrderStatus
  created_at: string
}

export interface ISignalRequest {
  symbol: string
  side: 'LONG' | 'SHORT'
}
```

### **2. Create Store**

```typescript
// src/stores/signal.store.ts
import { defineStore } from 'pinia'
import { get, post } from '@/lib/axios'
import { IApiResponse } from '@/interfaces/common'
import { ISignal, ISignalRequest } from '@/interfaces/signal'
import { showSuccess, showError } from '@/lib/sweetalert'

export const useSignalStore = defineStore('signal', {
  state: () => ({
    items: [] as ISignal[],
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async fetchSignals() {
      this.loading = true
      this.error = null
      try {
        const response = await get<IApiResponse<ISignal[]>>('/signals')
        this.items = response.data
      } catch (error: any) {
        this.error = error.message
        showError('Error', error.message)
      } finally {
        this.loading = false
      }
    },
    async createSignal(data: ISignalRequest) {
      try {
        const response = await post<IApiResponse<ISignal>>('/signals', data)
        this.items.push(response.data)
        showSuccess('Success', 'Signal created')
      } catch (error: any) {
        showError('Error', error.message)
        throw error
      }
    },
  },
})
```

### **3. Create Component**

```vue
<!-- src/components/SignalTable.vue -->
<script setup lang="ts">
import { ISignal } from '@/interfaces/signal'

defineProps<{
  signals: ISignal[]
}>()
</script>

<template>
  <div>
    <!-- Template -->
  </div>
</template>
```

---

## ✅ Code Review Checklist

- [ ] TypeScript interfaces menggunakan prefix `I`
- [ ] Type aliases menggunakan prefix `T`
- [ ] Generic parameters: `T` + descriptive (`TData`, `TEntity`)
- [ ] Component props typed dengan interface
- [ ] Error handling dengan try-catch
- [ ] Loading states untuk async operations
- [ ] SweetAlert untuk user feedback
- [ ] Axios interceptors untuk auth
- [ ] Pinia store untuk state management (API calls langsung di store)
- [ ] ESLint & Prettier pass
