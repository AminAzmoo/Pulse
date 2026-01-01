<template>
  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Device inventory</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="empty-state">
      <p>Device telemetry is unavailable. All metrics display as N/A until Pulse connects.</p>
    </div>
  </section>

  <section class="section">
    <table class="table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Role</th>
          <th>IP</th>
          <th>Location</th>
          <th>Status</th>
          <th>CPU</th>
          <th>RAM</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>N/A</td>
          <td>Unknown</td>
          <td>N/A</td>
          <td>Unknown</td>
          <td>Offline</td>
          <td>N/A</td>
          <td>N/A</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Devices' })
</script>
