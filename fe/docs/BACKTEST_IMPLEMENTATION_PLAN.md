# Backtest Feature - Implementation Plan

## 📋 Overview

Implementasi fitur Backtest untuk menjalankan strategy trading pada data historis dan melihat hasilnya secara detail.

**Timeline:** 3-4 hari kerja
**Priority:** High
**Status:** Planning

---

## 🎯 Goals

1. ✅ Create backtest dengan form input (name, symbol, strategy, days, capital)
2. ✅ List semua backtest dengan status monitoring
3. ✅ Detail backtest dengan chart (OHLCV + Equity Curve + Trade markers)
4. ✅ Delete backtest
5. ✅ Real-time status polling (PENDING → RUNNING → COMPLETED)

---

## 📦 Deliverables

### 1. **Interfaces & Types** (`src/interfaces/backtest.ts`)

```typescript
// Request
interface IBacktestRequest {
  name: string
  symbol: string
  strategy_id: number
  days: number
  capital: number
}

// Response - Main
interface IBacktestResponse {
  id: number
  name: string
  symbol: string
  strategy_id: number
  start_time: string
  end_time: string
  capital: number
  summary: IBacktestSummary
  equity_curve: IEquityPoint[]
  ohlcv: ICandleData[]
  trades: IBacktestTradeDTO[]
  status: TBacktestStatus
  error_message: string
  created_at: string
  completed_at: string | null
  strategy: IStrategyData | null
}

// Summary
interface IBacktestSummary {
  initial_balance: number
  final_balance: number
  net_profit: number
  net_profit_percent: number
  win_rate_pct: number
  total_trades: number
  winning_trades: number
  losing_trades: number
  expired_trades: number
  cancelled_trades: number
  max_drawdown_pct: number
  profit_factor: number
  avg_win: number
  avg_loss: number
  largest_win: number
  largest_loss: number
}

// Equity Curve
interface IEquityPoint {
  timestamp: number
  balance: number
  pnl: number
}

// OHLCV
interface ICandleData {
  timestamp: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

// Trade
interface IBacktestTradeDTO {
  trade_id: number
  trade_num: number
  side: 'BUY' | 'SELL'
  signal: string
  confidence: number
  trading_mode: string
  status: string
  targets: ITradeTargets
  entries: ITradeEntry[]
  exit: ITradeExit | null
  total_qty: number
  avg_entry_price: number
  total_capital: number
  pnl: number
  pnl_percent: number
  entry_time: string
  filled_time: string | null
  exit_time: string | null
  duration_minutes: number
  daily_stats: IDailyStatsSnapshot | null
}

// Trade sub-types
interface ITradeTargets {
  tp_price: number
  sl_price: number
  ratio: number
}

interface ITradeEntry {
  entry_num: number
  type: 'MARKET' | 'LIMIT'
  price: number
  qty: number
  timestamp: number
  status: string
}

interface ITradeExit {
  reason: string
  price: number
  timestamp: number
}

interface IDailyStatsSnapshot {
  trade_count: number
  pnl: number
  consecutive_loss: number
}

// List Item
interface IBacktestListItem {
  id: number
  name: string
  symbol: string
  strategy_name: string
  total_pnl: number
  total_pnl_percent: number
  win_rate: number
  total_trades: number
  status: TBacktestStatus
  created_at: string
}

// Types
type TBacktestStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED'
```

---

### 2. **API Service** (`src/services/backtest.service.ts`)

```typescript
import api from '@/lib/axios'
import type {
  IBacktestRequest,
  IBacktestResponse,
  IBacktestListItem
} from '@/interfaces/backtest'

export const backtestService = {
  /**
   * Create new backtest
   */
  async create(data: IBacktestRequest): Promise<IBacktestResponse> {
    const response = await api.post<IApiResponse<IBacktestResponse>>('/backtest', data)
    return response.data.data
  },

  /**
   * Get all backtests (list view)
   */
  async getAll(): Promise<IBacktestListItem[]> {
    const response = await api.get<IApiResponse<IBacktestListItem[]>>('/backtest')
    return response.data.data
  },

  /**
   * Get backtest by ID (detail view with OHLCV, equity curve, trades)
   */
  async getById(id: number): Promise<IBacktestResponse> {
    const response = await api.get<IApiResponse<IBacktestResponse>>(`/backtest/${id}`)
    return response.data.data
  },

  /**
   * Delete backtest
   */
  async delete(id: number): Promise<void> {
    await api.delete(`/backtest/${id}`)
  }
}
```

---

