<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  PhSquaresFour,
  PhGear,
  PhTrendUp,
  PhClock,
  PhChartLineUp,
  PhScales,
  PhSlidersHorizontal,
  PhList,
  PhChalkboardTeacher,
  PhCaretDown,
  PhRobot,
  PhCurrencyCircleDollar,
  PhFlask,
  PhTestTube
} from '@phosphor-icons/vue'

const route = useRoute()

// Menu structure
const menuItems = [
  {
    category: 'Dashboard',
    icon: PhSquaresFour,
    route: '/dashboard',
    children: []
  },
  {
    category: 'Settings',
    icon: PhGear,
    children: [
      { name: 'Timeframes', icon: PhClock, route: '/timeframes' },
      { name: 'Indicators', icon: PhChartLineUp, route: '/indicators' },
      { name: 'Thresholds', icon: PhScales, route: '/thresholds' },
      { name: 'Configs', icon: PhSlidersHorizontal, route: '/configs' }
    ]
  },
  {
    category: 'Trading',
    icon: PhTrendUp,
    children: [
      { name: 'Watchlists', icon: PhList, route: '/watchlists' },
      { name: 'Strategies', icon: PhChalkboardTeacher, route: '/strategies' },
      { name: 'Signal Analyze', icon: PhFlask, route: '/signal-analyze' },
      { name: 'Backtest', icon: PhTestTube, route: '/backtest' },
      { name: 'Bot Control', icon: PhRobot, route: '/bot-control' },
      { name: 'Trade History', icon: PhCurrencyCircleDollar, route: '/trades' }
    ]
  }
]

// State
const expandedCategories = ref<string[]>([])

// Toggle category expansion
const toggleCategory = (category: string) => {
  const index = expandedCategories.value.indexOf(category)
  if (index === -1) {
    expandedCategories.value.push(category)
  } else {
    expandedCategories.value.splice(index, 1)
  }
}

// Check if route is active
const isActiveRoute = (routePath: string) => {
  return route.path === routePath
}

// Check if category has active child
const hasActiveChild = (children: Array<{ route: string }>) => {
  return children.some(child => route.path === child.route)
}

// Check if category should be expanded
const isExpanded = (category: string, children: Array<{ route: string }>) => {
  // Always expand if has active child
  if (hasActiveChild(children)) {
    return true
  }
  // Otherwise check manual expansion
  return expandedCategories.value.includes(category)
}

// Emit for mobile close
const emit = defineEmits<{
  'close': []
}>()

const handleNavigation = () => {
  emit('close')
}
</script>

<template>
  <aside class="w-64 h-screen bg-white border-r border-gray-200 flex flex-col fixed left-0 top-0 z-40 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-50 hover:scrollbar-thumb-gray-400">
    <!-- Logo -->
    <div class="p-6 border-b border-gray-200">
      <h1 class="text-xl font-bold text-blue-600">TradingAnalyzer</h1>
    </div>

    <!-- Navigation -->
    <nav class="p-4 flex-1">
      <div v-for="item in menuItems" :key="item.category" class="mb-1">
        <!-- Parent Menu (with children) -->
        <template v-if="item.children && item.children.length > 0">
          <button
            class="w-full flex items-center justify-between px-6 py-3 text-gray-500 bg-transparent border-none cursor-pointer transition-all duration-200 rounded-lg
                   hover:bg-gray-50 hover:text-gray-900
                   focus:outline-none"
            :class="{
              'bg-gradient-to-r from-blue-50 to-transparent text-blue-600 border-r-4 border-blue-600': hasActiveChild(item.children)
            }"
            @click="toggleCategory(item.category)"
          >
            <div class="flex items-center gap-3">
              <component :is="item.icon" :size="20" :weight="hasActiveChild(item.children) ? 'fill' : 'regular'" />
              <span class="text-[15px] font-medium">{{ item.category }}</span>
            </div>
            <PhCaretDown
              :size="16"
              class="text-gray-400 transition-transform duration-200"
              :class="{ 'rotate-180': isExpanded(item.category, item.children) }"
            />
          </button>

          <!-- Submenu -->
          <Transition name="submenu">
            <div
              v-if="isExpanded(item.category, item.children)"
              class="mt-1 bg-gray-50 rounded-lg overflow-hidden"
            >
              <router-link
                v-for="child in item.children"
                :key="child.name"
                :to="child.route"
                class="flex items-center gap-3 px-6 py-2.5 text-sm text-gray-500 no-underline transition-all duration-200 relative hover:bg-gray-100 hover:text-gray-900"
                :class="{ 'bg-blue-50 text-blue-600': isActiveRoute(child.route) }"
                @click="handleNavigation"
              >
                <component :is="child.icon" :size="18" :weight="isActiveRoute(child.route) ? 'fill' : 'regular'" />
                <span class="font-medium flex-1">{{ child.name }}</span>
                <div
                  v-if="isActiveRoute(child.route)"
                  class="absolute left-4 top-1/2 -translate-y-1/2 w-[3px] h-5 bg-blue-600 rounded-full"
                ></div>
              </router-link>
            </div>
          </Transition>
        </template>

        <!-- Single Menu (no children) -->
        <template v-else>
          <router-link
            v-if="item.route"
            :to="item.route as string"
            class="w-full flex items-center gap-3 px-6 py-3 text-gray-500 no-underline transition-all duration-200 rounded-lg
                   hover:bg-gray-50 hover:text-gray-900"
            :class="{
              'bg-gradient-to-r from-blue-50 to-transparent text-blue-600 border-r-4 border-blue-600': isActiveRoute(item.route as string)
            }"
            @click="handleNavigation"
          >
            <component :is="item.icon" :size="20" :weight="isActiveRoute(item.route as string) ? 'fill' : 'regular'" />
            <span class="text-[15px] font-medium">{{ item.category }}</span>
          </router-link>
        </template>
      </div>
    </nav>
  </aside>
</template>

<style scoped>
/* Submenu Transitions - Only keep transition classes */
.submenu-enter-active,
.submenu-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.submenu-enter-from,
.submenu-leave-to {
  opacity: 0;
  max-height: 0;
}

.submenu-enter-to,
.submenu-leave-from {
  opacity: 1;
  max-height: 200px;
}
</style>
