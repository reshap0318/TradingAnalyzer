# Axios Configuration Guide

## 📦 Setup

Axios sudah dikonfigurasi di `src/lib/axios.ts` dengan:
- ✅ Auto Authorization header dari localStorage
- ✅ Request/Response interceptors
- ✅ Error handling
- ✅ TypeScript support

## 🚀 Usage

### Basic Usage

```typescript
import apiClient, { get, post, put, del } from '@/lib/axios'

// Method 1: Using instance
const response = await apiClient.get('/watchlists')

// Method 2: Using helper functions
const response = await get('/watchlists')
```

### Authentication Flow

```typescript
import { setToken, removeToken } from '@/lib/axios'
import { post } from '@/lib/axios'

// Login
const login = async (email: string, password: string) => {
  const response = await post('/auth/login', { email, password })
  const token = response.data.data.token
  
  // Save token to localStorage
  setToken(token)
  
  // Future requests will auto-include Authorization header
}

// Logout
const logout = () => {
  removeToken()
  // Redirect to login
}
```

### API Methods

```typescript
import { get, post, put, patch, del, request } from '@/lib/axios'

// GET
const watchlists = await get('/watchlists')

// GET with params
const watchlist = await get('/watchlists/1', {
  params: { include: 'details' }
})

// POST
const created = await post('/watchlists', {
  symbol: 'BTCUSDT',
  is_active: true
})

// PUT
const updated = await put('/watchlists/1', {
  symbol: 'ETHUSDT',
  is_active: false
})

// PATCH
const partial = await patch('/watchlists/1', {
  is_active: false
})

// DELETE
const deleted = await del('/watchlists/1')

// Custom request
const custom = await request({
  method: 'GET',
  url: '/watchlists',
  headers: { 'Custom-Header': 'value' }
})
```

### With TypeScript

```typescript
import { get, post } from '@/lib/axios'
import { IApiResponse } from '@/interfaces/common'
import { IWatchlist } from '@/interfaces/watchlist'

// Typed response
const response = await get<IApiResponse<IWatchlist[]>>('/watchlists')
const watchlists: IWatchlist[] = response.data.data
```

### Error Handling

```typescript
import { post } from '@/lib/axios'
import { showError } from '@/lib/sweetalert'

try {
  const response = await post('/watchlists', { symbol: 'BTCUSDT' })
  // Handle success
} catch (error: any) {
  if (error.response) {
    // Server responded with error
    const { status, data } = error.response
    
    switch (status) {
      case 400:
        showError('Bad Request', data.message)
        break
      case 401:
        showError('Unauthorized', 'Please login again')
        break
      case 403:
        showError('Forbidden', 'Access denied')
        break
      case 404:
        showError('Not Found', 'Resource not found')
        break
      case 500:
        showError('Server Error', 'Please try again later')
        break
      default:
        showError('Error', data.message || 'Something went wrong')
    }
  } else if (error.request) {
    // Request made but no response
    showError('Network Error', 'Please check your connection')
  } else {
    // Other errors
    showError('Error', error.message)
  }
}
```

## 🔧 Configuration

### Environment Variables

Create `.env` file:

```env
VITE_API_BASE_URL=http://localhost:8000/api
```

### Runtime Configuration

```typescript
import { setBaseURL, setTimeout } from '@/lib/axios'

// Change base URL
setBaseURL('https://api.example.com/api')

// Change timeout
setTimeout(60000) // 60 seconds
```

## 📝 API Reference

| Function | Parameters | Returns | Description |
|----------|-----------|---------|-------------|
| `get` | `(url, config?)` | `Promise<AxiosResponse<T>>` | GET request |
| `post` | `(url, data?, config?)` | `Promise<AxiosResponse<T>>` | POST request |
| `put` | `(url, data?, config?)` | `Promise<AxiosResponse<T>>` | PUT request |
| `patch` | `(url, data?, config?)` | `Promise<AxiosResponse<T>>` | PATCH request |
| `del` | `(url, config?)` | `Promise<AxiosResponse<T>>` | DELETE request |
| `request` | `(config)` | `Promise<AxiosResponse<T>>` | Custom request |
| `getToken` | `()` | `string \| null` | Get token from storage |
| `setToken` | `(token)` | `void` | Set token to storage |
| `removeToken` | `()` | `void` | Remove token from storage |
| `setBaseURL` | `(url)` | `void` | Set base URL |
| `setTimeout` | `(timeout)` | `void` | Set timeout |

## 🔐 Authorization Flow

```
┌─────────────┐
│   Login     │
└─────┬───────┘
      │
      ▼
┌─────────────────┐
│ setToken(token) │
└─────┬───────────┘
      │
      ▼
┌─────────────────────┐
│ localStorage.setItem│
└─────┬───────────────┘
      │
      ▼
┌─────────────────────────┐
│ Auto-attach to requests │
│ Authorization: Bearer   │
└─────────────────────────┘
```

## 🎯 Best Practices

1. **Always use helper functions** for consistency
2. **Type your responses** with TypeScript
3. **Handle errors gracefully** with try-catch
4. **Use interceptors** for global logic
5. **Store tokens securely** (consider httpOnly cookies for production)

## 🔗 Documentation

- [Axios Official Docs](https://axios-http.com/)
- [Axios GitHub](https://github.com/axios/axios)
