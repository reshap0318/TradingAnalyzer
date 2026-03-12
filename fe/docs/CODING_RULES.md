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
├── public/                  # Static assets (favicon, robots.txt)
├── src/
│   ├── assets/              # Static assets (images, styles)
│   │   ├── style/           # Global styles
│   │   │   └── main.css     # TailwindCSS + custom styles
│   │   └── *.svg            # Images, icons
│   ├── components/          # Reusable Vue components
│   │   ├── common/          # Shared components (Button, Input, Modal)
│   │   └── features/        # Feature-specific components
│   ├── helpers/             # Helper functions (optional)
│   ├── layouts/             # Layout components
│   ├── lib/                 # Library configurations + generic interfaces
│   │   ├── axios.ts         # Axios instance, interceptors, IApiResponse<T>
│   │   ├── storage.ts       # LocalStorage utilities
│   │   ├── sweetalert.ts    # SweetAlert2 wrapper
│   │   └── tailwind.ts      # Tailwind config
│   ├── pages/               # Page-level components (route-level)
│   ├── router/              # Vue Router configuration
│   ├── stores/              # Pinia stores + module-specific interfaces
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

> **Note:** Interfaces dan Types **tidak** dipisah di folder khusus, melainkan:
> - **Generic interfaces** (contoh: `IApiResponse<T>`) → didefinisikan di `src/lib/axios.ts`
> - **Module-specific interfaces** (contoh: `IUser`, `ILoginRequest`) → didefinisikan di dalam `src/stores/*.store.ts`

---

