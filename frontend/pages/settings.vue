<template>
  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Pulse settings</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="card-grid">
      <div class="card">
        <p class="card-title">API endpoint</p>
        <div class="card-value">N/A</div>
        <p class="subtle">Configure your data source to enable live updates.</p>
      </div>
      <div class="card">
        <p class="card-title">Notification channel</p>
        <div class="card-value">Unknown</div>
        <p class="subtle">No outbound integrations connected.</p>
      </div>
      <div class="card">
        <p class="card-title">Sync status</p>
        <div class="card-value">Offline</div>
        <p class="subtle">Pulse cannot sync without credentials.</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Settings' })
</script>
