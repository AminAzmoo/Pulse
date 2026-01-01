<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-title">Pulse</span>
        <span class="brand-subtitle">Network Console</span>
      </div>
      <nav class="nav">
        <NuxtLink v-for="item in navItems" :key="item.to" :to="item.to" class="nav-link">
          {{ item.label }}
        </NuxtLink>
      </nav>
    </aside>
    <div class="main">
      <header class="topbar">
        <div>
          <p class="eyebrow">Pulse overview</p>
          <h1 class="page-title">{{ pageTitle }}</h1>
          <p class="subtle">Data source: Offline · Metrics show N/A or Unknown</p>
        </div>
        <div class="status-pill">Status: Offline</div>
      </header>
      <main class="content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()

const navItems = [
  { label: 'Dashboard', to: '/dashboard' },
  { label: 'Devices', to: '/devices' },
  { label: 'Tunnels', to: '/tunnels' },
  { label: 'Services', to: '/services' },
  { label: 'Timeline', to: '/timeline' },
  { label: 'Settings', to: '/settings' }
]

const pageTitleMap: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/devices': 'Devices',
  '/tunnels': 'Tunnels',
  '/services': 'Services',
  '/timeline': 'Timeline',
  '/settings': 'Settings'
}

const pageTitle = computed(() => pageTitleMap[route.path] ?? 'Pulse')
</script>