## 🛠️ Tech Stack

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
| **Pages** | PascalCase dengan suffix `Page` | `LoginPage.vue`, `HomePage.vue` |
| **Composables** | camelCase dengan `use` prefix | `useWatchlist.ts`, `useAuth.ts` |
| **Stores** | camelCase dengan `.store.ts` | `watchlist.store.ts`, `auth.store.ts` |
| **Helpers** | camelCase | `formatters.ts`, `validators.ts` |

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
interface IUser {
  id: number
  email: string
  name?: string
  created_at?: string
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
type TOrderStatus = 'PENDING' | 'ACTIVE' | 'CLOSED'
type TTimestamp = string
type TUserCreateInput = Omit<IUser, 'id' | 'created_at'>
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
  email: string
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

**Import dari lib:**
```typescript
import { IApiResponse } from '@/lib/axios'
import { IUser, ILoginRequest } from '@/stores/auth.store'
```

> **💡 Rationale:** Dengan menaruh interface di file yang menggunakannya, kita meningkatkan **cohesion** dan memudahkan maintenance. Perubahan API = edit 1 file saja.

---

## 🎨 Vue Component Structure

### **1. Single File Component (SFC) Order**

```vue
<script setup lang="ts">
// 1. Imports (Vue, libraries, interfaces, stores)
import { ref, computed, onMounted } from 'vue'
import { IUser } from '@/stores/auth.store'
import { useAuthStore } from '@/stores/auth.store'
import { showSuccess, showError } from '@/lib/sweetalert'

// 2. Props & Emits
interface ILoginPageProps {
  initialEmail?: string
  rememberMe?: boolean
}

const props = withDefaults(defineProps<ILoginPageProps>(), {
  initialEmail: '',
  rememberMe: true,
})

const emit = defineEmits<{
  loginSuccess: [email: string]
}>()

// 3. State
const email = ref(props.initialEmail)
const password = ref('')
const isSubmitting = ref(false)

// 4. Computed
const isValidEmail = computed(() => email.value.includes('@'))

// 5. Methods
const handleSubmit = async () => {
  if (!isValidEmail.value) return
  isSubmitting.value = true
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
// src/stores/auth.store.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, post } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'
import { destroySession, getToken, setToken } from '@/lib/storage'

const BASE_URL = '/auth'

// Module-specific interfaces (defined in store file)
interface ILoginRequest {
  email: string
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

export const useAuthStore = defineStore('auth', () => {
  // State
  const loading = ref(false)
  const user = ref<IUser | null>(null)
  const token = ref<string | null>(getToken())

  // Getters
  const isAuthenticated = computed(() => !!token.value)

  // Actions
  async function loginAction(param: ILoginRequest): Promise<boolean> {
    loading.value = true

    try {
      const response = await post<IApiResponse<ILoginResponse>>(`${BASE_URL}/login`, param)
      setToken(response.data.data.token)
      
      if (response.data.data.user) {
        user.value = response.data.data.user
      }
      
      showSuccess('Welcome back!', `Hello, ${response.data.data.user?.name || 'User'}`)
      return true
    } catch (err: any) {
      const error = err.response?.data?.message || 'Login failed'
      showError('Login Failed', error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function logoutAction(): Promise<void> {
    try {
      await post<IApiResponse<void>>(`${BASE_URL}/logout`).catch(() => {
        console.warn('Logout endpoint not available')
      })
    } finally {
      destroySession()
      user.value = null
      token.value = null
      showSuccess('Logged out', 'You have been successfully logged out.')
    }
  }

  return {
    // State
    user,
    token,
    loading,

    // Getters
    isAuthenticated,

    // Actions
    loginAction,
    logoutAction
  }
})
```

### **2. Store Usage in Components**

```typescript
// Option 1: Direct access (recommended for simple cases)
<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()

// Access state directly
const { user, token, loading } = authStore

// Access getters
const isAuthenticated = authStore.isAuthenticated

// Call actions
await authStore.loginAction({ email: 'test@example.com', password: 'password' })
</script>

// Option 2: Using storeToRefs (for reactivity in complex components)
<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store'
import { storeToRefs } from 'pinia'

const authStore = useAuthStore()

// Destructure with reactivity
const { user, token, loading } = storeToRefs(authStore)

// Actions still accessed directly
const { loginAction, logoutAction } = authStore
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

### **2. Formatting Rules**

- **Semi**: `false` (no semicolons)
- **Single Quote**: `true`
- **Tab Width**: `2`
- **Print Width**: `100`
- **Trailing Comma**: `none`

### **3. Error Handling**

```typescript
// ✅ GOOD: Specific error handling dengan SweetAlert
import { useAuthStore } from '@/stores/auth.store'
import { showError } from '@/lib/sweetalert'

const authStore = useAuthStore()

try {
  await authStore.loginAction({ email: 'test@example.com', password: 'password' })
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
  await authStore.loginAction(credentials)
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
<SignalTable :signals="signals" :loading="loading" />

<!-- Child -->
<script setup lang="ts">
import { ISignal } from '@/stores/signal.store'

defineProps<{ signals: ISignal[]; loading?: boolean }>()
</script>
```

**Child → Parent**: Use `emit`
```vue
<!-- Child -->
<script setup lang="ts">
const emit = defineEmits<{
  select: [id: number]
  delete: [id: number]
}>()

emit('select', 123)
</script>
```

**Sibling**: Use Pinia store atau events

### **2. Async Operations**

```typescript
// ✅ GOOD: Loading & error states
const loading = ref(false)
const error = ref<string | null>(null)

const fetchData = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await get<IApiResponse<ISignal[]>>('/signals')
    items.value = response.data.data
  } catch (err: any) {
    error.value = err.message
    showError('Error', err.message)
  } finally {
    loading.value = false
  }
}
```

### **3. Computed vs Methods**

```typescript
// ✅ GOOD: Use computed for cached values
const activeSignals = computed(() =>
  signals.value.filter(s => s.status === 'ACTIVE')
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

### **5. Store Pattern**

```typescript
// ✅ GOOD: Setup store di component
<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()
const { user, isAuthenticated } = authStore
</script>

// ✅ GOOD: Call actions dengan await
const handleLogin = async () => {
  const success = await authStore.loginAction(credentials)
  if (success) {
    router.push('/dashboard')
  }
}
</script>
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [AXIOS_USAGE.md](./AXIOS_USAGE.md) | Axios configuration & usage |
| [SWEETALERT_USAGE.md](./SWEETALERT_USAGE.md) | SweetAlert2 usage |
| [TAILWIND_CONFIG.md](./TAILWIND_CONFIG.md) | TailwindCSS setup |

---

## 🚀 Quick Start: Adding New Feature

### **1. Create Store with Interfaces**

```typescript
// src/stores/signal.store.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type IApiResponse, get, post } from '@/lib/axios'
import { showSuccess, showError } from '@/lib/sweetalert'

// Module-specific interfaces (defined in store)
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

export const useSignalStore = defineStore('signal', () => {
  // State
  const items = ref<ISignal[]>([])
  const loading = ref(false)

  // Getters
  const activeSignals = computed(() => items.value.filter(s => s.status === 'ACTIVE'))

  // Actions
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

  async function createSignal(data: ISignalRequest) {
    try {
      const response = await post<IApiResponse<ISignal>>('/signals', data)
      items.value.push(response.data.data)
      showSuccess('Success', 'Signal created')
    } catch (error: any) {
      showError('Error', error.message)
      throw error
    }
  }

  return {
    items,
    loading,
    activeSignals,
    fetchSignals,
    createSignal
  }
})
```

### **2. Create Page Component**

```vue
<!-- src/pages/SignalsPage.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSignalStore, type ISignal } from '@/stores/signal.store'

const signalStore = useSignalStore()
const signals = ref<ISignal[]>([])

onMounted(async () => {
  await signalStore.fetchSignals()
  signals.value = signalStore.items
})
</script>

<template>
  <div>
    <h1>Signals</h1>
    <ul>
      <li v-for="signal in signals" :key="signal.id">
        {{ signal.symbol }} - {{ signal.side }}
      </li>
    </ul>
  </div>
</template>
```

### **3. Add Route**

```typescript
// src/router/index.ts
const routes: RouteRecordRaw[] = [
  {
    path: '/signals',
    name: 'signals',
    component: () => import('@/pages/SignalsPage.vue'),
    meta: { requiresAuth: true }
  }
]
```

---

## ✅ Code Review Checklist

- [ ] TypeScript interfaces menggunakan prefix `I`
- [ ] Type aliases menggunakan prefix `T`
- [ ] Generic parameters: `T` + descriptive (`TData`, `TEntity`)
- [ ] Component props typed dengan interface
- [ ] Interfaces didefinisikan di store file (bukan folder terpisah)
- [ ] Generic interfaces di `src/lib/axios.ts`
- [ ] Error handling dengan try-catch
- [ ] Loading states untuk async operations
- [ ] SweetAlert untuk user feedback
- [ ] Axios interceptors untuk auth
- [ ] Pinia store untuk state management (API calls langsung di store)
- [ ] ESLint & Prettier pass
