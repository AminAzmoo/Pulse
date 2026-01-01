<script setup lang="ts">
useHead({ title: 'Pulse | Settings' });
</script>

<template>
  <section class="page-header">
    <h1 class="page-title">Settings</h1>
    <p class="page-subtitle">
      Configure Pulse environments and connectivity preferences.
    </p>
  </section>

  <section class="section">
    <h2 class="section-title">Connectivity</h2>
    <div class="card">
      <span class="card-title">API Base URL</span>
      <span class="card-value">N/A</span>
      <span class="status-pill">Set via runtime config</span>
    </div>
  </section>

  <section class="section">
    <h2 class="section-title">Notifications</h2>
    <p class="empty-state">
      Notification preferences will appear once user profiles are enabled.
    </p>
  </section>
</template>
