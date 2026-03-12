# Authentication Module Documentation

## 📋 Overview

Authentication module handles user login and session management using JWT tokens.

**⚠️ Note:** This module only includes endpoints that exist in the backend API documentation (`be/docs/API_DOCUMENTATION.md`).

---

## 🏗️ Architecture

```
src/
├── interfaces/
│   └── auth.ts              # Auth interfaces (ILoginRequest, IUser)
├── types/
│   └── auth.ts              # Auth types (TLoginFormData)
├── services/
│   └── auth.service.ts      # API calls (login, logout)
├── stores/
│   └── auth.store.ts        # Pinia store (state management)
├── router/
│   └── index.ts             # Router with auth guards
└── pages/
    ├── LoginPage.vue        # Login page (placeholder)
    └── HomePage.vue         # Home page (placeholder)
```

---

## 🔐 Available Endpoints

Based on backend API documentation:

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/login` | Login user | ❌ |
| POST | `/auth/logout` | Logout user (optional) | ✅ |

### ⚠️ Endpoints NOT Included

The following endpoints are **NOT** implemented because they don't exist in the backend API docs:

- ❌ `/auth/register` - Register new user
- ❌ `/auth/me` - Get current user
- ❌ `/auth/change-password` - Change password
- ❌ `/auth/forgot-password` - Request password reset
- ❌ `/auth/reset-password` - Reset password

If you need these endpoints, please add them to the backend API first (`be/docs/API_DOCUMENTATION.md`).

---

## 📝 Interfaces

### **ILoginRequest**
```typescript
interface ILoginRequest {
  email: string
  password: string
}
```

### **ILoginResponse**
```typescript
interface ILoginResponse {
  token: string
  user?: IUser  // Optional, depends on backend response
}
```

### **IUser**
```typescript
interface IUser {
  id: number
  email: string
  name?: string
  created_at?: string
}
```

---

## 🚀 Usage Examples

### **1. Login**

```typescript
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()

// Login with email & password
const success = await authStore.loginAction('user@example.com', 'password123')

if (success) {
  // User is logged in
  console.log('Is authenticated:', authStore.isAuthenticated)
  
  // Redirect to dashboard
  router.push('/dashboard')
}
```

### **2. Logout**

```typescript
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()

// Logout
await authStore.logoutAction()

// Redirect to login
router.push('/login')
```

### **3. Check Authentication Status**

```typescript
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()

if (authStore.isAuthenticated) {
  // User is logged in
  console.log('User:', authStore.user)
} else {
  // User is not logged in
  console.log('Please login')
}
```

---

## 🛡️ Router Guards

### **Protected Routes**

Add `meta: { requiresAuth: true }` to routes that require authentication:

```typescript
{
  path: '/dashboard',
  name: 'dashboard',
  component: () => import('@/pages/DashboardPage.vue'),
  meta: { requiresAuth: true }
}
```

### **Guest Routes**

Add `meta: { guest: true }` to routes only for non-authenticated users:

```typescript
{
  path: '/login',
  name: 'login',
  component: () => import('@/pages/LoginPage.vue'),
  meta: { guest: true }
}
```

---

## 🗄️ Store Structure

### **State**

```typescript
{
  user: IUser | null,          // Current user data (if available)
  loading: boolean,            // Loading state
  error: string | null,        // Error message
  isAuthenticated: boolean     // Computed: based on token existence
}
```

### **Getters (Computed)**

```typescript
userEmail: string      // user.value?.email ?? ''
userName: string       // user.value?.name ?? user.value?.email ?? 'User'
userId: number | null  // user.value?.id ?? null
```

### **Actions**

```typescript
loginAction(email, password, remember)     // Login user
logoutAction()                             // Logout user
clearUserState()                           // Clear state
```

---

## 🔑 Token Management

### **Storage**

Tokens are stored in `localStorage`:

```typescript
// src/lib/axios.ts
const TOKEN_KEY = 'auth_token'

export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token)
}

export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY)
}

export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
}
```

### **Auto-Attach to Requests**

```typescript
// Axios request interceptor
apiClient.interceptors.request.use((config) => {
  const token = getToken()
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

---

## 💡 Best Practices

### **1. Use Store Actions**

```typescript
// ✅ GOOD
const authStore = useAuthStore()
await authStore.loginAction(email, password)

// ❌ BAD: Direct API call
import authApi from '@/services/auth.service'
await authApi.login({ email, password })
```

### **2. Handle Loading States**

```vue
<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store'

const authStore = useAuthStore()

const handleLogin = async () => {
  if (authStore.loading) return // Prevent double submit
  
  const success = await authStore.loginAction(email, password)
  if (success) {
    router.push('/dashboard')
  }
}
</script>

<template>
  <button :disabled="authStore.loading">
    {{ authStore.loading ? 'Logging in...' : 'Login' }}
  </button>
</template>
```

### **3. Protect Routes**

```typescript
// ✅ Add meta to protected routes
{
  path: '/dashboard',
  component: Dashboard,
  meta: { requiresAuth: true }
}
```

---

## 📚 Related Documentation

| Document | Description |
|----------|-------------|
| [CODING_RULES.md](./CODING_RULES.md) | Overall coding standards |
| [INTERFACES_AND_TYPES.md](./INTERFACES_AND_TYPES.md) | Interface & type structure |
| [AXIOS_USAGE.md](./AXIOS_USAGE.md) | API client configuration |

---

## 🔗 Backend API Reference

See: [`be/docs/API_DOCUMENTATION.md`](../../be/docs/API_DOCUMENTATION.md)

---

## 🎯 Next Steps

To complete the authentication feature:

1. **Implement Login Page UI**
   - Create `src/pages/LoginPage.vue` with form
   - Use `useAuthStore().loginAction()`
   - Add form validation with Vuelidate

2. **Add Protected Routes**
   - Add `meta: { requiresAuth: true }` to dashboard routes
   - Implement dashboard page

3. **Add Logout Button**
   - Add to navigation/header
   - Call `useAuthStore().logoutAction()`

---

## ⚠️ Important Notes

1. **Only login/logout endpoints are implemented** - Other auth endpoints (register, change password, etc.) are not available in the backend API documentation.

2. **User data is optional** - The backend may or may not return user data in the login response. The store handles both cases.

3. **Token-based authentication** - All authenticated requests require `Authorization: Bearer <token>` header, which is automatically added by axios interceptor.

4. **To add more auth endpoints** - First add them to the backend API documentation (`be/docs/API_DOCUMENTATION.md`), then update this module.
