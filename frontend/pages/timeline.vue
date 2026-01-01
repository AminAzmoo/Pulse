<template>
  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Activity timeline</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="empty-state">
      <p>No activity feed is connected. Events will appear here when Pulse is online.</p>
      <p class="subtle">Current status: Unknown</p>
    </div>
  </section>

  <section class="section">
    <div class="card">
      <p class="card-title">Latest event</p>
      <div class="card-value">N/A</div>
      <p class="subtle">Timeline data is offline.</p>
    </div>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Timeline' })
</script>
