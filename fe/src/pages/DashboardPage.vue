<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { DefaultLayout } from '@/layouts'
import {
  PhClock,
  PhChartLineUp,
  PhScales,
  PhSlidersHorizontal,
  PhList,
  PhChalkboardTeacher,
  PhTarget,
  PhTrendUp
} from '@phosphor-icons/vue'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)

// Quick access menu items
const quickAccess = [
  {
    title: 'Timeframes',
    description: 'Manage trading timeframes',
    icon: PhClock,
    route: '/timeframes',
    color: 'blue'
  },
  {
    title: 'Indicators',
    description: 'Configure trading indicators',
    icon: PhChartLineUp,
    route: '/indicators',
    color: 'green'
  },
  {
    title: 'Thresholds',
    description: 'Set signal thresholds',
    icon: PhScales,
    route: '/thresholds',
    color: 'purple'
  },
  {
    title: 'Configs',
    description: 'System configurations',
    icon: PhSlidersHorizontal,
    route: '/configs',
    color: 'orange'
  },
  {
    title: 'Watchlists',
    description: 'Manage trading pairs',
    icon: PhList,
    route: '/watchlists',
    color: 'cyan'
  },
  {
    title: 'Strategies',
    description: 'Trading strategies',
    icon: PhChalkboardTeacher,
    route: '/strategies',
    color: 'pink'
  },
  {
    title: 'Scanner',
    description: 'Market scanner',
    icon: PhTarget,
    route: '/scanner',
    color: 'red'
  }
]

const colorClasses: Record<string, string> = {
  blue: 'bg-blue-500',
  green: 'bg-green-500',
  purple: 'bg-purple-500',
  orange: 'bg-orange-500',
  cyan: 'bg-cyan-500',
  pink: 'bg-pink-500',
  red: 'bg-red-500'
}

const handleNavigate = (route: string) => {
  router.push(route)
}
</script>

<template>
  <DefaultLayout>
    <template #header-title>Dashboard</template>

    <div class="dashboard">
      <!-- Welcome Section -->
      <div class="welcome-card">
        <div class="welcome-card__content">
          <h1 class="welcome-card__title">
            Welcome back, {{ user?.name || 'User' }}! 👋
          </h1>
          <p class="welcome-card__text">
            Manage your trading bot settings and monitor market signals from one place.
          </p>
        </div>
        <div class="welcome-card__icon">
          <PhTrendUp :size="48" weight="fill" />
        </div>
      </div>

      <!-- Quick Access Grid -->
      <div class="quick-access">
        <h2 class="section-title">Quick Access</h2>
        <div class="quick-access__grid">
          <button
            v-for="item in quickAccess"
            :key="item.title"
            class="quick-access-card"
            @click="handleNavigate(item.route)"
          >
            <div class="quick-access-card__icon" :class="colorClasses[item.color]">
              <component :is="item.icon" :size="24" weight="fill" color="white" />
            </div>
            <div class="quick-access-card__content">
              <h3 class="quick-access-card__title">{{ item.title }}</h3>
              <p class="quick-access-card__description">{{ item.description }}</p>
            </div>
            <PhTarget :size="20" class="quick-access-card__arrow" />
          </button>
        </div>
      </div>

      <!-- Stats Overview (Placeholder) -->
      <div class="stats-overview">
        <h2 class="section-title">Overview</h2>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-card__header">
              <span class="stat-card__label">Active Watchlists</span>
              <PhList :size="20" class="stat-card__icon" />
            </div>
            <div class="stat-card__value">--</div>
            <p class="stat-card__text">Trading pairs being monitored</p>
          </div>

          <div class="stat-card">
            <div class="stat-card__header">
              <span class="stat-card__label">Active Strategies</span>
              <PhChalkboardTeacher :size="20" class="stat-card__icon" />
            </div>
            <div class="stat-card__value">--</div>
            <p class="stat-card__text">Trading strategies configured</p>
          </div>

          <div class="stat-card">
            <div class="stat-card__header">
              <span class="stat-card__label">Scanner Status</span>
              <PhTarget :size="20" class="stat-card__icon" />
            </div>
            <div class="stat-card__value">--</div>
            <p class="stat-card__text">Market scanner activity</p>
          </div>

          <div class="stat-card">
            <div class="stat-card__header">
              <span class="stat-card__label">Timeframes</span>
              <PhClock :size="20" class="stat-card__icon" />
            </div>
            <div class="stat-card__value">--</div>
            <p class="stat-card__text">Configured timeframes</p>
          </div>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>

<style scoped>
.dashboard {
  max-width: 1280px;
  margin: 0 auto;
}

/* Welcome Card */
.welcome-card {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-radius: 1rem;
  padding: 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  box-shadow: 0 4px 6px rgba(59, 130, 246, 0.2);
}

.welcome-card__content {
  flex: 1;
}

.welcome-card__title {
  font-size: 1.875rem;
  font-weight: 700;
  color: white;
  margin-bottom: 0.5rem;
}

.welcome-card__text {
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.9);
  max-width: 600px;
}

.welcome-card__icon {
  color: rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Section Title */
.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 1.5rem;
}

/* Quick Access */
.quick-access {
  margin-bottom: 2rem;
}

.quick-access__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.quick-access-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  width: 100%;
}

.quick-access-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 6px rgba(59, 130, 246, 0.1);
  transform: translateY(-2px);
}

.quick-access-card__icon {
  width: 48px;
  height: 48px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.quick-access-card__content {
  flex: 1;
  min-width: 0;
}

.quick-access-card__title {
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 0.25rem;
}

.quick-access-card__description {
  font-size: 0.875rem;
  color: #6b7280;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quick-access-card__arrow {
  color: #9ca3af;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.quick-access-card:hover .quick-access-card__arrow {
  color: #3b82f6;
  transform: translateX(4px);
}

/* Stats Overview */
.stats-overview {
  margin-bottom: 2rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1rem;
}

.stat-card {
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.75rem;
  padding: 1.5rem;
  transition: all 0.2s ease;
}

.stat-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 6px rgba(59, 130, 246, 0.1);
}

.stat-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.stat-card__label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #6b7280;
}

.stat-card__icon {
  color: #9ca3af;
}

.stat-card__value {
  font-size: 2rem;
  font-weight: 700;
  color: #111827;
  margin-bottom: 0.25rem;
}

.stat-card__text {
  font-size: 0.875rem;
  color: #6b7280;
}

/* Responsive */
@media (max-width: 768px) {
  .welcome-card {
    flex-direction: column;
    text-align: center;
    gap: 1rem;
  }

  .welcome-card__title {
    font-size: 1.5rem;
  }

  .welcome-card__icon {
    margin: 0 auto;
  }

  .quick-access__grid {
    grid-template-columns: 1fr;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
