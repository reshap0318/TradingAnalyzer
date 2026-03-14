<script setup lang="ts">
import { ref } from 'vue'
import AppSidebar from '@/components/common/AppSidebar.vue'
import AppHeader from '@/components/common/AppHeader.vue'

// Mobile sidebar state
const isSidebarOpen = ref(false)

const closeSidebar = () => {
  isSidebarOpen.value = false
}
</script>

<template>
  <div class="layout">
    <!-- Desktop Sidebar -->
    <div class="layout__sidebar-desktop">
      <AppSidebar />
    </div>

    <!-- Mobile Sidebar Overlay -->
    <Transition name="overlay">
      <div v-if="isSidebarOpen" class="layout__overlay" @click="closeSidebar"></div>
    </Transition>

    <!-- Mobile Sidebar Drawer -->
    <Transition name="drawer">
      <div v-if="isSidebarOpen" class="layout__sidebar-mobile">
        <div class="sidebar-mobile-header">
          <h1 class="text-xl font-bold text-primary">TradingAnalyzer</h1>
          <button class="close-btn" @click="closeSidebar" aria-label="Close menu">
            <PhX :size="24" />
          </button>
        </div>
        <AppSidebar @close="closeSidebar" />
      </div>
    </Transition>

    <!-- Main Content Area -->
    <div class="layout__content">
      <!-- Header -->
      <AppHeader @menu-toggle="isSidebarOpen = true">
        <template #title>
          <slot name="header-title">Dashboard</slot>
        </template>
      </AppHeader>

      <!-- Page Content -->
      <main class="layout__main">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  background-color: #f9fafb;
}

.layout__sidebar-desktop {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  z-index: 40;
}

.layout__overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 50;
}

.layout__sidebar-mobile {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  width: 256px;
  z-index: 60;
  background-color: white;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.1);
}

.sidebar-mobile-header {
  display: none;
}

.layout__content {
  margin-left: 256px;
  min-height: 100vh;
}

.layout__content {
  margin-left: 256px;
  min-height: 100vh;
}

.layout__main {
  padding: 80px 2rem 2rem 2rem;
  min-height: 100vh;
}

/* Overlay Transitions */
.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.3s ease;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

/* Drawer Transitions */
.drawer-enter-active,
.drawer-leave-active {
  transition: transform 0.3s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  transform: translateX(-100%);
}

/* Mobile Styles */
@media (max-width: 768px) {
  .layout__sidebar-desktop {
    display: none;
  }

  .layout__content {
    margin-left: 0;
  }

  .sidebar-mobile-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem;
    background: transparent;
    border: none;
    cursor: pointer;
    color: #6b7280;
    border-radius: 0.375rem;
    transition: all 0.2s ease;
  }

  .close-btn:hover {
    background-color: #f3f4f6;
    color: #111827;
  }

  /* Hide desktop sidebar in mobile drawer */
  .layout__sidebar-mobile :deep(.sidebar__logo) {
    display: none;
  }
}
</style>
