<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="brand">
        <span class="brand-mark">Pulse</span>
        <span class="brand-subtitle">Network operations</span>
      </div>
      <nav class="nav-links">
        <NuxtLink to="/">Dashboard</NuxtLink>
        <NuxtLink to="/devices">Devices</NuxtLink>
        <NuxtLink to="/services">Services</NuxtLink>
        <NuxtLink to="/tunnels">Tunnels</NuxtLink>
        <NuxtLink to="/timeline">Timeline</NuxtLink>
        <NuxtLink to="/settings">Settings</NuxtLink>
      </nav>
    </header>
    <main class="app-main">
      <slot />
    </main>
  </div>
</template>
