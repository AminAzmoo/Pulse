<script setup lang="ts">
useHead({ title: 'Pulse | Tunnels' });
</script>

<template>
  <section class="page-header">
    <h1 class="page-title">Tunnels</h1>
    <p class="page-subtitle">
      Secure tunnel status and routing coverage.
    </p>
  </section>

  <section class="section">
    <h2 class="section-title">Tunnel Overview</h2>
    <div class="card-grid">
      <div class="card">
        <span class="card-title">Active Tunnels</span>
        <span class="card-value">N/A</span>
      </div>
      <div class="card">
        <span class="card-title">Degraded Links</span>
        <span class="card-value">N/A</span>
      </div>
      <div class="card">
        <span class="card-title">Last Update</span>
        <span class="card-value">N/A</span>
      </div>
    </div>
    <p class="empty-state">No tunnel data is available yet.</p>
  </section>
</template>
