<template>
  <section class="section">
    <div class="card-grid">
      <div v-for="card in summaryCards" :key="card.label" class="card">
        <p class="card-title">{{ card.label }}</p>
        <div class="card-value">{{ card.value }}</div>
        <p class="subtle">{{ card.note }}</p>
      </div>
    </div>
  </section>

  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Network health</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="card">
      <p class="subtle">
        Live telemetry is unavailable. Connect Pulse to a data source to replace N/A values with
        real-time metrics.
      </p>
      <div class="card-grid" style="margin-top: 16px;">
        <div class="card">
          <p class="card-title">Average latency</p>
          <div class="card-value">N/A</div>
        </div>
        <div class="card">
          <p class="card-title">Active alerts</p>
          <div class="card-value">Unknown</div>
        </div>
        <div class="card">
          <p class="card-title">Packet loss</p>
          <div class="card-value">N/A</div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Dashboard' })

const summaryCards = [
  { label: 'Devices online', value: 'N/A', note: 'Awaiting device status feed' },
  { label: 'Active tunnels', value: 'N/A', note: 'No tunnel metrics available' },
  { label: 'Services live', value: 'N/A', note: 'Service health unknown' },
  { label: 'Incidents', value: 'Unknown', note: 'Alert feed offline' }
]
</script>
