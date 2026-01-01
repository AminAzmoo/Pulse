<script setup lang="ts">
const { statusLabel } = useApiStatus('/api/health');
</script>

<template>
  <section class="page-header">
    <h1 class="page-title">Pulse Dashboard</h1>
    <p class="page-subtitle">
      Real-time visibility across devices, tunnels, and services.
    </p>
  </section>

  <section class="card-grid">
    <div class="card">
      <span class="card-title">API Status</span>
      <span class="card-value">{{ statusLabel }}</span>
      <span class="status-pill">Health endpoint</span>
    </div>
    <div class="card">
      <span class="card-title">Active Devices</span>
      <span class="card-value">N/A</span>
      <span class="status-pill">Awaiting data source</span>
    </div>
    <div class="card">
      <span class="card-title">Active Tunnels</span>
      <span class="card-value">N/A</span>
      <span class="status-pill">Awaiting data source</span>
    </div>
    <div class="card">
      <span class="card-title">Service Incidents</span>
      <span class="card-value">N/A</span>
      <span class="status-pill">Awaiting data source</span>
    </div>
  </section>

  <section class="section">
    <h2 class="section-title">Operational Summary</h2>
    <p class="empty-state">
      Pulse will surface operational insights once connected to the production
      API.
    </p>
  </section>
</template>
