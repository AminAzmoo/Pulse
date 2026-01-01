<template>
  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Tunnel status</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="empty-state">
      <p>Tunnel orchestration data is offline. Pulse is showing Unknown status values.</p>
    </div>
  </section>

  <section class="section">
    <table class="table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Path</th>
          <th>Type</th>
          <th>Status</th>
          <th>Latency</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>N/A</td>
          <td>N/A</td>
          <td>Unknown</td>
          <td>Offline</td>
          <td>N/A</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Tunnels' })
</script>
