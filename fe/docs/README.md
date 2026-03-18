# Frontend Documentation Index

Dokumentasi lengkap untuk frontend Trading Bot v4 (Vue 3 + TypeScript).

## 📚 Documentation Files

| File | Description |
|------|-------------|
| [AUTH_MODULE.md](./AUTH_MODULE.md) | Authentication module & JWT token management |
| [AXIOS_USAGE.md](./AXIOS_USAGE.md) | API client configuration & usage |
| [CODING_RULES.md](./CODING_RULES.md) | Frontend coding standards & conventions |
| [INTERFACES_AND_TYPES.md](./INTERFACES_AND_TYPES.md) | TypeScript interfaces & types structure |
| [SWEETALERT_USAGE.md](./SWEETALERT_USAGE.md) | SweetAlert2 usage guide |
| [TAILWIND_CONFIG.md](./TAILWIND_CONFIG.md) | TailwindCSS v4 configuration |
| [TYPESCRIPT_CONVENTION.md](./TYPESCRIPT_CONVENTION.md) | TypeScript naming conventions |

## 🚀 Quick Start

### Installation

```bash
# Install dependencies
yarn install

# Copy environment file
cp .env.example .env

# Run dev server
yarn dev
```

### Build for Production

```bash
# Build
yarn build

# Preview production build
yarn preview
```

### Linting

```bash
# ESLint
yarn lint

# Format with Prettier
yarn format
```

## 📋 Key Topics

### 1. Authentication

Lihat [AUTH_MODULE.md](./AUTH_MODULE.md) untuk:
- Login/logout flow
- JWT token management
- Protected routes
- Store usage patterns

### 2. API Communication

Lihat [AXIOS_USAGE.md](./AXIOS_USAGE.md) untuk:
- Axios configuration
- Request/Response interceptors
- Error handling
- TypeScript integration

### 3. Coding Standards

Lihat [CODING_RULES.md](./CODING_RULES.md) untuk:
- Project structure
- Naming conventions
- Vue component structure
- State management (Pinia)
- Form validation (Vuelidate)
- UI components usage
- Phosphor Icons

### 4. TypeScript

Lihat [TYPESCRIPT_CONVENTION.md](./TYPESCRIPT_CONVENTION.md) dan [INTERFACES_AND_TYPES.md](./INTERFACES_AND_TYPES.md) untuk:
- Interface vs Type usage
- Naming conventions (I prefix, T prefix)
- Generic types
- File organization

### 5. Styling

Lihat [TAILWIND_CONFIG.md](./TAILWIND_CONFIG.md) untuk:
- Custom theme colors
- Trading signal colors
- Container utilities
- Dark mode

### 6. UI Components

**Reusable Components** (`src/components/common/`):
- `UiInput` - Input field dengan label & error handling
- `UiButton` - Button dengan variants (primary, danger, outline)
- `UiPassword` - Password input dengan show/hide toggle

**Icons** - Phosphor Icons Vue (50,000+ icons):
```vue
<script setup lang="ts">
import { PhPlus, PhTrash, PhPencilSimple } from '@phosphor-icons/vue'
</script>

<template>
  <PhPlus :size="20" weight="bold" class="text-blue-500" />
</template>
```

## 🏗️ Architecture Overview

```
src/
├── components/
│   ├── common/          # Reusable UI components
│   └── features/        # Feature-specific components
├── lib/
│   ├── axios.ts         # API client + IApiResponse<T>
│   ├── storage.ts       # LocalStorage utilities
│   ├── sweetalert.ts    # SweetAlert2 wrapper
│   └── validation.ts    # Validation helper
├── stores/              # Pinia stores (form state + validation)
├── services/            # API services
├── pages/               # Page-level components
└── router/              # Vue Router config
```

## 📝 Common Patterns

### Form State & Validation di Store

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

### Component Menggunakan Store

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

## 🔗 Related Documentation

- [Backend API Documentation](../../be/docs/API_DOCUMENTATION.md)
- [Backend Coding Rules](../../be/docs/CODING_RULES.md)
- [Project README](../../README.md)