### 3. **Pinia Store** (`src/stores/backtest.store.ts`)

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { backtestService } from '@/services/backtest.service'
import type {
  IBacktestRequest,
  IBacktestResponse,
  IBacktestListItem
} from '@/interfaces/backtest'
import { useVuelidate } from '@vuelidate/core'
import { required, minLength, maxLength, numeric, between } from '@vuelidate/validators'
import { sweetalert } from '@/lib/sweetalert'

export const useBacktestStore = defineStore('backtest', () => {
  // State
  const backtests = ref<IBacktestListItem[]>([])
  const currentBacktest = ref<IBacktestResponse | null>(null)
  const isLoading = ref(false)
  const pollingInterval = ref<number | null>(null)

  // Form State - Create
  const createReq = ref<IBacktestRequest>({
    name: '',
    symbol: 'BTCUSDT',
    strategy_id: 0,
    days: 30,
    capital: 1000
  })

  const createRules = ref({
    name: { required, minLength: minLength(3), maxLength: maxLength(100) },
    symbol: { required, minLength: minLength(3) },
    strategy_id: { required, numeric },
    days: { required, numeric, between: between(1, 30) },
    capital: { required, numeric, between: between(10, 1000000) }
  })

  const createReqValid = useVuelidate(createRules, createReq)

  // Getters
  const getBacktestById = computed(() => (id: number) => 
    backtests.value.find(bt => bt.id === id)
  )

  // Actions
  async function fetchBacktests() {
    isLoading.value = true
    try {
      backtests.value = await backtestService.getAll()
    } catch (error) {
      sweetalert.error('Failed to load backtests')
      throw error
    } finally {
      isLoading.value = false
    }
  }

  async function createBacktest(): Promise<boolean> {
    const valid = await createReqValid.value.$validate()
    if (!valid) return false

    isLoading.value = true
    try {
      const result = await backtestService.create(createReq.value)
      await fetchBacktests()
      
      sweetalert.success('Backtest created', `Backtest "${result.name}" is running...`)
      
      // Start polling for status
      startPolling(result.id)
      
      return true
    } catch (error) {
      sweetalert.error('Failed to create backtest')
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function fetchBacktestDetail(id: number) {
    isLoading.value = true
    try {
      currentBacktest.value = await backtestService.getById(id)
    } catch (error) {
      sweetalert.error('Failed to load backtest detail')
      throw error
    } finally {
      isLoading.value = false
    }
  }

  async function deleteBacktest(id: number): Promise<boolean> {
    const confirm = await sweetalert.confirm(
      'Delete Backtest',
      'Are you sure you want to delete this backtest?',
      'warning'
    )

    if (!confirm) return false

    try {
      await backtestService.delete(id)
      await fetchBacktests()
      sweetalert.success('Backtest deleted successfully')
      return true
    } catch (error) {
      sweetalert.error('Failed to delete backtest')
      return false
    }
  }

  function startPolling(id: number) {
    // Clear existing interval
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value)
    }

    // Poll every 2 seconds
    pollingInterval.value = window.setInterval(async () => {
      try {
        const detail = await backtestService.getById(id)
        
        if (detail.status === 'COMPLETED' || detail.status === 'FAILED') {
          stopPolling()
          currentBacktest.value = detail
          
          if (detail.status === 'COMPLETED') {
            sweetalert.success('Backtest completed!', `Net Profit: $${detail.summary.net_profit}`)
          } else {
            sweetalert.error('Backtest failed', detail.error_message)
          }
        }
      } catch (error) {
        console.error('Polling error:', error)
      }
    }, 2000)
  }

  function stopPolling() {
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value)
      pollingInterval.value = null
    }
  }

  function resetForm() {
    createReq.value = {
      name: '',
      symbol: 'BTCUSDT',
      strategy_id: 0,
      days: 30,
      capital: 1000
    }
    createReqValid.value.$reset()
  }

  return {
    // State
    backtests,
    currentBacktest,
    isLoading,
    createReq,
    createReqValid,
    
    // Getters
    getBacktestById,
    
    // Actions
    fetchBacktests,
    createBacktest,
    fetchBacktestDetail,
    deleteBacktest,
    startPolling,
    stopPolling,
    resetForm
  }
})
```

---

### 4. **Pages Structure**

```
src/pages/
├── backtest/
│   ├── BacktestListPage.vue      # List all backtests
│   ├── BacktestDetailPage.vue    # Detail view with charts
│   └── BacktestCreateModal.vue   # Modal for creating backtest
```

---

### 5. **Router Configuration** (`src/router/index.ts`)

```typescript
{
  path: '/backtest',
  name: 'backtest',
  children: [
    {
      path: '',
      name: 'backtest-list',
      component: () => import('@/pages/backtest/BacktestListPage.vue'),
      meta: { title: 'Backtest', requiresAuth: true }
    },
    {
      path: ':id',
      name: 'backtest-detail',
      component: () => import('@/pages/backtest/BacktestDetailPage.vue'),
      meta: { title: 'Backtest Detail', requiresAuth: true }
    }
  ]
}
```

---

### 6. **Components Breakdown**

#### **BacktestListPage.vue**
- Table/List of all backtests
- Status badge (PENDING/RUNNING/COMPLETED/FAILED)
- Quick stats (PnL, Win Rate, Total Trades)
- Actions: View Detail, Delete
- Create button → opens modal

#### **BacktestCreateModal.vue**
- Form with validation
- Fields:
  - Name (text input)
  - Symbol (text input with autocomplete)
  - Strategy (dropdown from strategies API)
  - Days (number input, 1-30)
  - Capital (number input, min 10)
- Submit → Create & start polling

#### **BacktestDetailPage.vue**
- **Header Section:**
  - Backtest name, symbol, strategy
  - Status badge with polling indicator
  - Time range, capital
  
- **Summary Cards:**
  - Net Profit (USDT + %)
  - Win Rate (%)
  - Total Trades
  - Profit Factor
  - Max Drawdown (%)
  - Avg Win/Loss

- **Charts Section:**
  - OHLCV Chart (candlestick)
  - Equity Curve (line chart overlay or separate)
  - Trade markers on chart (entry/exit points)

- **Trades Table:**
  - List all trades with expandable details
  - Entry/Exit info
  - PnL per trade
  - Duration

---

## 🎨 UI/UX Considerations

### **Color Coding**
- **PENDING:** Gray/Yellow
- **RUNNING:** Blue with spinner
- **COMPLETED:** Green
- **FAILED:** Red

### **Profit/Loss Colors**
- **Positive PnL:** Green (`text-green-500`)
- **Negative PnL:** Red (`text-red-500`)

### **Loading States**
- Skeleton loader for cards
- Spinner for polling status
- Progress bar for batch fetching

### **Responsive Design**
- Mobile: Stack cards vertically
- Desktop: Grid layout for summary cards
- Chart: Full width with fixed height

---

## 📊 Chart Implementation Plan

### **Recommended Library:** Lightweight Charts (TradingView)
- Free & open source
- Lightweight (~50KB gzipped)
- Professional financial charts
- Supports candlestick + line + markers

### **Installation:**
```bash
yarn add lightweight-charts
```

### **Chart Components:**
1. **Candlestick Series** - OHLCV data
2. **Line Series** - Equity curve (overlay or separate)
3. **Markers** - Trade entry/exit points

---

## 🔧 Dependencies to Add

```json
{
  "dependencies": {
    "lightweight-charts": "^4.1.0"
  }
}
```

---

## 📝 Implementation Steps

### **Day 1: Setup & List**
1. Create interfaces (`backtest.ts`)
2. Create API service (`backtest.service.ts`)
3. Create Pinia store (`backtest.store.ts`)
4. Create BacktestListPage.vue
5. Add router

### **Day 2: Create Modal**
1. Create BacktestCreateModal.vue
2. Add form validation
3. Integrate with strategy dropdown
4. Test create & polling

### **Day 3: Detail Page**
1. Create BacktestDetailPage.vue
2. Install lightweight-charts
3. Create chart component
4. Display summary cards
5. Trades table

### **Day 4: Polish & Testing**
1. Error handling
2. Loading states
3. Responsive design
4. Testing all flows
5. Documentation

---

## ✅ Acceptance Criteria

- [ ] Can create backtest with valid form
- [ ] Form validation works correctly
- [ ] Backtest list shows all backtests
- [ ] Status updates in real-time (polling)
- [ ] Detail page shows OHLCV chart
- [ ] Equity curve displayed
- [ ] Trade markers visible on chart
- [ ] Summary cards show correct data
- [ ] Delete backtest works
- [ ] Error handling in place
- [ ] Responsive on mobile

---

## 🚀 Future Enhancements

1. **Compare Backtests** - Side-by-side comparison
2. **Export Results** - PDF/CSV export
3. **Advanced Filters** - Filter by strategy, date range, PnL
4. **Batch Backtest** - Run multiple backtests at once
5. **Optimization** - Strategy parameter optimization

---

## 📚 References

- [Backend API Docs](../../be/docs/BACKTEST_API.md)
- [Backend Flow](../../be/docs/TRADE_BACKTEST_FLOW.md)
- [Frontend Coding Rules](./CODING_RULES.md)
- [Lightweight Charts Docs](https://github.com/tradingview/lightweight-charts)
